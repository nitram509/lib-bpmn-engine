package bpmn_engine

import (
	"fmt"
	"sort"
	"time"

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
				VariableHolder: NewVarHolder(nil, variableContext),
				CreatedAt:      time.Now(),
				ActivityState:  Ready,
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
	return instance, state.run(instance.ProcessInfo.definitions.Process, instance, instance)
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
	return instance, state.run(instance.ProcessInfo.definitions.Process, instance, instance)
}

// RunOrContinueInstance runs or continues a process instance by a given processInstanceKey.
// returns the process instances, when found;
// does nothing, if process is already in ProcessInstanceCompleted State;
// returns nil, nil when no process instance was found;
// might return BpmnEngineError or ExpressionEvaluationError.
func (state *BpmnEngineState) RunOrContinueInstance(processInstanceKey int64) (*processInstanceInfo, error) {
	for _, pi := range state.processInstances {
		if processInstanceKey == pi.InstanceKey {
			return pi, state.run(pi.ProcessInfo.definitions.Process, pi, pi)
		}
	}
	return nil, nil
}

func (state *BpmnEngineState) run(process BPMN20.ProcessElement, instance *processInstanceInfo, currentActivity activity) (err error) {
	var commandQueue []command

	switch currentActivity.State() {
	case Ready:
		// use start events to start the instance
		for _, startEvent := range process.GetStartEvents() {
			var be BPMN20.BaseElement = startEvent
			commandQueue = append(commandQueue, activityCommand{
				element: &be,
			})
		}
		currentActivity.SetState(Active)
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
			nextFlows := BPMN20.FindSequenceFlows(process, []string{flowId})
			if BPMN20.ExclusiveGateway == (*sourceActivity.Element()).GetType() {
				nextFlows, err = exclusivelyFilterByConditionExpression(nextFlows, instance.VariableHolder.Variables())
				if err != nil {
					instance.ActivityState = Failed
					return err
				}
			}
			for _, flow := range nextFlows {
				state.exportSequenceFlowEvent(*instance.ProcessInfo, *instance, flow)
				baseElements := BPMN20.FindBaseElementsById(process, flow.TargetRef)
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
			nextCommands := state.handleElement(process, currentActivity, instance, element, originActivity)
			state.exportElementEvent(process, *instance, *element, exporter.ElementCompleted)
			commandQueue = append(commandQueue, nextCommands...)
		case continueActivityType:
			element := cmd.(continueActivityCommand).activity.Element()
			originActivity := cmd.(continueActivityCommand).originActivity
			nextCommands := state.handleElement(process, currentActivity, instance, element, originActivity)
			commandQueue = append(commandQueue, nextCommands...)
		case errorType:
			err = cmd.(errorCommand).err
			instance.ActivityState = Failed
			// *activityState = Failed            // TODO: check if meaningful
			break
		case eventSubProcessCompletedType:
			subProcessActivity := cmd.(eventSubProcessCompletedCommand).activity
			instance.SetState(subProcessActivity.State())
			state.exportElementEvent(process, *instance, process, exporter.ElementCompleted)
			break
		case checkExclusiveGatewayDoneType:
			activity := cmd.(checkExclusiveGatewayDoneCommand).gatewayActivity
			state.checkExclusiveGatewayDone(activity)
		default:
			return newEngineErrorf("[invariant check] command type check not fully implemented")
		}
	}

	if instance.ActivityState == Completed || instance.ActivityState == Failed {
		// TODO need to send failed State
		state.exportEndProcessEvent(*instance.ProcessInfo, *instance)
	}

	return err
}

