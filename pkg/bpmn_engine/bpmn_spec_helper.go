package bpmn_engine

import "github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"

func (state *BpmnEngineState) findBaseElementsById(process *ProcessInfo, id string) (elements []BPMN20.BaseElement) {
	// todo refactor into foundation package
	appender := func(element BPMN20.BaseElement) {
		if element.GetId() == id {
			elements = append(elements, element)
		}
	}
	for _, task := range process.definitions.Process.ServiceTasks {
		appender(task)
	}
	for _, endEvent := range process.definitions.Process.EndEvents {
		appender(endEvent)
	}
	for _, parallelGateway := range process.definitions.Process.ParallelGateway {
		appender(parallelGateway)
	}
	return elements
}

func (state *BpmnEngineState) findNextBaseElements(process *ProcessInfo, refIds []string) []BPMN20.BaseElement {
	targetRefs := make([]string, 0)
	for _, id := range refIds {
		withId := func(s string) bool { return s == id }
		targetRefs = append(targetRefs, BPMN20.FindTargetRefs(process.definitions.Process.SequenceFlows, withId)...)
	}

	elements := make([]BPMN20.BaseElement, 0)
	for _, targetRef := range targetRefs {
		elements = append(elements, state.findBaseElementsById(process, targetRef)...)
	}
	return elements
}
