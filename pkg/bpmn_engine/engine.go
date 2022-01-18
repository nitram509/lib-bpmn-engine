package bpmn_engine

import (
	"crypto/md5"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/activity"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
	"io/ioutil"
	"time"
)

type BpmnEngine interface {
	LoadFromFile(filename string) (*ProcessInfo, error)
	LoadFromBytes(xmlData []byte) (*ProcessInfo, error)
	AddTaskHandler(taskType string, handler func(context ProcessInstanceContext))
	CreateInstance(processKey int64, variableContext map[string]string) (*ProcessInstanceInfo, error)
	CreateAndRunInstance(processKey int64, variableContext map[string]string) (*ProcessInstanceInfo, error)
	RunOrContinueInstance(processInstanceKey int64) (*ProcessInstanceInfo, error)
	GetName() string
	GetProcessInstances() []*ProcessInstanceInfo
}

const ContinueNextElement = true

// New creates an engine with an arbitrary name of the engine;
// useful in case you have multiple ones
func New(name string) BpmnEngineState {
	return BpmnEngineState{
		name:                 name,
		processes:            []ProcessInfo{},
		processInstances:     []*ProcessInstanceInfo{},
		handlers:             map[string]func(context ProcessInstanceContext){},
		jobs:                 []*Job{},
		messageSubscriptions: []*MessageSubscription{},
	}
}

// CreateInstance creates a new instance for a process with given processKey
func (state *BpmnEngineState) CreateInstance(processKey int64, variableContext map[string]string) (*ProcessInstanceInfo, error) {
	if variableContext == nil {
		variableContext = map[string]string{}
	}
	for _, process := range state.processes {
		if process.ProcessKey == processKey {
			processInstanceInfo := ProcessInstanceInfo{
				processInfo:     &process,
				instanceKey:     generateKey(),
				variableContext: variableContext,
				createdAt:       time.Now(),
				state:           process_instance.READY,
			}
			state.processInstances = append(state.processInstances, &processInstanceInfo)
			return &processInstanceInfo, nil
		}
	}
	return nil, nil
}

// CreateAndRunInstance creates a new instance and executes it immediately.
// The provided variableContext can be nil
func (state *BpmnEngineState) CreateAndRunInstance(processKey int64, variableContext map[string]string) (*ProcessInstanceInfo, error) {
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
func (state *BpmnEngineState) RunOrContinueInstance(processInstanceKey int64) (*ProcessInstanceInfo, error) {
	for _, pi := range state.processInstances {
		if processInstanceKey == pi.instanceKey {
			return pi, state.run(pi)
		}
	}
	return nil, nil
}

func (state *BpmnEngineState) run(instance *ProcessInstanceInfo) error {
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
	default:
		panic("Unknown process instance state.")
	}

	for len(queue) > 0 {
		element := queue[0].baseElement
		inboundFlowId := queue[0].inboundFlowId
		queue = queue[1:]

		continueNextElement := state.handleElement(process, instance, element)

		if continueNextElement {
			if inboundFlowId != "" {
				state.scheduledFlows = remove(state.scheduledFlows, inboundFlowId)
			}
			for _, flowId := range element.GetOutgoingAssociation() {
				targetRef := BPMN20.FindTargetRefs(process.definitions.Process.SequenceFlows, flowId)
				state.scheduledFlows = append(state.scheduledFlows, flowId)
				baseElements := BPMN20.FindBaseElementsById(process.definitions, targetRef[0])
				if len(baseElements) < 1 {
					panic(fmt.Sprintf("Can't find flow element with ID=%s. "+
						"This is likely because there are elements in the definition, "+
						"which this engine does not support (yet).", targetRef[0]))
				}
				targetBaseElement := baseElements[0]
				queue = append(queue, queueElement{
					inboundFlowId: flowId,
					baseElement:   targetBaseElement,
				})
			}
		}
	}
	return nil
}

func (state *BpmnEngineState) findIntermediateCatchEventsForContinuation(process *ProcessInfo, instance *ProcessInstanceInfo) (ret []*BPMN20.TIntermediateCatchEvent) {
	for _, event := range instance.caughtEvents {
		for _, ice := range process.definitions.Process.IntermediateCatchEvent {
			if event.Name == ice.Name {
				ret = append(ret, &ice)
			}
		}
	}
	ice := checkDueTimersAndFindIntermediateCatchEvent(state.timers, process.definitions.Process.IntermediateCatchEvent, instance)
	if ice != nil {
		ret = append(ret, ice)
	}
	return eliminateEventsWhichComeFromTheSameGateway(process.definitions, ret)
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

func (state *BpmnEngineState) LoadFromFile(filename string) (*ProcessInfo, error) {
	xmlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return state.LoadFromBytes(xmlData)
}

func (state *BpmnEngineState) LoadFromBytes(xmlData []byte) (*ProcessInfo, error) {
	md5sum := md5.Sum(xmlData)
	var definitions BPMN20.TDefinitions
	err := xml.Unmarshal(xmlData, &definitions)
	if err != nil {
		return nil, err
	}

	processInfo := ProcessInfo{
		Version:     1,
		definitions: definitions,
	}
	for _, process := range state.processes {
		if process.BpmnProcessId == definitions.Process.Id {
			if areEqual(process.checksumBytes, md5sum) {
				return &process, nil
			} else {
				processInfo.Version = process.Version + 1
			}
		}
	}
	processInfo.BpmnProcessId = definitions.Process.Id
	processInfo.ProcessKey = generateKey()
	processInfo.checksumBytes = md5sum
	state.processes = append(state.processes, processInfo)

	return &processInfo, nil
}

// AddTaskHandler registers a handler for a given taskType
func (state *BpmnEngineState) AddTaskHandler(taskId string, handler func(context ProcessInstanceContext)) {
	if nil == state.handlers {
		state.handlers = make(map[string]func(context ProcessInstanceContext))
	}
	state.handlers[taskId] = handler
}

func (state *BpmnEngineState) handleElement(process *ProcessInfo, instance *ProcessInstanceInfo, element BPMN20.BaseElement) bool {
	id := element.GetId()
	switch element.GetType() {
	case BPMN20.ServiceTask:
		state.handleServiceTask(id, process, instance)
	case BPMN20.ParallelGateway:
		return state.handleParallelGateway(element)
	case BPMN20.EndEvent:
		state.handleEndEvent(instance)
		return false
	case BPMN20.IntermediateCatchEvent:
		return state.handleIntermediateCatchEvent(process, instance, element)
	case BPMN20.EventBasedGateway:
		// TODO improve precondition tests
		// simply proceed
		return true
	default:
		// do nothing
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

func (state *BpmnEngineState) findProcessInstance(processInstanceKey int64) *ProcessInstanceInfo {
	for _, pi := range state.processInstances {
		if pi.GetInstanceKey() == processInstanceKey {
			return pi
		}
	}
	return nil
}

func generateKey() int64 {
	return time.Now().UnixNano() << 1
}
