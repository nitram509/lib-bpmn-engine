package bpmn_engine

import (
	"errors"
	"fmt"
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
	CreateAndRunInstance(processKey int64, variableContext map[string]interface{}) (*processInstanceInfo, error)
	RunOrContinueInstance(processInstanceKey int64) (*processInstanceInfo, error)
	GetName() string
	GetProcessInstances() []*processInstanceInfo
	FindProcessInstanceById(processInstanceKey int64) *processInstanceInfo
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

// CreateInstance creates a new instance for a process with given processKey
// will return (nil, nil), when no process with given was found
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
	return nil, nil
}

// CreateAndRunInstance creates a new instance and executes it immediately.
// The provided variableContext can be nil or refers to a variable map,
// which is provided to every service task handler function.
func (state *BpmnEngineState) CreateAndRunInstance(processKey int64, variableContext map[string]interface{}) (*processInstanceInfo, error) {
	instance, err := state.CreateInstance(processKey, variableContext)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, errors.New(fmt.Sprint("can't find process with processKey=", processKey, "."))
	}

	err = state.run(instance)
	return instance, err
}

// RunOrContinueInstance runs or continues a process instance by a given processInstanceKey.
// returns the process instances, when found
// does nothing, if process is already in ProcessInstanceCompleted State
// returns nil, when no process instance was found
// Additionally, every time this method is called, former completed instances are 'garbage collected'.
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

	switch instance.State {
	case Ready:
		// use start events to start the instance
		for _, startEvent := range process.definitions.Process.StartEvents {
			var be BPMN20.BaseElement = startEvent
			instance.appendCommand(tActivityCommand{
				element: &be,
			})
		}
		instance.State = Active
		// TODO: check? export process EVENT
	case Active:
		jobs := state.findActiveJobsForContinuation(instance)
		for _, j := range jobs {
			element := BPMN20.FindBaseElementsById(process.definitions, j.ElementId)[0]
			activity := tActivity{
				key:     j.JobKey,
				state:   j.JobState,
				element: &element,
			}
			instance.appendCommand(tContinueActivityCommand{
				activity: &activity,
			})
		}
		activeSubscriptions := state.findActiveSubscriptions(instance)
		for _, subscr := range activeSubscriptions {
			element := BPMN20.FindBaseElementsById(process.definitions, subscr.ElementId)[0]
			activity := tActivity{
				key:     subscr.ElementInstanceKey,
				state:   subscr.State,
				element: &element,
			}
			instance.appendCommand(tContinueActivityCommand{
				activity:       &activity,
				originActivity: subscr.originActivity,
			})
		}
		createdTimers := state.findCreatedTimers(instance)
		for _, timer := range createdTimers {
			element := BPMN20.FindBaseElementsById(process.definitions, timer.ElementId)[0]
			activity := tActivity{
				key:     timer.ElementInstanceKey,
				state:   Active,
				element: &element,
			}
			instance.appendCommand(tContinueActivityCommand{
				activity:       &activity,
				originActivity: timer.originActivity,
			})
		}
	}

	// *** MAIN LOOP ***
	for instance.hasCommands() {
		cmd := instance.popCommand()

		switch cmd.Type() {
		case flowTransitionType:
			originActivity := cmd.(flowTransitionCommand).OriginActivity()
			flowId := cmd.(flowTransitionCommand).SequenceFlowId()
			nextFlows := BPMN20.FindSequenceFlows(&process.definitions.Process.SequenceFlows, []string{flowId})
			if BPMN20.ExclusiveGateway == (*originActivity.Element()).GetType() {
				nextFlows, err = exclusivelyFilterByConditionExpression(nextFlows, instance.VariableHolder.Variables())
				if err != nil {
					instance.State = Failing
					return err
				}
			}
			for _, flow := range nextFlows {
				state.exportSequenceFlowEvent(*process, *instance, flow)
				baseElements := BPMN20.FindBaseElementsById(process.definitions, flow.TargetRef)
				targetBaseElement := baseElements[0]
				aCmd := tActivityCommand{
					sourceId:       flowId,
					originActivity: originActivity,
					element:        &targetBaseElement,
				}
				instance.appendCommand(aCmd)
			}
		case activityType:
			element := cmd.(activityCommand).Element()
			inboundFlowId := cmd.(activityCommand).InboundFlowId()
			originActivity := cmd.(activityCommand).OriginActivity()
			nextCommands := state.startActivity(process, instance, element, inboundFlowId, originActivity)
			state.exportElementEvent(*process, *instance, *element, exporter.ElementCompleted)
			for _, c := range nextCommands {
				instance.appendCommand(c)
			}
		case continueActivityType:
			element := cmd.(continueActivityCommand).Element()
			activity := cmd.(continueActivityCommand).Activity()
			nextCommands := state.continueActivity(process, instance, element, activity)
			for _, c := range nextCommands {
				instance.appendCommand(c)
			}
		case errorType:
			err = cmd.(ErrorCommand).Error()
			instance.State = Failed
			break
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
		if ms.ProcessInstanceKey == instance.InstanceKey && ms.State == Active {
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

func (state *BpmnEngineState) startActivity(process *ProcessInfo, instance *processInstanceInfo, element *BPMN20.BaseElement, inboundFlowId string, originActivity Activity) []command {
	state.exportElementEvent(*process, *instance, *element, exporter.ElementActivated)
	createFlowTransitions := true
	var activity Activity
	var nextCommands []command
	switch (*element).GetType() {
	case BPMN20.StartEvent:
		createFlowTransitions = true
		activity = &tActivity{
			key:     state.generateKey(),
			state:   Completed,
			element: element,
		}
	case BPMN20.EndEvent:
		state.handleEndEvent(process, instance)
		state.exportElementEvent(*process, *instance, *element, exporter.ElementCompleted) // special case here, to end the instance
		createFlowTransitions = false
		activity = &tActivity{
			key:     state.generateKey(),
			state:   Completed,
			element: element,
		}
	case BPMN20.ServiceTask:
		taskElement := (*element).(BPMN20.TaskElement)
		var j *job
		_, j = state.handleServiceTask(process, instance, &taskElement)
		activity = &tActivity{
			key:     j.JobKey,
			state:   j.JobState,
			element: element,
		}
		createFlowTransitions = activity.State() == Completed
	case BPMN20.UserTask:
		taskElement := (*element).(BPMN20.TaskElement)
		j := state.handleUserTask(process, instance, &taskElement)
		activity = &tActivity{
			key:     j.JobKey,
			state:   j.JobState,
			element: element,
		}
		createFlowTransitions = j.JobState == Completed
	case BPMN20.IntermediateCatchEvent:
		ice := (*element).(BPMN20.TIntermediateCatchEvent)
		if ice.MessageEventDefinition.Id != "" {
			var ms *MessageSubscription
			var err error
			createFlowTransitions, ms, err = state.handleIntermediateMessageCatchEvent(process, instance, ice)
			if err != nil {
				nextCommands = append(nextCommands, tErrorCommand{
					err:         err,
					elementId:   (*element).GetId(),
					elementName: (*element).GetName(),
				})
			}
			activity = &tActivity{
				key:     ms.ElementInstanceKey,
				state:   ms.State,
				element: element,
			}
			ms.originActivity = originActivity
		} else if ice.TimerEventDefinition.Id != "" {
			var timer *Timer
			var err error
			createFlowTransitions, timer, err = state.handleIntermediateTimerCatchEvent(process, instance, ice)
			if err != nil {
				nextCommands = append(nextCommands, tErrorCommand{
					err:         err,
					elementId:   (*element).GetId(),
					elementName: (*element).GetName(),
				})
			}
			if timer != nil {
				activity = &tActivity{
					key:     timer.ElementInstanceKey,
					state:   Active, // FIXME: transform from imer.TimerState,
					element: element,
				}
				timer.originActivity = originActivity
			}
		} else if ice.LinkEventDefinition.Id != "" {
			activity = &tActivity{
				key:     state.generateKey(),
				state:   Active,
				element: element,
			}
			throwLinkName := (*originActivity.Element()).(BPMN20.TIntermediateThrowEvent).LinkEventDefinition.Name
			catchLinkName := ice.LinkEventDefinition.Name
			elementVarHolder := var_holder.New(&instance.VariableHolder, nil)
			if err := propagateProcessInstanceVariables(&elementVarHolder, ice.Output); err != nil {
				msg := fmt.Sprintf("Can't evaluate expression in element id=%s name=%s", ice.Id, ice.Name)
				nextCommands = append(nextCommands, &tErrorCommand{
					err:         &ExpressionEvaluationError{Msg: msg, Err: err},
					elementId:   ice.Id,
					elementName: ice.Name,
				})
			} else {
				createFlowTransitions = throwLinkName == catchLinkName // just stating the obvious
			}
		}
	case BPMN20.IntermediateThrowEvent:
		activity = &tActivity{
			key:     state.generateKey(),
			state:   Active,
			element: element,
		}
		cmds := state.handleIntermediateThrowEvent(process, instance, (*element).(BPMN20.TIntermediateThrowEvent), activity, inboundFlowId)
		nextCommands = append(nextCommands, cmds...)
		createFlowTransitions = false
	case BPMN20.ParallelGateway:
		activity = &tGatewayActivity{
			key:      state.generateKey(),
			state:    Active,
			element:  element,
			parallel: true,
		}
		createFlowTransitions = state.handleParallelGateway(instance, (*element).(BPMN20.TParallelGateway), activity, inboundFlowId)
	case BPMN20.ExclusiveGateway:
		activity = &tActivity{
			key:     state.generateKey(),
			state:   Active,
			element: element,
		}
		createFlowTransitions = true
	case BPMN20.EventBasedGateway:
		activity = &tEventBasedGatewayActivity{
			key:     state.generateKey(),
			state:   Completed,
			element: element,
		}
		instance.appendActivity(activity)
		createFlowTransitions = true
	default:
		panic(fmt.Sprintf("unsupported element: id=%s, type=%s", (*element).GetId(), (*element).GetType()))
	}
	if createFlowTransitions {
		nextCommands = createNextCommands(process, instance, element, activity)
	}
	return nextCommands
}

func (state *BpmnEngineState) continueActivity(process *ProcessInfo, instance *processInstanceInfo, element *BPMN20.BaseElement, activity Activity) []command {
	createFlowTransitions := false
	var nextCommands []command
	switch (*element).GetType() {
	case BPMN20.ServiceTask:
		taskElement := (*element).(BPMN20.TaskElement)
		_, j := state.handleServiceTask(process, instance, &taskElement)
		activity = &tActivity{
			key:     j.JobKey,
			state:   j.JobState,
			element: element,
		}
		createFlowTransitions = j.JobState == Completed
	case BPMN20.UserTask:
		taskElement := (*element).(BPMN20.TaskElement)
		j := state.handleUserTask(process, instance, &taskElement)
		activity = &tActivity{
			key:     j.JobKey,
			state:   j.JobState,
			element: element,
		}
		createFlowTransitions = j.JobState == Completed
	case BPMN20.IntermediateCatchEvent:
		var err error
		createFlowTransitions, err = state.handleIntermediateCatchEvent(process, instance, (*element).(BPMN20.TIntermediateCatchEvent), activity)
		if err != nil {
			nextCommands = append(nextCommands, tErrorCommand{
				err:         err,
				elementId:   (*element).GetId(),
				elementName: (*element).GetName(),
			})
		}
	}
	if createFlowTransitions {
		nextCommands = createNextCommands(process, instance, element, activity)
	}
	return nextCommands
}

func createNextCommands(process *ProcessInfo, instance *processInstanceInfo, element *BPMN20.BaseElement, activity Activity) (cmds []command) {
	nextFlows := BPMN20.FindSequenceFlows(&process.definitions.Process.SequenceFlows, (*element).GetOutgoingAssociation())
	var err error
	if (*element).GetType() == BPMN20.ExclusiveGateway {
		nextFlows, err = exclusivelyFilterByConditionExpression(nextFlows, instance.VariableHolder.Variables())
		if err != nil {
			instance.State = Failed
			cmds = append(cmds, tErrorCommand{
				err:         err,
				elementId:   (*element).GetId(),
				elementName: (*element).GetName(),
			})
			return cmds
		}
	}
	for _, flow := range nextFlows {
		cmds = append(cmds, tFlowTransitionCommand{
			sourceId:       (*element).GetId(),
			sourceActivity: activity,
			sequenceFlowId: flow.Id,
		})
	}
	return cmds
}

func (state *BpmnEngineState) handleIntermediateCatchEvent(process *ProcessInfo, instance *processInstanceInfo, ice BPMN20.TIntermediateCatchEvent, activity Activity) (continueFlow bool, err error) {
	if ice.MessageEventDefinition.Id != "" {
		var ms *MessageSubscription
		continueFlow, ms, err = state.handleIntermediateMessageCatchEvent(process, instance, ice)
		ms.originActivity = activity
		return continueFlow, err
	}
	if ice.TimerEventDefinition.Id != "" {
		var timer *Timer
		continueFlow, timer, err = state.handleIntermediateTimerCatchEvent(process, instance, ice)
		timer.originActivity = activity
		return continueFlow, err
	}
	return false, err
}

func (state *BpmnEngineState) handleEndEvent(process *ProcessInfo, instance *processInstanceInfo) {
	completedJobs := true
	for _, job := range state.jobs {
		if job.ProcessInstanceKey == instance.GetInstanceKey() && (job.JobState == Ready || job.JobState == Active) {
			completedJobs = false
			break
		}
	}
	if completedJobs && !state.hasActiveSubscriptions(process, instance) {
		instance.State = Completed
	}
}

func (state *BpmnEngineState) handleParallelGateway(instance *processInstanceInfo, element BPMN20.TParallelGateway, activity Activity, inboundFlowId string) bool {
	existingActivity := instance.findActiveActivityByElementId(element.Id)
	if existingActivity != nil {
		activity = existingActivity
	} else {
		instance.appendActivity(activity)
	}
	var ga GatewayActivity
	ga = activity.(GatewayActivity)
	ga.SetInboundFlowIdCompleted(inboundFlowId)
	if ga.IsParallel() && ga.AreInboundFlowsCompleted() {
		ga.SetState(Completed)
	}
	return ga.IsParallel() && ga.AreInboundFlowsCompleted()
}

func (state *BpmnEngineState) hasActiveSubscriptions(process *ProcessInfo, instance *processInstanceInfo) bool {
	activeSubscriptions := map[string]bool{}
	for _, ms := range state.messageSubscriptions {
		if ms.ProcessInstanceKey == instance.GetInstanceKey() {
			activeSubscriptions[ms.ElementId] = ms.State == Ready || ms.State == Active
		}
	}
	// eliminate the active subscriptions, which are from one 'parent' EventBasedGateway
	for _, gateway := range process.definitions.Process.EventBasedGateway {
		flows := BPMN20.FindSequenceFlows(&process.definitions.Process.SequenceFlows, gateway.OutgoingAssociation)
		isOneEventCompleted := true
		for _, flow := range flows {
			isOneEventCompleted = isOneEventCompleted && !activeSubscriptions[flow.TargetRef]
		}
		for _, flow := range flows {
			activeSubscriptions[flow.TargetRef] = isOneEventCompleted
		}
	}
	for _, v := range activeSubscriptions {
		if v {
			return true
		}
	}
	return false
}
