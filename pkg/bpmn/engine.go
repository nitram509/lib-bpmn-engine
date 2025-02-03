package bpmn

import (
	"fmt"
	"log"
	"time"

	rqliteExporter "github.com/pbinitiative/zenbpm/pkg/bpmn/exporter/rqlite"
	rqlite "github.com/pbinitiative/zenbpm/pkg/bpmn/persistence/rqlite"
	"github.com/pbinitiative/zenbpm/pkg/bpmn/var_holder"

	"github.com/pbinitiative/zenbpm/pkg/bpmn/exporter"
	"github.com/pbinitiative/zenbpm/pkg/bpmn/model/bpmn20"
)

type BpmnEngine interface {
	LoadFromFile(filename string) (*ProcessInfo, error)
	LoadFromBytes(xmlData []byte) (*ProcessInfo, error)
	NewTaskHandler() NewTaskHandlerCommand1
	CreateInstance(process *ProcessInfo, variableContext map[string]interface{}) (*processInstanceInfo, error)
	CreateInstanceById(processId string, variableContext map[string]interface{}) (*processInstanceInfo, error)
	CreateAndRunInstance(processKey int64, variableContext map[string]interface{}) (*processInstanceInfo, error)
	CreateAndRunInstanceById(processId string, variableContext map[string]interface{}) (*processInstanceInfo, error)
	RunOrContinueInstance(processInstanceKey int64) (*processInstanceInfo, error)
	Name() string
	FindProcessInstance(processInstanceKey int64) *processInstanceInfo
	FindProcessesById(id string) []*ProcessInfo
	JobCompleteById(jobId int64)
	GetPersistence() *rqlite.BpmnEnginePersistenceRqlite
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
	state := BpmnEngineState{
		name:         name,
		taskHandlers: []*taskHandler{},
		snowflake:    snowflakeIdGenerator,
		exporters:    []exporter.EventExporter{},
		persistence:  nil,
	}

	rqliteService := rqlite.NewBpmnEnginePersistenceRqlite(snowflakeIdGenerator)
	// defer rqliteService.RqliteStop()

	var p BpmnEnginePersistenceService = NewBpmnEnginePersistenceRqlite(snowflakeIdGenerator, &state, rqliteService)

	state.persistence = p // create the client
	exporter, _ := rqliteExporter.NewExporter(rqliteService)
	// register the exporter
	state.AddEventExporter(&exporter)
	return state
}

func (state *BpmnEngineState) GetPersistence() *rqlite.BpmnEnginePersistenceRqlite {
	return state.persistence.GetPersistence()
}

// CreateInstanceById creates a new instance for a process with given process ID and uses latest version (if available)
// Might return BpmnEngineError, when no process with given ID was found
func (state *BpmnEngineState) CreateInstanceById(processId string, variableContext map[string]interface{}) (*processInstanceInfo, error) {
	process := state.persistence.FindProcessById(processId)

	if process != nil {
		return state.CreateInstance(process, variableContext)
	}

	return nil, newEngineErrorf("no process with id=%s was found (prior loaded into the engine)", processId)
}

// CreateInstance creates a new instance for a process with given processKey
// Might return BpmnEngineError, if process key was not found
func (state *BpmnEngineState) CreateInstance(process *ProcessInfo, variableContext map[string]interface{}) (*processInstanceInfo, error) {
	processInstanceInfo := processInstanceInfo{
		ProcessInfo:    process,
		InstanceKey:    state.generateKey(),
		VariableHolder: var_holder.New(nil, variableContext),
		CreatedAt:      time.Now(),
		State:          Ready,
		CaughtEvents:   []catchEvent{},
		activities:     []activity{},
	}
	err := state.persistence.PersistProcessInstance(&processInstanceInfo)
	if err != nil {
		return nil, err
	}
	state.exportProcessInstanceEvent(*process, processInstanceInfo)
	return &processInstanceInfo, nil
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
	return instance, state.run(instance)
}

// CreateAndRunInstance creates a new instance and executes it immediately.
// The provided variableContext can be nil or refers to a variable map,
// which is provided to every service task handler function.
// Might return BpmnEngineError or ExpressionEvaluationError.
func (state *BpmnEngineState) CreateAndRunInstance(processKey int64, variableContext map[string]interface{}) (*processInstanceInfo, error) {
	process := state.persistence.FindProcessByKey(processKey)

	instance, err := state.CreateInstance(process, variableContext)
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
	pi := state.persistence.FindProcessInstanceByKey(processInstanceKey)
	if pi == nil {
		return nil, nil
	}
	return pi, state.run(pi)
}