func (state *BpmnEngineState) handleElement(process BPMN20.ProcessElement, act activity, instance *processInstanceInfo, element *BPMN20.BaseElement, originActivity activity) []command {
	state.exportElementEvent(process, *instance, *element, exporter.ElementActivated) // FIXME: don't create event on continuation ?!?!
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
		createFlowTransitions = state.handleEndEvent(process, act, instance)
		activity = act
		state.exportElementEvent(process, *instance, *element, exporter.ElementCompleted) // special case here, to end the instance
	case BPMN20.ServiceTask:
		taskElement := (*element).(BPMN20.TaskElement)
		_, job, jobErr := state.handleServiceTask(process, instance, &taskElement)
		err = jobErr
		activity = job
		if err != nil {
			nextCommands = append(nextCommands, errorCommand{
				err:         err,
				elementId:   (*element).GetId(),
				elementName: (*element).GetName(),
			})
		} else if job.ErrorCode != "" {
			// The current process will remain ACTIVE until the event sub-processes have completed.
			nextCommands = handleErrorEvent(process, instance, element, job.ErrorCode)
			createFlowTransitions = false // TODO confirm
		} else {
			// Only follow sequence flow if there are no Technical or Business Errors
			createFlowTransitions = activity.State() == Completed
		}
	case BPMN20.UserTask:
		taskElement := (*element).(BPMN20.TaskElement)
		job, jobErr := state.handleUserTask(process, instance, &taskElement)
		err = jobErr
		activity = job
		if err != nil {
			nextCommands = append(nextCommands, errorCommand{
				err:         err,
				elementId:   (*element).GetId(),
				elementName: (*element).GetName(),
			})
		} else if job.ErrorCode != "" {
			nextCommands = handleErrorEvent(process, instance, element, job.ErrorCode)
			createFlowTransitions = false
		} else {
			// Only follow sequence flow if there are no Technical or Business Errors
			createFlowTransitions = activity.State() == Completed
		}
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
		activity = &elementActivity{
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
	case BPMN20.InclusiveGateway:
		activity = &elementActivity{
			key:     state.generateKey(),
			state:   Active,
			element: element,
		}
		createFlowTransitions = true
	case BPMN20.SubProcess:
		subProcessElement := (*element).(BPMN20.TSubProcess)
		subProcess, subProcessErr := state.handleSubProcess(instance, &subProcessElement)
		activity = subProcess
		err = subProcessErr
		if err != nil {
			nextCommands = append(nextCommands, errorCommand{
				err:         err,
				elementId:   (*element).GetId(),
				elementName: (*element).GetName(),
			})
		} else if subProcessElement.TriggeredByEvent {
			// We need to complete the parent process when an event sub-process has completed. but we cant do it here
			nextCommands = append(nextCommands, eventSubProcessCompletedCommand{
				activity: subProcess,
			})
		}
		createFlowTransitions = activity.State() == Completed
	case BPMN20.BoundaryEvent:
		boundary := (*element).(BPMN20.TBoundaryEvent)
		activity, err = state.handleBoundaryEvent(&boundary, instance)
	default:
		nextCommands = append(nextCommands, errorCommand{
			err:         newEngineErrorf("[invariant check] unsupported element: id=%s, type=%s", (*element).GetId(), (*element).GetType()),
			elementId:   (*element).GetId(),
			elementName: (*element).GetName(),
		})
	}
	if createFlowTransitions && err == nil {
		nextCommands = append(nextCommands, createNextCommands(process, instance, element, activity)...)
	}
	return nextCommands
}

