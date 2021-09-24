package engine

import (
	"encoding/xml"
	"github.com/nitram509/golib-bpmn-model/pgk/spec/BPMN20"
	"io/ioutil"
)

type Node struct {
	Name string
	Id   string
}

type registeredProcess struct {
	processInfo WorkflowMetadata
}

type BpmnEngine interface {
	LoadFromDefinitions(definitions BPMN20.TDefinitions)
}

type BpmnEngineState struct {
	registry    []registeredProcess
	definitions BPMN20.TDefinitions
	queue       []BPMN20.BaseElement
	handlers    map[string]func(id string)
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

func (state *BpmnEngineState) LoadFromFile(filename string) error {
	xmldata, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var definitions BPMN20.TDefinitions
	err = xml.Unmarshal(xmldata, &definitions)
	if err != nil {
		return err
	}
	state.definitions = definitions
	return nil
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

func (state *BpmnEngineState) AddHandler(taskId string, handler func(id string)) {
	if nil == state.handlers {
		state.handlers = make(map[string]func(id string))
	}
	state.handlers[taskId] = handler
}
