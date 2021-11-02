package bpmn_engine

import (
	"crypto/md5"
	"encoding/xml"
	"errors"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"io/ioutil"
	"time"
)

type Node struct {
	Name string
	Id   string
}

type BpmnEngine interface {
	LoadFromFile(filename string) (InstanceInfo, error)
	LoadFromBytes(xmlData []byte, resourceName string) (InstanceInfo, error)
	AddTaskHandler(taskId string, handler func(id string))
	GetProcesses() []InstanceInfo
}

type ProcessInfo struct {
	BpmnProcessId string   // The ID as defined in the BPMN file
	Version       int32    // A version of the process, default=1, incremented, when another process with the same ID is loaded
	ProcessKey    int64    // The engines key for this given process with version
	ResourceName  string   // Just for information, the provided resource name
	checksumBytes [16]byte // internal checksum to identify different versions
}

type InstanceInfo struct {
	ProcessInfo     *ProcessInfo
	InstanceKey     int64
	VariableContext map[string]interface{}
}

func NewNamedResourceState() *BpmnEngineNamedResourceState {
	return &BpmnEngineNamedResourceState{}
}

type BpmnEngineNamedResourceState struct {
	processes        []ProcessInfo
	processInstances []InstanceInfo
	definitions      BPMN20.TDefinitions
	queue            []BPMN20.BaseElement
	handlers         map[string]func(id string)
}

type BpmnEngineState struct {
	states map[string]*BpmnEngineNamedResourceState
}

func New() BpmnEngineState {
	return BpmnEngineState{
		states: map[string]*BpmnEngineNamedResourceState{},
	}
}

// GetProcessInstances returns an ordered list instance information.
func (state *BpmnEngineState) GetProcessInstances(resourceName string) []InstanceInfo {
	value, ok := state.states[resourceName]
	if !ok {
		return []InstanceInfo{}
	}

	return value.processInstances
}

func (state *BpmnEngineState) CreateInstance(resourceName string) (InstanceInfo, error) {
	theState, ok := state.states[resourceName]
	if !ok {
		return InstanceInfo{}, errors.New("resource name not found")
	}

	info := InstanceInfo{
		ProcessInfo:     &ProcessInfo{}, // TODO: link the actual process
		InstanceKey:     time.Now().UnixNano() << 1,
		VariableContext: map[string]interface{}{},
	}

	theState.processInstances = append(theState.processInstances, info)
	return info, nil
}

func (state *BpmnEngineState) CreateAndRunInstance(resourceName string) error {
	_, err := state.CreateInstance(resourceName)
	if err != nil {
		return err
	}

	// TODO: remove this, in favor of sing the instance information from above
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

func (state *BpmnEngineState) LoadFromFile(filename, resourceName string) (*ProcessInfo, error) {
	xmlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return state.LoadFromBytes(xmlData, resourceName)
}

func (state *BpmnEngineState) LoadFromBytes(xmlData []byte, resourceName string) (*ProcessInfo, error) {
	md5sum := md5.Sum(xmlData)
	var definitions BPMN20.TDefinitions
	err := xml.Unmarshal(xmlData, &definitions)
	if err != nil {
		return nil, err
	}

	var theState *BpmnEngineNamedResourceState

	theState, ok := state.states[resourceName]
	if !ok {
		theState = NewNamedResourceState()
	}

	theState.definitions = definitions

	processInfo := ProcessInfo{
		Version: 1,
	}
	for _, process := range theState.processes {
		if process.BpmnProcessId == definitions.Process.Id {
			if areEqual(process.checksumBytes, md5sum) {
				return &process, nil
			} else {
				processInfo.Version = process.Version + 1
			}
		}
	}
	processInfo.ResourceName = resourceName
	processInfo.BpmnProcessId = definitions.Process.Id
	processInfo.ProcessKey = time.Now().UnixNano() << 1
	processInfo.checksumBytes = md5sum
	theState.processes = append(theState.processes, processInfo)

	state.states[resourceName] = theState
	return &processInfo, nil
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