func handleErrorEvent(process BPMN20.ProcessElement, instance *processInstanceInfo, element *BPMN20.BaseElement, errorCode string) []command {
	// Find the error by code on the process
	if errT, found := findErrorDefinition(instance.ProcessInfo.definitions, errorCode); found {

		// Find the boundary events for the task
		boundaryEvents := findBoundaryEventsForTypeAndReference(instance.ProcessInfo.definitions, BPMN20.ErrorBoundary, (*element).GetId())
		if boundaryEvent, foundBoundary := findBoundaryEventForError(boundaryEvents, errT.Id); foundBoundary {
			return []command{
				activityCommand{element: BPMN20.Ptr[BPMN20.BaseElement](boundaryEvent)},
			}
		}

		// If we still haven't found a command then we should look to see if there is an event sub process we can follow
		if subProcess, subFound := findEventSubprocessForError(process, errT.Id); subFound {
			return []command{
				activityCommand{element: BPMN20.Ptr[BPMN20.BaseElement](subProcess)},
			}
		}

		// If not see if there is a catch-all boundary event
		if boundaryEvent, foundBoundary := findBoundaryEventForError(boundaryEvents, ""); foundBoundary {
			return []command{
				activityCommand{element: BPMN20.Ptr[BPMN20.BaseElement](boundaryEvent)},
			}
		}

		// If not find an event sub process matching catchall
		if subProcess, subFound := findEventSubprocessForError(process, ""); subFound {
			return []command{
				activityCommand{element: BPMN20.Ptr[BPMN20.BaseElement](subProcess)},
			}
		}

		// TODO continue lookup up to the parent process if this is a sub process

		return []command{
			errorCommand{
				err:         newEngineErrorf("Could not find suitable handler for ErrorCode event id=%s, code=%s", errT.Id, errT.ErrorCode),
				elementId:   (*element).GetId(),
				elementName: (*element).GetName(),
			},
		}
	} else {
		return []command{
			errorCommand{
				err:         newEngineErrorf("Could not find error definition \"%s\"", errorCode),
				elementId:   (*element).GetId(),
				elementName: (*element).GetName(),
			},
		}
	}
}

func findEventSubprocessForError(process BPMN20.ProcessElement, errorReferenceID string) (BPMN20.TSubProcess, bool) {
	// Look for event sub-processes in the process
	for _, subProcess := range process.GetSubProcess() {
		// Check if this is an event sub-process (triggered by event)
		if subProcess.TriggeredByEvent {
			// Look for start events in the sub-process
			for _, startEvent := range subProcess.StartEvents {
				// Check if this start event has an error event definition
				if startEvent.ErrorEventDefinition.ErrorRef == errorReferenceID {
					// We found an event sub-process with an error start event
					return subProcess, true
				}
			}
		}
	}

	// No matching event sub-process found
	return BPMN20.TSubProcess{}, false
}

// findBoundaryEventsForReference finds all boundary events attached to the provided element
func findBoundaryEventsForTypeAndReference(definitions BPMN20.TDefinitions, boundaryType BPMN20.BoundaryType, referenceID string) []BPMN20.TBoundaryEvent {
	boundaryEvents := make([]BPMN20.TBoundaryEvent, 0)
	for _, boundary := range definitions.Process.BoundaryEvent {
		if boundary.AttachedToRef == referenceID && boundary.GetBoundaryType() == boundaryType {
			boundaryEvents = append(boundaryEvents, boundary)
		}
	}
	return boundaryEvents
}

func findBoundaryEventForError(boundaryEvents []BPMN20.TBoundaryEvent, errorID string) (BPMN20.TBoundaryEvent, bool) {
	for _, boundaryEvent := range boundaryEvents {
		// Check if this boundary event has an error event definition
		if boundaryEvent.ErrorEventDefinition.ErrorRef == errorID {
			return boundaryEvent, true
		}
	}
	return BPMN20.TBoundaryEvent{}, false
}