func (state *BpmnEngineState) run(instance *processInstanceInfo) (err error) {
	process := instance.ProcessInfo
	var commandQueue []command

	switch instance.State {
	case Ready:
		// use start events to start the instance
		for _, startEvent := range process.definitions.Process.StartEvents {
			var be bpmn20.BaseElement = startEvent
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
			nextFlows := bpmn20.FindSequenceFlows(&process.definitions.Process.SequenceFlows, []string{flowId})
			if bpmn20.ExclusiveGateway == (*sourceActivity.Element()).GetType() {
				nextFlows, err = exclusivelyFilterByConditionExpression(nextFlows, instance.VariableHolder.Variables())
				if err != nil {
					instance.State = Failed
					return err
				}
			}
			for _, flow := range nextFlows {
				state.exportSequenceFlowEvent(*process, *instance, flow)
				baseElements := bpmn20.FindBaseElementsById(&process.definitions, flow.TargetRef)
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
			panic("[invariant check] command type check not fully implemented")
		}
	}

	if instance.State == Completed || instance.State == Failed {
		// TODO need to send failed State
		state.exportEndProcessEvent(*process, *instance)
	}
	// TODO: persistently update state
	state.persistence.PersistProcessInstance(instance)

	return err
}

func (state *BpmnEngineState) handleElement(process *ProcessInfo, instance *processInstanceInfo, element *bpmn20.BaseElement, originActivity activity) []command {
	state.exportElementEvent(*process, *instance, *element, exporter.ElementActivated) // FIXME: don't create event on continuation ?!?!
	createFlowTransitions := true
	var activity activity
	var nextCommands []command
	var err error
	switch (*element).GetType() {
	case bpmn20.StartEvent:
		createFlowTransitions = true
		activity = &elementActivity{
			key:     state.generateKey(),
			state:   Completed,
			element: element,
		}
	case bpmn20.EndEvent:
		state.handleEndEvent(process, instance)
		state.exportElementEvent(*process, *instance, *element, exporter.ElementCompleted) // special case here, to end the instance
		createFlowTransitions = false
		activity = &elementActivity{
			key:     state.generateKey(),
			state:   Completed,
			element: element,
		}
	case bpmn20.ServiceTask:
		taskElement := (*element).(bpmn20.TaskElement)
		_, activity = state.handleServiceTask(process, instance, &taskElement)
		createFlowTransitions = activity.State() == Completed
	case bpmn20.UserTask:
		taskElement := (*element).(bpmn20.TaskElement)
		activity = state.handleUserTask(process, instance, &taskElement)
		createFlowTransitions = activity.State() == Completed
	case bpmn20.IntermediateCatchEvent:
		ice := (*element).(bpmn20.TIntermediateCatchEvent)
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

		ms, err := activity.(*MessageSubscription)
		if !err {
			// Handle the case when activity is not a MessageSubscription
			// For example, you can return an error or log a message
			log.Panicf("Unexpected activity type: %T", activity)
		}
		state.persistence.PersistNewMessageSubscription(ms)
	case bpmn20.IntermediateThrowEvent:
		activity = &elementActivity{
			key:     state.generateKey(),
			state:   Active, // FIXME: should be Completed?
			element: element,
		}
		cmds := state.handleIntermediateThrowEvent(process, instance, (*element).(bpmn20.TIntermediateThrowEvent), activity)
		nextCommands = append(nextCommands, cmds...)
		createFlowTransitions = false
	case bpmn20.ParallelGateway:
		createFlowTransitions, activity = state.handleParallelGateway(process, instance, (*element).(bpmn20.TParallelGateway), originActivity)
	case bpmn20.ExclusiveGateway:
		activity = elementActivity{
			key:     state.generateKey(),
			state:   Active,
			element: element,
		}
		createFlowTransitions = true
	case bpmn20.EventBasedGateway:
		activity = &eventBasedGatewayActivity{
			key:     state.generateKey(),
			state:   Completed,
			element: element,
		}
		instance.appendActivity(activity)
		createFlowTransitions = true
	case bpmn20.InclusiveGateway:
		activity = elementActivity{
			key:     state.generateKey(),
			state:   Active,
			element: element,
		}
		createFlowTransitions = true
	default:
		panic(fmt.Sprintf("[invariant check] unsupported element: id=%s, type=%s", (*element).GetId(), (*element).GetType()))
	}
	if createFlowTransitions && err == nil {
		state.exportElementEvent(*process, *instance, *element, exporter.ElementCompleted)
		nextCommands = append(nextCommands, createNextCommands(process, instance, element, activity)...)
	}
	return nextCommands
}

func createCheckExclusiveGatewayDoneCommand(originActivity activity) (cmds []command) {
	if (*originActivity.Element()).GetType() == bpmn20.EventBasedGateway {
		evtBasedGatewayActivity := originActivity.(*eventBasedGatewayActivity)
		cmds = append(cmds, checkExclusiveGatewayDoneCommand{
			gatewayActivity: *evtBasedGatewayActivity,
		})
	}
	return cmds
}

func createNextCommands(process *ProcessInfo, instance *processInstanceInfo, element *bpmn20.BaseElement, activity activity) (cmds []command) {
	nextFlows := bpmn20.FindSequenceFlows(&process.definitions.Process.SequenceFlows, (*element).GetOutgoingAssociation())
	var err error
	switch (*element).GetType() {
	case bpmn20.ExclusiveGateway:
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
	case bpmn20.InclusiveGateway:
		nextFlows, err = inclusivelyFilterByConditionExpression(nextFlows, instance.VariableHolder.Variables())
		if err != nil {
			instance.State = Failed
			return []command{
				errorCommand{
					elementId:   (*element).GetId(),
					elementName: (*element).GetName(),
					err:         err,
				},
			}
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

func (state *BpmnEngineState) handleIntermediateCatchEvent(process *ProcessInfo, instance *processInstanceInfo, ice bpmn20.TIntermediateCatchEvent, originActivity activity) (continueFlow bool, activity activity, err error) {
	continueFlow = false
	if ice.MessageEventDefinition.Id != "" {
		continueFlow, activity, err = state.handleIntermediateMessageCatchEvent(process, instance, ice, originActivity)
	} else if ice.TimerEventDefinition.Id != "" {
		continueFlow, activity, err = state.handleIntermediateTimerCatchEvent(instance, ice, originActivity)
	} else if ice.LinkEventDefinition.Id != "" {
		var be bpmn20.BaseElement = ice
		activity = &elementActivity{
			key:     state.generateKey(),
			state:   Active, // FIXME: should be Completed?
			element: &be,
		}
		throwLinkName := (*originActivity.Element()).(bpmn20.TIntermediateThrowEvent).LinkEventDefinition.Name
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
	// FIXME: check if this is correct to seems wrong i need to check if there are any tokens in this process not only messages subscriptions but elements also
	if len(state.persistence.FindMessageSubscription(-1, instance, "", Active)) > 0 {
		activeMessageSubscriptions = true
	}
	if len(state.persistence.FindMessageSubscription(-1, instance, "", Ready)) > 0 {
		activeMessageSubscriptions = true
	}

	if !activeMessageSubscriptions {
		instance.State = Completed
	}
}

func (state *BpmnEngineState) handleParallelGateway(process *ProcessInfo, instance *processInstanceInfo, element bpmn20.TParallelGateway, originActivity activity) (continueFlow bool, resultActivity activity) {
	resultActivity = instance.findActiveActivityByElementId(element.Id)
	if resultActivity == nil {
		var be bpmn20.BaseElement = element
		resultActivity = &gatewayActivity{
			key:      state.generateKey(),
			state:    Active,
			element:  &be,
			parallel: true,
		}
		instance.appendActivity(resultActivity)
	}
	sourceFlow := bpmn20.FindFirstSequenceFlow(&process.definitions.Process.SequenceFlows, (*originActivity.Element()).GetId(), element.GetId())
	resultActivity.(*gatewayActivity).SetInboundFlowCompleted(sourceFlow.Id)
	continueFlow = resultActivity.(*gatewayActivity).parallel && resultActivity.(*gatewayActivity).AreInboundFlowsCompleted()
	if continueFlow {
		resultActivity.(*gatewayActivity).SetState(Completed)
	}
	return continueFlow, resultActivity
}

func (state *BpmnEngineState) findActiveJobsForContinuation(instance *processInstanceInfo) (ret []*job) {
	return state.persistence.FindJobs("", instance, -1, Active, Completing)
}

// findActiveSubscriptions returns active subscriptions;
// if ids are provided, the result gets filtered;
// if no ids are provided, all active subscriptions are returned
func (state *BpmnEngineState) findActiveSubscriptions(instance *processInstanceInfo) (result []*MessageSubscription) {
	for _, ms := range state.persistence.FindMessageSubscription(-1, instance, "", Active) {
		bes := bpmn20.FindBaseElementsById(&instance.ProcessInfo.definitions, ms.ElementId)
		if len(bes) == 0 {
			continue
		}
		ms.baseElement = bes[0]
		// FIXME: rewrite this hack
		// instance.findActivity(ms.originActivity.Key())
		result = append(result, ms)
	}
	return result
}

// findCreatedTimers the list of all scheduled/creates timers in the engine, not yet completed
func (state *BpmnEngineState) findCreatedTimers(instance *processInstanceInfo) (result []*Timer) {
	for _, t := range state.persistence.FindTimers(-1, instance.InstanceKey, TimerCreated) {
		bes := bpmn20.FindBaseElementsById(&instance.ProcessInfo.definitions, t.ElementId)
		if len(bes) == 0 {
			continue
		}
		t.baseElement = bes[0]
		result = append(result, t)
	}
	return result
}
