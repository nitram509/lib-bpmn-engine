package bpmn_engine

import (
	"crypto/md5"
	"encoding/xml"
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

type BpmnEngineState struct {
	processes        []ProcessInfo
	processInstances []InstanceInfo
	definitions      BPMN20.TDefinitions
	queue            []BPMN20.BaseElement
	handlers         map[string]func(id string)
}

func New() BpmnEngineState {
	return BpmnEngineState{
		processes:        []ProcessInfo{},
		processInstances: []InstanceInfo{},
		queue:            []BPMN20.BaseElement{},
		handlers:         map[string]func(id string){},
	}
}

// GetProcessInstances returns an ordered list instance information.
func (state *BpmnEngineState) GetProcessInstances(resourceName string) []InstanceInfo {
	return state.processInstances
}

func (state *BpmnEngineState) CreateInstance(resourceName string) (InstanceInfo, error) {
	info := InstanceInfo{
		ProcessInfo:     &ProcessInfo{}, // TODO: link the actual process
		InstanceKey:     time.Now().UnixNano() << 1,
		VariableContext: map[string]interface{}{},
	}
	state.processInstances = append(state.processInstances, info)
	return info, nil
}

func (state *BpmnEngineState) CreateAndRunInstance(resourceName string) error {
	_, err := state.CreateInstance(resourceName)
	if err != nil {
		return err
	}

	queue := make([]BPMN20.BaseElement, 0)
	for _, event := range state.definitions.Process.StartEvents {
		queue = append(queue, event)
	}
	state.queue = queue

	for len(queue) > 0 {
		element := queue[0]
		queue = queue[1:]
		state.handleElement(element)
		queue = append(queue, state.findNextBaseElements(element.GetOutgoing())...)
	}

	return nil
}

func (state *BpmnEngineState) handleElement(element BPMN20.BaseElement) {
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

	state.definitions = definitions

	processInfo := ProcessInfo{
		Version: 1,
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
	processInfo.ResourceName = resourceName
	processInfo.BpmnProcessId = definitions.Process.Id
	processInfo.ProcessKey = time.Now().UnixNano() << 1
	processInfo.checksumBytes = md5sum
	state.processes = append(state.processes, processInfo)

	return &processInfo, nil
}

func (state *BpmnEngineState) findNextBaseElements(refIds []string) []BPMN20.BaseElement {
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

func (state *BpmnEngineState) findBaseElementsById(id string) (elements []BPMN20.BaseElement) {
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

func (state *BpmnEngineState) AddTaskHandler(taskId string, handler func(id string)) {
	if nil == state.handlers {
		state.handlers = make(map[string]func(id string))
	}
	state.handlers[taskId] = handler
}
