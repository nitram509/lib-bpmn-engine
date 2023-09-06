package bpmn_engine

import (
	"fmt"
	"sort"
	"time"

	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/var_holder"

	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/exporter"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
)

type BpmnEngine interface {
	LoadFromFile(filename string) (*ProcessInfo, error)
	LoadFromBytes(xmlData []byte) (*ProcessInfo, error)
	NewTaskHandler() NewTaskHandlerCommand1
	CreateInstance(processKey int64, variableContext map[string]interface{}) (*processInstanceInfo, error)
	CreateInstanceById(processId string, variableContext map[string]interface{}) (*processInstanceInfo, error)
	CreateAndRunInstance(processKey int64, variableContext map[string]interface{}) (*processInstanceInfo, error)
	CreateAndRunInstanceById(processId string, variableContext map[string]interface{}) (*processInstanceInfo, error)
	RunOrContinueInstance(processInstanceKey int64) (*processInstanceInfo, error)
	Name() string
	ProcessInstances() []*processInstanceInfo
	FindProcessInstance(processInstanceKey int64) *processInstanceInfo
	FindProcessesById(id string) []*ProcessInfo
}

// New creates a new instance of the BPMN Engine;
func New() BpmnEngineState {
	return NewWithName(fmt.Sprintf("Bpmn-Engine-%d", getGlobalSnowflakeIdGenerator().Generate().Int64()))
}

// NewWithName creates an engine with an arbitrary name of the engine;
// useful in case you have multiple ones, in order to distinguish them;
// also stored in when marshalling a process instance state, in case you want to store some special identifier
func NewWithName(name string) BpmnEngineState {
	snowflakeIdGenerator := getGlobalSnowflakeIdGenerator()
	return BpmnEngineState{
		name:                 name,
		processes:            []*ProcessInfo{},
		processInstances:     []*processInstanceInfo{},
		taskHandlers:         []*taskHandler{},
		jobs:                 []*job{},
		messageSubscriptions: []*MessageSubscription{},
		snowflake:            snowflakeIdGenerator,
		exporters:            []exporter.EventExporter{},
	}
}

// CreateInstanceById creates a new instance for a process with given process ID and uses latest version (if available)
// Might return BpmnEngineError, when no process with given ID was found
func (state *BpmnEngineState) CreateInstanceById(processId string, variableContext map[string]interface{}) (*processInstanceInfo, error) {
	var processes []*ProcessInfo
	for _, process := range state.processes {
		if process.BpmnProcessId == processId {
			processes = append(processes, process)
		}
	}
	if len(processes) > 0 {
		sort.SliceStable(processes, func(i, j int) bool {
			return processes[i].Version > processes[j].Version
		})
		return state.CreateInstance(processes[0].ProcessKey, variableContext)
	}
	return nil, newEngineErrorf("no process with id=%s was found (prior loaded into the engine)", processId)
}

// CreateInstance creates a new instance for a process with given processKey
// Might return BpmnEngineError, if process key was not found
func (state *BpmnEngineState) CreateInstance(processKey int64, variableContext map[string]interface{}) (*processInstanceInfo, error) {
	for _, process := range state.processes {
		if process.ProcessKey == processKey {
			processInstanceInfo := processInstanceInfo{
				ProcessInfo:    process,
				InstanceKey:    state.generateKey(),
				VariableHolder: var_holder.New(nil, variableContext),
				CreatedAt:      time.Now(),
				State:          Ready,
			}
			state.processInstances = append(state.processInstances, &processInstanceInfo)
			state.exportProcessInstanceEvent(*process, processInstanceInfo)
			return &processInstanceInfo, nil
		}
	}
	return nil, newEngineErrorf("no process with key=%d was found (prior loaded into the engine)", processKey)
}

// CreateAndRunInstanceById creates a new instance by process ID (and uses latest process version), and executes it immediately.
// The provided variableContext can be nil or refers to a variable map,
// which is provided to every service task handler function.
// Might return BpmnEngineError or ExpressionEvaluationError.
func (state *BpmnEngineState) CreateAndRunInstanceById(processId string, variableContext map[string]interface{}) (*processInstanceInfo, error) {
	instance, err := state.CreateInstanceById(processId, variableContext)
	if err != nil {
		return nil, err
	}
	err = state.run(instance)
	return instance, state.run(instance)
}

