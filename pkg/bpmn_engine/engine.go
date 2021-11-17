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

type BpmnEngine interface {
	LoadFromFile(filename string) (*ProcessInfo, error)
	LoadFromBytes(xmlData []byte) (*ProcessInfo, error)
	AddTaskHandler(taskType string, handler func(context ProcessInstanceContext))
	CreateInstance(processKey int64, variableContext map[string]string) (*InstanceInfo, error)
	CreateAndRunInstance(processKey int64, variableContext map[string]string) error
	GetName() string
	GetProcessInstances() []InstanceInfo
}

// New creates an engine with an arbitrary name of the engine;
// useful in case you have multiple ones
func New(name string) BpmnEngineState {
	return BpmnEngineState{
		name:             name,
		processes:        []ProcessInfo{},
		processInstances: []InstanceInfo{},
		queue:            []BPMN20.BaseElement{},
		handlers:         map[string]func(context ProcessInstanceContext){},
	}
}

// CreateInstance creates a new instance for a process with given processKey
func (state *BpmnEngineState) CreateInstance(processKey int64, variableContext map[string]string) (*InstanceInfo, error) {
	if variableContext == nil {
		variableContext = map[string]string{}
	}
	for _, process := range state.processes {
		if process.ProcessKey == processKey {
			info := InstanceInfo{
				processInfo:     &process,
				InstanceKey:     time.Now().UnixNano() << 1,
				VariableContext: variableContext,
				createdAt:       time.Now(),
			}
			state.processInstances = append(state.processInstances, info)
			return &info, nil
		}
	}
	return nil, nil
}

// CreateAndRunInstance creates a new instance and executes it immediately.
// The provided variableContext can be nil
func (state *BpmnEngineState) CreateAndRunInstance(processKey int64, variableContext map[string]string) error {
	instance, err := state.CreateInstance(processKey, variableContext)
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
		state.handleElement(element, process, instance)
		queue = append(queue, state.findNextBaseElements(process, element.GetOutgoing())...)
	}

	return nil
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

// AddTaskHandler registers a handler for a given taskType
func (state *BpmnEngineState) AddTaskHandler(taskId string, handler func(context ProcessInstanceContext)) {
	if nil == state.handlers {
		state.handlers = make(map[string]func(context ProcessInstanceContext))
	}
	state.handlers[taskId] = handler
}

func (state *BpmnEngineState) handleElement(element BPMN20.BaseElement, process *ProcessInfo, instance *InstanceInfo) {
	id := element.GetId()
	if nil != state.handlers && nil != state.handlers[id] {
		data := ProcessInstanceContextData{
			taskId:       id,
			processInfo:  process,
			instanceInfo: instance,
		}
		state.handlers[id](&data)
	}
}

type ProcessInstanceContextData struct {
	taskId       string
	processInfo  *ProcessInfo
	instanceInfo *InstanceInfo
}

func (data *ProcessInstanceContextData) GetTaskId() string {
	return data.taskId
}

func (data *ProcessInstanceContextData) GetVariable(name string) string {
	return data.instanceInfo.VariableContext[name]
}

func (data *ProcessInstanceContextData) SetVariable(name string, value string) {
	data.instanceInfo.VariableContext[name] = value
}
