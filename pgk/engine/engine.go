package engine

import (
	"encoding/xml"
	"fmt"
	"github.com/nitram509/golib-bpmn-model/pgk/spec/BPMN/20100501/BPMN20"
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
}

func (state *BpmnEngineState) Execute() {
	queue := make([]BPMN20.BaseElement, 0)
	for _, event := range state.definitions.Process.StartEvents {
		element := BPMN20.BaseElement{}
		element.Id = event.Id
		element.Outgoing = event.OutgoingAssociation
		queue = append(queue, element)
	}
	state.queue = queue

	for len(queue) > 0 {
		element := queue[0]
		println(element.Id)
		queue = queue[1:]
		queue = append(queue, state.findNextBaseElements(element.Outgoing)...)
	}
}

func (state *BpmnEngineState) LoadFromFile(filename string) {
	xmldata, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Print(err)
		return
	}

	var definitions BPMN20.TDefinitions
	err = xml.Unmarshal(xmldata, &definitions)
	state.definitions = definitions
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
			baseElement := BPMN20.BaseElement{}
			baseElement.Id = task.Id
			baseElement.Incoming = task.IncomingAssociation
			baseElement.Outgoing = task.OutgoingAssociation
			//baseElement.Type = BPMN20.ServiceTaskType
			elements = append(elements, baseElement)
		}
	}
	// todo find smarter solution
	for _, endEvent := range state.definitions.Process.EndEvents {
		if endEvent.Id == id {
			baseElement := BPMN20.BaseElement{}
			baseElement.Id = endEvent.Id
			baseElement.Incoming = endEvent.IncomingAssociation
			baseElement.Outgoing = endEvent.OutgoingAssociation
			elements = append(elements, baseElement)
		}
	}
	return elements
}