// CreateAndRunInstance creates a new instance and executes it immediately.
// The provided variableContext can be nil or refers to a variable map,
// which is provided to every service task handler function.
// Might return BpmnEngineError or ExpressionEvaluationError.
func (state *BpmnEngineState) CreateAndRunInstance(processKey int64, variableContext map[string]interface{}) (*processInstanceInfo, error) {
	instance, err := state.CreateInstance(processKey, variableContext)
	if err != nil {
		return nil, err
	}
	return instance, state.run(instance)
}

// RunOrContinueInstance runs or continues a process instance by a given processInstanceKey.
// returns the process instances, when found;
// does nothing, if process is already in ProcessInstanceCompleted State;
// returns nil, nil when no process instance was found;
// might return BpmnEngineError or ExpressionEvaluationError.
func (state *BpmnEngineState) RunOrContinueInstance(processInstanceKey int64) (*processInstanceInfo, error) {
	for _, pi := range state.processInstances {
		if processInstanceKey == pi.InstanceKey {
			return pi, state.run(pi)
		}
	}
	return nil, nil
}

func (state *BpmnEngineState) run(instance *processInstanceInfo) (err error) {
	process := instance.ProcessInfo
	var commandQueue []command

	switch instance.State {
	case Ready:
		// use start events to start the instance
		for _, startEvent := range process.definitions.Process.StartEvents {
			var be BPMN20.BaseElement = startEvent
			commandQueue = append(commandQueue, activityCommand{
				element: &be,
			})
		}
		instance.State = Active
		// TODO: check? export process EVENT
	case Active:
		jobs := state.findActiveJobsForContinuation(instance)
		for _, j := range jobs {
			commandQueue = append(commandQueue, continueActivityCommand{
				activity: j,
			})
		}
		activeSubscriptions := state.findActiveSubscriptions(instance)
		for _, subscr := range activeSubscriptions {
			commandQueue = append(commandQueue, continueActivityCommand{
				activity:       subscr,
				originActivity: subscr.originActivity,
			})
		}
		createdTimers := state.findCreatedTimers(instance)
		for _, timer := range createdTimers {
			commandQueue = append(commandQueue, continueActivityCommand{
				activity:       timer,
				originActivity: timer.originActivity,
			})
		}
	}

	// *** MAIN LOOP ***
	for len(commandQueue) > 0 {
		cmd := commandQueue[0]
		commandQueue = commandQueue[1:]

		switch cmd.Type() {
		case flowTransitionType:
			sourceActivity := cmd.(flowTransitionCommand).sourceActivity
			flowId := cmd.(flowTransitionCommand).sequenceFlowId
			nextFlows := BPMN20.FindSequenceFlows(&process.definitions.Process.SequenceFlows, []string{flowId})
			if BPMN20.ExclusiveGateway == (*sourceActivity.Element()).GetType() {
				nextFlows, err = exclusivelyFilterByConditionExpression(nextFlows, instance.VariableHolder.Variables())
				if err != nil {
					instance.State = Failed
					return err
				}
			}
			for _, flow := range nextFlows {
				state.exportSequenceFlowEvent(*process, *instance, flow)
				baseElements := BPMN20.FindBaseElementsById(&process.definitions, flow.TargetRef)
				targetBaseElement := baseElements[0]
				aCmd := activityCommand{
					sourceId:       flowId,
					originActivity: sourceActivity,
					element:        targetBaseElement,
				}
				commandQueue = append(commandQueue, aCmd)
			}
		case activityType:
			element := cmd.(activityCommand).element
			originActivity := cmd.(activityCommand).originActivity
			nextCommands := state.handleElement(process, instance, element, originActivity)
			state.exportElementEvent(*process, *instance, *element, exporter.ElementCompleted)
			commandQueue = append(commandQueue, nextCommands...)
		case continueActivityType:
			element := cmd.(continueActivityCommand).activity.Element()
			originActivity := cmd.(continueActivityCommand).originActivity
			nextCommands := state.handleElement(process, instance, element, originActivity)
			commandQueue = append(commandQueue, nextCommands...)
		case errorType:
			err = cmd.(errorCommand).err
			instance.State = Failed
			break
		case checkExclusiveGatewayDoneType:
			activity := cmd.(checkExclusiveGatewayDoneCommand).gatewayActivity
			state.checkExclusiveGatewayDone(activity)
		default:
			panic("invariants for command type check not fully implemented")
		}
	}

	if instance.State == Completed || instance.State == Failed {
		// TODO need to send failed State
		state.exportEndProcessEvent(*process, *instance)
	}

	return err
}

