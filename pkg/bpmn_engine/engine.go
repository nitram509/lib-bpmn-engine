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
		name:              name,
		processes:         []ProcessInfo{},
		processInstances:  []*ProcessInstanceInfo{},
		handlers:          map[string]func(context ProcessInstanceContext){},
		activationCounter: map[string]int64{},
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
				instanceKey:     time.Now().UnixNano() << 1,
				variableContext: variableContext,
				createdAt:       time.Now(),
				state:           BPMN20.ProcessInstanceReady,
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
	queue := make([]BPMN20.BaseElement, 0)
	process := instance.processInfo

	switch instance.state {
	case BPMN20.ProcessInstanceReady:
		for _, event := range process.definitions.Process.StartEvents {
			queue = append(queue, event)
		}
		instance.state = BPMN20.ProcessInstanceActive
	case BPMN20.ProcessInstanceActive:
		for _, event := range instance.caughtEvents {
			for _, ice := range process.definitions.Process.IntermediateCatchEvent {
				if event.Name == ice.Name {
					queue = append(queue, ice)
				}
			}
		}
	case BPMN20.ProcessInstanceCompleted:
		return nil
	default:
		panic("Unknown process instance state.")
	}

	for len(queue) > 0 {
		element := queue[0]
		queue = queue[1:]

		counter, _ := state.activationCounter[element.GetId()]
		if element.GetTypeName() == BPMN20.ParallelGatewayType && counter == 1 {
			// do nothing, because after a parallel join, the execution is just once
		} else {
			state.activationCounter[element.GetId()] = counter + 1
			continueNextElement := state.handleElement(element, process, instance)
			if continueNextElement {
				queue = append(queue, state.findNextBaseElements(process, element.GetOutgoing())...)
			}
		}
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

func (state *BpmnEngineState) handleElement(element BPMN20.BaseElement, process *ProcessInfo, instance *ProcessInstanceInfo) bool {
	id := element.GetId()
	switch element.GetTypeName() {
	case BPMN20.ServiceTaskType:
		state.handleServiceTask(id, process, instance)
	case BPMN20.EndEventType:
		instance.state = BPMN20.ProcessInstanceCompleted
	case BPMN20.IntermediateCatchEventType:
		return state.handleIntermediateCatchEvent(id, element.GetName(), instance)
	default:
		// TODO: somehow complain, that this is an unsupported element
	}
	return true
}

func (state *BpmnEngineState) handleServiceTask(id string, process *ProcessInfo, instance *ProcessInstanceInfo) {
	if nil != state.handlers && nil != state.handlers[id] {
		data := ProcessInstanceContextData{
			taskId:       id,
			processInfo:  process,
			instanceInfo: instance,
		}
		state.handlers[id](&data)
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
