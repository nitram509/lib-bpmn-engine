package bpmn_engine

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/exporter"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/activity"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
	"hash/adler32"
	"os"
	"time"
)

type BpmnEngine interface {
	LoadFromFile(filename string) (*ProcessInfo, error)
	LoadFromBytes(xmlData []byte) (*ProcessInfo, error)
	AddTaskHandler(taskId string, handler func(job ActivatedJob))
	CreateInstance(processKey int64, variableContext map[string]interface{}) (*ProcessInstanceInfo, error)
	CreateAndRunInstance(processKey int64, variableContext map[string]interface{}) (*ProcessInstanceInfo, error)
	RunOrContinueInstance(processInstanceKey int64) (*ProcessInstanceInfo, error)
	GetName() string
	GetProcessInstances() []*ProcessInstanceInfo
	FindProcessInstanceById(processInstanceKey int64) *ProcessInstanceInfo
}

const continueNextElement = true

// New creates an engine with an arbitrary name of the engine;
// useful in case you have multiple ones, in order to distinguish them.
func New(name string) BpmnEngineState {
	snowflakeIdGenerator := initializeSnowflakeIdGenerator()
	return BpmnEngineState{
		name:                 name,
		processes:            []ProcessInfo{},
		processInstances:     []*ProcessInstanceInfo{},
		handlers:             map[string]func(job ActivatedJob){},
		jobs:                 []*job{},
		messageSubscriptions: []*MessageSubscription{},
		snowflake:            snowflakeIdGenerator,
		exporters:            []exporter.EventExporter{},
	}
}

// CreateInstance creates a new instance for a process with given processKey
// will return (nil, nil), when no process with given was found
func (state *BpmnEngineState) CreateInstance(processKey int64, variableContext map[string]interface{}) (*ProcessInstanceInfo, error) {
	if variableContext == nil {
		variableContext = map[string]interface{}{}
	}
	for _, process := range state.processes {
		if process.ProcessKey == processKey {
			processInstanceInfo := ProcessInstanceInfo{
				processInfo:     &process,
				instanceKey:     state.generateKey(),
				variableContext: variableContext,
				createdAt:       time.Now(),
				state:           process_instance.READY,
			}
			state.processInstances = append(state.processInstances, &processInstanceInfo)
			state.exportProcessInstanceEvent(process, processInstanceInfo)
			return &processInstanceInfo, nil
		}
	}
	return nil, nil
}