func (state *BpmnEngineState) handleElement(process *ProcessInfo, instance *processInstanceInfo, element *BPMN20.BaseElement, originActivity activity) []command {
	state.exportElementEvent(*process, *instance, *element, exporter.ElementActivated) // FIXME: don't create event on continuation ?!?!
	createFlowTransitions := true
	var activity activity
	var nextCommands []command
	var err error
	switch (*element).GetType() {
	case BPMN20.StartEvent:
		createFlowTransitions = true
		activity = &elementActivity{
			key:     state.generateKey(),
			state:   Completed,
			element: element,
		}
	case BPMN20.EndEvent:
		state.handleEndEvent(process, instance)
		state.exportElementEvent(*process, *instance, *element, exporter.ElementCompleted) // special case here, to end the instance
		createFlowTransitions = false
		activity = &elementActivity{
			key:     state.generateKey(),
			state:   Completed,
			element: element,
		}
	case BPMN20.ServiceTask:
		taskElement := (*element).(BPMN20.TaskElement)
		_, activity = state.handleServiceTask(process, instance, &taskElement)
		createFlowTransitions = activity.State() == Completed
	case BPMN20.UserTask:
		taskElement := (*element).(BPMN20.TaskElement)
		activity = state.handleUserTask(process, instance, &taskElement)
		createFlowTransitions = activity.State() == Completed
	case BPMN20.IntermediateCatchEvent:
		ice := (*element).(BPMN20.TIntermediateCatchEvent)
		createFlowTransitions, activity, err = state.handleIntermediateCatchEvent(process, instance, ice, originActivity)
		if err != nil {
			nextCommands = append(nextCommands, errorCommand{
				err:         err,
				elementId:   (*element).GetId(),
				elementName: (*element).GetName(),
			})
		} else {
			nextCommands = append(nextCommands, createCheckExclusiveGatewayDoneCommand(originActivity)...)
		}
	case BPMN20.IntermediateThrowEvent:
		activity = &elementActivity{
			key:     state.generateKey(),
			state:   Active, // FIXME: should be Completed?
			element: element,
		}
		cmds := state.handleIntermediateThrowEvent(process, instance, (*element).(BPMN20.TIntermediateThrowEvent), activity)
		nextCommands = append(nextCommands, cmds...)
		createFlowTransitions = false
	case BPMN20.ParallelGateway:
		createFlowTransitions, activity = state.handleParallelGateway(process, instance, (*element).(BPMN20.TParallelGateway), originActivity)
	case BPMN20.ExclusiveGateway:
		activity = elementActivity{
			key:     state.generateKey(),
			state:   Active,
			element: element,
		}
		createFlowTransitions = true
	case BPMN20.EventBasedGateway:
		activity = &eventBasedGatewayActivity{
			key:     state.generateKey(),
			state:   Completed,
			element: element,
		}
		instance.appendActivity(activity)
		createFlowTransitions = true
	default:
		panic(fmt.Sprintf("unsupported element: id=%s, type=%s", (*element).GetId(), (*element).GetType()))
	}
	if createFlowTransitions && err == nil {
		nextCommands = append(nextCommands, createNextCommands(process, instance, element, activity)...)
	}
	return nextCommands
}

func createCheckExclusiveGatewayDoneCommand(originActivity activity) (cmds []command) {
	if (*originActivity.Element()).GetType() == BPMN20.EventBasedGateway {
		evtBasedGatewayActivity := originActivity.(*eventBasedGatewayActivity)
		cmds = append(cmds, checkExclusiveGatewayDoneCommand{
			gatewayActivity: *evtBasedGatewayActivity,
		})
	}
	return cmds
}

func createNextCommands(process *ProcessInfo, instance *processInstanceInfo, element *BPMN20.BaseElement, activity activity) (cmds []command) {
	nextFlows := BPMN20.FindSequenceFlows(&process.definitions.Process.SequenceFlows, (*element).GetOutgoingAssociation())
	var err error
	if (*element).GetType() == BPMN20.ExclusiveGateway {
		nextFlows, err = exclusivelyFilterByConditionExpression(nextFlows, instance.VariableHolder.Variables())
		if err != nil {
			instance.State = Failed
			cmds = append(cmds, errorCommand{
				err:         err,
				elementId:   (*element).GetId(),
				elementName: (*element).GetName(),
			})
			return cmds
		}
	}
	for _, flow := range nextFlows {
		cmds = append(cmds, flowTransitionCommand{
			sourceId:       (*element).GetId(),
			sourceActivity: activity,
			sequenceFlowId: flow.Id,
		})
	}
	return cmds
}

