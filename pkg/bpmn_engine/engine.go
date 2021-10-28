package bpmn_engine

import (
	"crypto/md5"
	"encoding/xml"
	"errors"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/zeebe"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"io/ioutil"
	"time"
)

type Node struct {
	Name string
	Id   string
}

type BpmnEngine interface {
	LoadFromFile(filename string) (zeebe.WorkflowMetadata, error)
	LoadFromBytes(xmldata []byte, resourceName string) (zeebe.WorkflowMetadata, error)
	AddTaskHandler(taskId string, handler func(id string))
	GetProcesses() []zeebe.WorkflowMetadata
}

func NewNamedResourceState() *BpmnEngineNamedResourceState {
	return &BpmnEngineNamedResourceState{}
}

type BpmnEngineNamedResourceState struct {
	processes   []zeebe.WorkflowMetadata
	definitions BPMN20.TDefinitions
	queue       []BPMN20.BaseElement
	handlers    map[string]func(id string)
}

type BpmnEngineState struct {
	states map[string]*BpmnEngineNamedResourceState
}

func New() BpmnEngineState {
	return BpmnEngineState{
		states: map[string]*BpmnEngineNamedResourceState{},
	}
}

func (state *BpmnEngineState) GetProcesses(resourceName string) []zeebe.WorkflowMetadata {
	value, ok := state.states[resourceName]
	if !ok {
		return []zeebe.WorkflowMetadata{}
	}

	return value.processes
}

type ProcessInstance struct {
	WorkflowMetadata zeebe.WorkflowMetadata
}

func (state *BpmnEngineState) Execute(resourceName string) error {
	theState, ok := state.states[resourceName]
	if !ok {
		return errors.New("resource name not found")
	}

	queue := make([]BPMN20.BaseElement, 0)
	for _, event := range theState.definitions.Process.StartEvents {
		queue = append(queue, event)
	}
	theState.queue = queue

	for len(queue) > 0 {
		element := queue[0]
		queue = queue[1:]
		theState.handleElement(element)
		queue = append(queue, theState.findNextBaseElements(element.GetOutgoing())...)
	}

	return nil
}

func (state *BpmnEngineNamedResourceState) handleElement(element BPMN20.BaseElement) {
	id := element.GetId()
	if nil != state.handlers && nil != state.handlers[id] {
		state.handlers[id](id)
	}
}

func (state *BpmnEngineState) LoadFromFile(filename, resourceName string) (*zeebe.WorkflowMetadata, error) {
	xmldata, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return state.LoadFromBytes(xmldata, resourceName)
}

func (state *BpmnEngineState) LoadFromBytes(xmldata []byte, resourceName string) (*zeebe.WorkflowMetadata, error) {
	md5sum := md5.Sum(xmldata)
	var definitions BPMN20.TDefinitions
	err := xml.Unmarshal(xmldata, &definitions)
	if err != nil {
		return nil, err
	}

	var theState *BpmnEngineNamedResourceState

	theState, ok := state.states[resourceName]
	if !ok {
		theState = NewNamedResourceState()
	}

	theState.definitions = definitions

	metadata := zeebe.WorkflowMetadata{
		VariableContext: map[string]interface{}{},
		Version: 1,
	}
	for _, process := range theState.processes {
		if process.BpmnProcessId == definitions.Process.Id {
			if areEqual(process.ChecksumBytes, md5sum) {
				return &process, nil
			} else {
				metadata.Version = process.Version + 1
			}
		}
	}
	metadata.ResourceName = resourceName
	metadata.BpmnProcessId = definitions.Process.Id
	metadata.ProcessKey = time.Now().UnixNano() << 1
	metadata.ChecksumBytes = md5sum
	theState.processes = append(theState.processes, metadata)

	state.states[resourceName] = theState
	return &metadata, nil
}

func (state *BpmnEngineNamedResourceState) findNextBaseElements(refIds []string) []BPMN20.BaseElement {
	targetRefs := make([]string, 0)
	for _, id := range refIds {
		withId := func(s string) bool { return s == id }
		targetRefs = append(targetRefs, BPMN20.FindTargetRefs(state.definitions.Process.SequenceFlows, withId)...)
	}

	elements := make([]BPMN20.BaseElement, 0)
	for _, targetRef := range targetRefs {
		elements = append(elements, state.findBaseElementsById(targetRef)...)
	}
	return elements
}

func (state *BpmnEngineNamedResourceState) findBaseElementsById(id string) (elements []BPMN20.BaseElement) {
	// todo refactor into foundation package
	// todo find smarter solution
	for _, task := range state.definitions.Process.ServiceTasks {
		if task.Id == id {
			elements = append(elements, task)
		}
	}
	// todo find smarter solution
	for _, endEvent := range state.definitions.Process.EndEvents {
		if endEvent.Id == id {
			elements = append(elements, endEvent)
		}
	}
	return elements
}

func (state *BpmnEngineState) AddTaskHandler(resourceName string, taskId string, handler func(id string)) {
	theState, ok := state.states[resourceName]
	if !ok {
		theState = NewNamedResourceState()
	}

	if nil == theState.handlers {
		theState.handlers = make(map[string]func(id string))
	}
	theState.handlers[taskId] = handler

	state.states[resourceName] = theState
}
