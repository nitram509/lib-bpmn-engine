package bpmn_engine

import (
	"crypto/md5"
	"encoding/xml"
	"errors"
	"fmt"
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
	BpmnProcessId string // The ID as defined in the BPMN file
	Version       int32  // A version of the process, default=1, incremented, when another process with the same ID is loaded
	ProcessKey    int64  // The engines key for this given process with version

	definitions   BPMN20.TDefinitions // parsed file content
	checksumBytes [16]byte            // internal checksum to identify different versions
}

type InstanceInfo struct {
	processInfo     *ProcessInfo
	InstanceKey     int64
	VariableContext map[string]interface{}
	createdAt       time.Time
}

type BpmnEngineState struct {
	name             string
	processes        []ProcessInfo
	processInstances []InstanceInfo
	queue            []BPMN20.BaseElement
	handlers         map[string]func(id string)
}

// New creates an engine with an arbitrary name of the engine;
// useful in case you have multiple ones
func New(name string) BpmnEngineState {
	return BpmnEngineState{
		name:             name,
		processes:        []ProcessInfo{},
		processInstances: []InstanceInfo{},
		queue:            []BPMN20.BaseElement{},
		handlers:         map[string]func(id string){},
	}
}

// GetProcessInstances returns an ordered list instance information.
func (state *BpmnEngineState) GetProcessInstances() []InstanceInfo {
	return state.processInstances
}

func (state *BpmnEngineState) GetName() string {
	return state.name
}

func (state *BpmnEngineState) CreateInstance(processKey int64) (*InstanceInfo, error) {
	for _, process := range state.processes {
		if process.ProcessKey == processKey {
			info := InstanceInfo{
				processInfo:     &process,
				InstanceKey:     time.Now().UnixNano() << 1,
				VariableContext: map[string]interface{}{},
				createdAt:       time.Now(),
			}
			state.processInstances = append(state.processInstances, info)
			return &info, nil
		}
	}
	return nil, nil
}

func (state *BpmnEngineState) CreateAndRunInstance(processKey int64) error {
	instance, err := state.CreateInstance(processKey)
	if err != nil {
		return err
	}
	if instance == nil {
		return errors.New(fmt.Sprint("can't find process with processKey=", processKey, "."))
	}

	process := instance.processInfo
	queue := make([]BPMN20.BaseElement, 0)
	for _, event := range process.definitions.Process.StartEvents {
		queue = append(queue, event)
	}
	state.queue = queue

	for len(queue) > 0 {
		element := queue[0]
		queue = queue[1:]
		state.handleElement(element)
		queue = append(queue, state.findNextBaseElements(process, element.GetOutgoing())...)
	}

	return nil
}

func (state *BpmnEngineState) handleElement(element BPMN20.BaseElement) {
	id := element.GetId()
	if nil != state.handlers && nil != state.handlers[id] {
		state.handlers[id](id)
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
	processInfo.ProcessKey = time.Now().UnixNano() << 1
	processInfo.checksumBytes = md5sum
	state.processes = append(state.processes, processInfo)

	return &processInfo, nil
}

func (state *BpmnEngineState) AddTaskHandler(taskId string, handler func(id string)) {
	if nil == state.handlers {
		state.handlers = make(map[string]func(id string))
	}
	state.handlers[taskId] = handler
}