func (state *BpmnEngineState) handleIntermediateCatchEvent(process *ProcessInfo, instance *processInstanceInfo, ice BPMN20.TIntermediateCatchEvent, originActivity activity) (continueFlow bool, activity activity, err error) {
	continueFlow = false
	if ice.MessageEventDefinition.Id != "" {
		continueFlow, activity, err = state.handleIntermediateMessageCatchEvent(process, instance, ice, originActivity)
	} else if ice.TimerEventDefinition.Id != "" {
		continueFlow, activity, err = state.handleIntermediateTimerCatchEvent(instance, ice, originActivity)
	} else if ice.LinkEventDefinition.Id != "" {
		var be BPMN20.BaseElement = ice
		activity = &elementActivity{
			key:     state.generateKey(),
			state:   Active, // FIXME: should be Completed?
			element: &be,
		}
		throwLinkName := (*originActivity.Element()).(BPMN20.TIntermediateThrowEvent).LinkEventDefinition.Name
		catchLinkName := ice.LinkEventDefinition.Name
		elementVarHolder := var_holder.New(&instance.VariableHolder, nil)
		if err := propagateProcessInstanceVariables(&elementVarHolder, ice.Output); err != nil {
			msg := fmt.Sprintf("Can't evaluate expression in element id=%s name=%s", ice.Id, ice.Name)
			err = &ExpressionEvaluationError{Msg: msg, Err: err}
		} else {
			continueFlow = throwLinkName == catchLinkName // just stating the obvious
		}
	}
	return continueFlow, activity, err
}

func (state *BpmnEngineState) handleEndEvent(process *ProcessInfo, instance *processInstanceInfo) {
	activeMessageSubscriptions := false
	for _, ms := range state.messageSubscriptions {
		activeMessageSubscriptions = activeMessageSubscriptions || ms.State() == Active || ms.State() == Ready
		if activeMessageSubscriptions {
			break
		}
	}
	if !activeMessageSubscriptions {
		instance.State = Completed
	}
}

func (state *BpmnEngineState) handleParallelGateway(process *ProcessInfo, instance *processInstanceInfo, element BPMN20.TParallelGateway, originActivity activity) (continueFlow bool, resultActivity activity) {
	resultActivity = instance.findActiveActivityByElementId(element.Id)
	if resultActivity == nil {
		var be BPMN20.BaseElement = element
		resultActivity = &gatewayActivity{
			key:      state.generateKey(),
			state:    Active,
			element:  &be,
			parallel: true,
		}
		instance.appendActivity(resultActivity)
	}
	sourceFlow := BPMN20.FindSequenceFlow(&process.definitions.Process.SequenceFlows, (*originActivity.Element()).GetId(), element.GetId())
	resultActivity.(*gatewayActivity).SetInboundFlowCompleted(sourceFlow.Id)
	continueFlow = resultActivity.(*gatewayActivity).parallel && resultActivity.(*gatewayActivity).AreInboundFlowsCompleted()
	if continueFlow {
		resultActivity.(*gatewayActivity).SetState(Completed)
	}
	return continueFlow, resultActivity
}

func (state *BpmnEngineState) findActiveJobsForContinuation(instance *processInstanceInfo) (ret []*job) {
	for _, job := range state.jobs {
		if job.ProcessInstanceKey == instance.InstanceKey && job.JobState == Active {
			ret = append(ret, job)
		}
	}
	return ret
}

// findActiveSubscriptions returns active subscriptions;
// if ids are provided, the result gets filtered;
// if no ids are provided, all active subscriptions are returned
func (state *BpmnEngineState) findActiveSubscriptions(instance *processInstanceInfo) (result []*MessageSubscription) {
	for _, ms := range state.messageSubscriptions {
		if ms.ProcessInstanceKey == instance.InstanceKey && ms.MessageState == Active {
			result = append(result, ms)
		}
	}
	return result
}

// findCreatedTimers the list of all scheduled/creates timers in the engine, not yet completed
func (state *BpmnEngineState) findCreatedTimers(instance *processInstanceInfo) (result []*Timer) {
	for _, t := range state.timers {
		if instance.InstanceKey == t.ProcessInstanceKey && t.TimerState == TimerCreated {
			result = append(result, t)
		}
	}
	return result
}
