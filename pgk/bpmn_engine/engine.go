package bpmn_engine

import (
	"crypto/md5"
	"encoding/xml"
	"github.com/nitram509/golib-bpmn-model/pgk/bpmn_engine/zeebe"
	"github.com/nitram509/golib-bpmn-model/pgk/spec/BPMN20"
	"io/ioutil"
	"time"
)

type Node struct {
	Name string
	Id   string
}

type registeredProcess struct {
	workflowMetadata zeebe.WorkflowMetadata
}

type BpmnEngine interface {
	LoadFromFile(filename string) (zeebe.WorkflowMetadata, error)
	GetProcesses() []zeebe.WorkflowMetadata
}

func New() BpmnEngineState {
	return BpmnEngineState{}
}

type BpmnEngineState struct {
	processes   []zeebe.WorkflowMetadata
	definitions BPMN20.TDefinitions
	queue       []BPMN20.BaseElement
	handlers    map[string]func(id string)
}

func (state *BpmnEngineState) GetProcesses() []zeebe.WorkflowMetadata {
	return state.processes
}

func (state *BpmnEngineState) Execute() {
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
}

func (state *BpmnEngineState) handleElement(element BPMN20.BaseElement) {
	id := element.GetId()
	if nil != state.handlers && nil != state.handlers[id] {
		state.handlers[id](id)
	}
}

func (state *BpmnEngineState) LoadFromFile(filename string) (zeebe.WorkflowMetadata, error) {
	xmldata, err := ioutil.ReadFile(filename)
	md5sum := md5.Sum(xmldata)
	if err != nil {
		return zeebe.WorkflowMetadata{}, err
	}

	var definitions BPMN20.TDefinitions
	err = xml.Unmarshal(xmldata, &definitions)
	if err != nil {
		return zeebe.WorkflowMetadata{}, err
	}
	state.definitions = definitions

	metadata := zeebe.WorkflowMetadata{Version: 1}
	for _, process := range state.processes {
		if process.BpmnProcessId == definitions.Process.Id {
			if areEqual(process.Md5sum, md5sum) {
				return process, nil
			} else {
				metadata.Version = process.Version + 1
			}
		}
	}
	metadata.ResourceName = filename
	metadata.BpmnProcessId = definitions.Process.Id
	metadata.ProcessKey = time.Now().UnixNano() << 1
	metadata.Md5sum = md5sum
	state.processes = append(state.processes, metadata)
	return metadata, nil
}

func areEqual(a [16]byte, b [16]byte) bool {
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
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
