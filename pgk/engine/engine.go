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

//func (state BpmnEngineState) LoadFromDefinitions(definitions BPMN20.TDefinitions) (DeployWorkflowResponse, error) {
//	info := WorkflowMetadata{
//		bpmnProcessId: "123",
//		version:       1,
//		processKey:    456,
//		resourceName:  "xxx",
//	}
//
//	state.registry = append(state.registry, registeredProcess{info})
//
//	result := DeployWorkflowResponse{}
//	result.key = "1234567890"
//	result.processes = append(result.processes, info)
//	return result, nil
//}

// see https://github.com/camunda-cloud/zeebe/blob/0.13.1/gateway-protocol/src/main/proto/gateway.proto
type DeployWorkflowResponse struct {
	key       string
	processes []WorkflowMetadata
}

// see https://github.com/camunda-cloud/zeebe/blob/0.13.1/gateway-protocol/src/main/proto/gateway.proto
type WorkflowMetadata struct {
	bpmnProcessId string
	version       int32
	processKey    int64
	resourceName  string
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
		queue = append(queue, state.findNextElements(element.Outgoing)...)
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

func findTargetRefs(sequenceFlows []BPMN20.TSequenceFlow, withId func(string) bool) (ret []string) {
	for _, flow := range sequenceFlows {
		if withId(flow.Id) {
			ret = append(ret, flow.TargetRef)
		}
	}
	return
}

func (state *BpmnEngineState) findNextElements(refIds []string) []BPMN20.BaseElement {
	targetRefs := make([]string, 0)
	for _, id := range refIds {
		withId := func(s string) bool { return s == id }
		targetRefs = append(targetRefs, findTargetRefs(state.definitions.Process.SequenceFlows, withId)...)
	}

	elements := make([]BPMN20.BaseElement, 0)
	for _, targetRef := range targetRefs {
		// todo find smarter solution
		for _, task := range state.definitions.Process.ServiceTasks {
			if task.Id == targetRef {
				baseElement := BPMN20.BaseElement{}
				baseElement.Id = task.Id
				baseElement.Incoming = task.IncomingAssociation
				baseElement.Outgoing = task.OutgoingAssociation
				elements = append(elements, baseElement)
			}
		}
		// todo find smarter solution
		for _, endEvent := range state.definitions.Process.EndEvents {
			if endEvent.Id == targetRef {
				baseElement := BPMN20.BaseElement{}
				baseElement.Id = endEvent.Id
				baseElement.Incoming = endEvent.IncomingAssociation
				baseElement.Outgoing = endEvent.OutgoingAssociation
				elements = append(elements, baseElement)
			}
		}
	}
	return elements
}