func findErrorDefinition(definitions BPMN20.TDefinitions, errorCode string) (BPMN20.TError, bool) {

	// Iterate through all errors in the definitions
	for _, err := range definitions.Errors {
		// Check if the error code matches the requested code
		if err.ErrorCode == errorCode {
			return err, true
		}
	}

	// Return empty error if not found
	return BPMN20.TError{}, false

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

func createNextCommands(process BPMN20.ProcessElement, instance *processInstanceInfo, element *BPMN20.BaseElement, activity activity) (cmds []command) {
	nextFlows := BPMN20.FindSequenceFlows(process, (*element).GetOutgoingAssociation())
	var err error
	switch (*element).GetType() {
	case BPMN20.ExclusiveGateway:
		nextFlows, err = exclusivelyFilterByConditionExpression(nextFlows, instance.VariableHolder.Variables())
		if err != nil {
			instance.ActivityState = Failed
			cmds = append(cmds, errorCommand{
				err:         err,
				elementId:   (*element).GetId(),
				elementName: (*element).GetName(),
			})
			return cmds
		}
	case BPMN20.InclusiveGateway:
		nextFlows, err = inclusivelyFilterByConditionExpression(nextFlows, instance.VariableHolder.Variables())
		if err != nil {
			instance.ActivityState = Failed
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

func (state *BpmnEngineState) handleIntermediateCatchEvent(process BPMN20.ProcessElement, instance *processInstanceInfo, ice BPMN20.TIntermediateCatchEvent, originActivity activity) (continueFlow bool, activity activity, err error) {
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
		elementVarHolder := NewVarHolder(&instance.VariableHolder, nil)
		if err := propagateProcessInstanceVariables(&elementVarHolder, ice.Output); err != nil {
			msg := fmt.Sprintf("Can't evaluate expression in element id=%s name=%s", ice.Id, ice.Name)
			err = &ExpressionEvaluationError{Msg: msg, Err: err}
		} else {
			continueFlow = throwLinkName == catchLinkName // just stating the obvious
		}
	}
	return continueFlow, activity, err
}

func (state *BpmnEngineState) handleEndEvent(process BPMN20.ProcessElement, act activity, instance *processInstanceInfo) bool {
	activeMessageSubscriptions := false
	for _, ms := range state.messageSubscriptions {
		if ms.ProcessInstanceKey == instance.InstanceKey {
			activeMessageSubscriptions = activeMessageSubscriptions || ms.State() == Active || ms.State() == Ready
		}
		if activeMessageSubscriptions {
			break
		}
	}
	if !activeMessageSubscriptions {
		act.SetState(Completed)
	}
	switch process.(type) {
	case *BPMN20.TProcess:
		return false
	case *BPMN20.TSubProcess:
		act.SetState(Completed)
		return true
	}
	return false
}

func (state *BpmnEngineState) handleParallelGateway(process BPMN20.ProcessElement, instance *processInstanceInfo, element BPMN20.TParallelGateway, originActivity activity) (continueFlow bool, resultActivity activity) {
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
	sourceFlow := BPMN20.FindFirstSequenceFlow(process, (*originActivity.Element()).GetId(), element.GetId())
	resultActivity.(*gatewayActivity).SetInboundFlowCompleted(sourceFlow.Id)
	continueFlow = resultActivity.(*gatewayActivity).parallel && resultActivity.(*gatewayActivity).AreInboundFlowsCompleted()
	if continueFlow {
		resultActivity.(*gatewayActivity).SetState(Completed)
	}
	return continueFlow, resultActivity
}

func (state *BpmnEngineState) handleSubProcess(instance *processInstanceInfo, subProcessElement *BPMN20.TSubProcess) (subProcessActivity activity, err error) {
	var be BPMN20.BaseElement = subProcessElement
	subProcessActivity = &subProcessInfo{
		ElementId:       subProcessElement.GetId(),
		ProcessInstance: instance,
		ProcessId:       state.generateKey(),
		CreatedAt:       time.Now(),
		processState:    Ready,
		variableHolder:  NewVarHolder(&instance.VariableHolder, nil),
		baseElement:     &be,
	}
	err = state.run(subProcessElement, instance, subProcessActivity)
	return subProcessActivity, err
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

func (state *BpmnEngineState) handleBoundaryEvent(element *BPMN20.TBoundaryEvent, instance *processInstanceInfo) (activity, error) {
	var be BPMN20.BaseElement = element
	activity := &elementActivity{
		key:     state.generateKey(),
		state:   Completed,
		element: &be,
	}
	variableHolder := NewVarHolder(&instance.VariableHolder, nil)
	err := propagateProcessInstanceVariables(&variableHolder, element.GetOutputMapping())
	if err != nil {
		instance.ActivityState = Failed
	}

	return activity, err
}