// CreateAndRunInstance creates a new instance and executes it immediately.
// The provided variableContext can be nil or refers to a variable map,
// which is provided to every service task handler function.
func (state *BpmnEngineState) CreateAndRunInstance(processKey int64, variableContext map[string]interface{}) (*ProcessInstanceInfo, error) {
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
func (state *BpmnEngineState) RunOrContinueInstance(processInstanceKey int64) (*ProcessInstanceInfo, error) {
	for _, pi := range state.processInstances {
		if processInstanceKey == pi.instanceKey {
			return pi, state.run(pi)
		}
	}
	return nil, nil
}

func (state *BpmnEngineState) run(instance *ProcessInstanceInfo) (err error) {
	type queueElement struct {
		inboundFlowId string
		baseElement   BPMN20.BaseElement
	}

	queue := make([]queueElement, 0)
	process := instance.processInfo

	switch instance.state {
	case process_instance.READY:
		// use start events to start the instance
		for _, event := range process.definitions.Process.StartEvents {
			queue = append(queue, queueElement{
				inboundFlowId: "",
				baseElement:   event,
			})
		}
		instance.state = process_instance.ACTIVE
	case process_instance.ACTIVE:
		intermediateCatchEvents := state.findIntermediateCatchEventsForContinuation(process, instance)
		for _, ice := range intermediateCatchEvents {
			queue = append(queue, queueElement{
				inboundFlowId: "",
				baseElement:   ice,
			})
		}
	case process_instance.COMPLETED:
		return nil
	case process_instance.FAILED:
		return nil
	default:
		panic("Unknown process instance state.")
	}

	for len(queue) > 0 {
		element := queue[0].baseElement
		inboundFlowId := queue[0].inboundFlowId
		queue = queue[1:]

		continueNextElement := state.handleElement(process, instance, element)

		if continueNextElement {
			state.exportElementEvent(*process, *instance, element, exporter.ElementCompleted)

			if inboundFlowId != "" {
				state.scheduledFlows = remove(state.scheduledFlows, inboundFlowId)
			}
			nextFlows := BPMN20.FindSequenceFlows(&process.definitions.Process.SequenceFlows, element.GetOutgoingAssociation())
			if element.GetType() == BPMN20.ExclusiveGateway {
				nextFlows, err = exclusivelyFilterByConditionExpression(nextFlows, instance.variableContext)
				if err != nil {
					instance.state = process_instance.FAILED
					break
				}
			}
			for _, flow := range nextFlows {

				state.exportSequenceFlowEvent(*process, *instance, flow)

				// TODO: create test for that
				//if len(flows) < 1 {
				//	panic(fmt.Sprintf("Can't find 'sequenceFlow' element with ID=%s. "+
				//		"This is likely because your BPMN is invalid.", flows[0]))
				//}
				state.scheduledFlows = append(state.scheduledFlows, flow.Id)
				baseElements := BPMN20.FindBaseElementsById(process.definitions, flow.TargetRef)
				// TODO: create test for that
				//if len(baseElements) < 1 {
				//	panic(fmt.Sprintf("Can't find flow element with ID=%s. "+
				//		"This is likely because there are elements in the definition, "+
				//		"which this engine does not support (yet).", flow.Id))
				//}
				targetBaseElement := baseElements[0]
				queue = append(queue, queueElement{
					inboundFlowId: flow.Id,
					baseElement:   targetBaseElement,
				})
			}
		}
	}

	if instance.state == process_instance.COMPLETED || instance.state == process_instance.FAILED {
		// TODO need to send failed state
		state.exportEndProcessEvent(*process, *instance)
	}

	return err
}

func (state *BpmnEngineState) findIntermediateCatchEventsForContinuation(process *ProcessInfo, instance *ProcessInstanceInfo) (ret []*BPMN20.TIntermediateCatchEvent) {
	var messageRef2IntermediateCatchEventMapping = map[string]BPMN20.TIntermediateCatchEvent{}
	for _, ice := range process.definitions.Process.IntermediateCatchEvent {
		messageRef2IntermediateCatchEventMapping[ice.MessageEventDefinition.MessageRef] = ice
	}
	for _, caughtEvent := range instance.caughtEvents {
		if caughtEvent.IsConsumed == true {
			// skip consumed ones
			continue
		}
		for _, msg := range process.definitions.Messages {
			// find the matching message definition
			if msg.Name == caughtEvent.Name {
				// find potential event definitions
				event := messageRef2IntermediateCatchEventMapping[msg.Id]
				if state.hasActiveMessageSubscriptionForId(event.Id) {
					ret = append(ret, &event)
				}
			}
		}
	}
	ice := checkDueTimersAndFindIntermediateCatchEvent(state.timers, process.definitions.Process.IntermediateCatchEvent, instance)
	if ice != nil {
		ret = append(ret, ice)
	}
	return eliminateEventsWhichComeFromTheSameGateway(process.definitions, ret)
}

func (state *BpmnEngineState) hasActiveMessageSubscriptionForId(id string) bool {
	for _, subscription := range state.messageSubscriptions {
		if id == subscription.ElementId && subscription.State == activity.Active {
			return true
		}
	}
	return false
}

func eliminateEventsWhichComeFromTheSameGateway(definitions BPMN20.TDefinitions, events []*BPMN20.TIntermediateCatchEvent) (ret []*BPMN20.TIntermediateCatchEvent) {
	// a bubble-sort-like approach to find elements, which have the same incoming association
	for len(events) > 0 {
		event := events[0]
		events = events[1:]
		if event == nil {
			continue
		}
		ret = append(ret, event)
		for i := 0; i < len(events); i++ {
			if haveEqualInboundBaseElement(definitions, event, events[i]) {
				events[i] = nil
			}
		}
	}
	return ret
}

func haveEqualInboundBaseElement(definitions BPMN20.TDefinitions, event1 *BPMN20.TIntermediateCatchEvent, event2 *BPMN20.TIntermediateCatchEvent) bool {
	if event1 == nil || event2 == nil {
		return false
	}
	checkOnlyOneAssociationOrPanic(event1)
	checkOnlyOneAssociationOrPanic(event2)
	ref1 := BPMN20.FindSourceRefs(definitions.Process.SequenceFlows, event1.IncomingAssociation[0])[0]
	ref2 := BPMN20.FindSourceRefs(definitions.Process.SequenceFlows, event2.IncomingAssociation[0])[0]
	baseElement1 := BPMN20.FindBaseElementsById(definitions, ref1)[0]
	baseElement2 := BPMN20.FindBaseElementsById(definitions, ref2)[0]
	return baseElement1.GetId() == baseElement2.GetId()
}

func checkOnlyOneAssociationOrPanic(event *BPMN20.TIntermediateCatchEvent) {
	if len(event.IncomingAssociation) != 1 {
		panic(fmt.Sprintf("Element with id=%s has %d incoming associations, but only 1 is supported by this engine.",
			event.Id, len(event.IncomingAssociation)))
	}
}

// AddTaskHandler registers a handler function to be called for service tasks with a given taskId
func (state *BpmnEngineState) AddTaskHandler(taskId string, handler func(job ActivatedJob)) {
	if nil == state.handlers {
		state.handlers = make(map[string]func(job ActivatedJob))
	}
	state.handlers[taskId] = handler
}

func (state *BpmnEngineState) handleElement(process *ProcessInfo, instance *ProcessInstanceInfo, element BPMN20.BaseElement) bool {
	id := element.GetId()
	state.exportElementEvent(*process, *instance, element, exporter.ElementActivated)
	switch element.GetType() {
	case BPMN20.StartEvent:
		return true
	case BPMN20.ServiceTask:
		return state.handleServiceTask(id, process, instance)
	case BPMN20.ParallelGateway:
		return state.handleParallelGateway(element)
	case BPMN20.EndEvent:
		state.handleEndEvent(instance)
		state.exportElementEvent(*process, *instance, element, exporter.ElementCompleted) // special case here, to end the instance
		return false
	case BPMN20.IntermediateCatchEvent:
		return state.handleIntermediateCatchEvent(process, instance, element)
	case BPMN20.EventBasedGateway:
		// TODO improve precondition tests
		// simply proceed
		return true
	default:
		// do nothing
		// TODO: should we print a warning?
	}
	return true
}

func (state *BpmnEngineState) handleIntermediateCatchEvent(process *ProcessInfo, instance *ProcessInstanceInfo, element BPMN20.BaseElement) bool {
	for _, ice := range process.definitions.Process.IntermediateCatchEvent {
		if ice.Id == element.GetId() {
			if ice.MessageEventDefinition.Id != "" {
				return state.handleIntermediateMessageCatchEvent(ice.Id, element.GetName(), instance)
			}
			if ice.TimerEventDefinition.Id != "" {
				return state.handleIntermediateTimerCatchEvent(process, instance, ice)
			}
		}
	}
	return false
}

func (state *BpmnEngineState) handleParallelGateway(element BPMN20.BaseElement) bool {
	// check incoming flows, if ready, then continue
	allInboundsAreScheduled := true
	for _, inFlowId := range element.GetIncomingAssociation() {
		allInboundsAreScheduled = contains(state.scheduledFlows, inFlowId) && allInboundsAreScheduled
	}
	return allInboundsAreScheduled
}

func (state *BpmnEngineState) handleEndEvent(instance *ProcessInstanceInfo) {
	var activeSubscriptions = false
	for _, ms := range state.messageSubscriptions {
		if ms.ProcessInstanceKey == instance.GetInstanceKey() && ms.State == activity.Ready {
			activeSubscriptions = true
			break
		}
	}
	var completedJobs = true
	for _, job := range state.jobs {
		if job.ProcessInstanceKey == instance.GetInstanceKey() && job.State != activity.Completed {
			completedJobs = false
			break
		}
	}
	if completedJobs && !activeSubscriptions {
		instance.state = process_instance.COMPLETED
	}
}

func (state *BpmnEngineState) generateKey() int64 {
	return state.snowflake.Generate().Int64()
}

func initializeSnowflakeIdGenerator() *snowflake.Node {
	hash32 := adler32.New()
	for _, e := range os.Environ() {
		hash32.Sum([]byte(e))
	}
	snowflakeNode, err := snowflake.NewNode(int64(hash32.Sum32()))
	if err != nil {
		panic("Can't initialize snowflake ID generator. Message: " + err.Error())
	}
	return snowflakeNode
}
