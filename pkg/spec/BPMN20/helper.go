package BPMN20

import (
	"html"
	"strings"
)

func FindSequenceFlows(sequenceFlows *[]TSequenceFlow, ids []string) (ret []TSequenceFlow) {
	for _, flow := range *sequenceFlows {
		for _, id := range ids {
			if id == flow.Id {
				ret = append(ret, flow)
			}
		}
	}
	return ret
}

// FindFirstSequenceFlow returns the first flow definition for any given source and target element ID
func FindFirstSequenceFlow(sequenceFlows *[]TSequenceFlow, sourceId string, targetId string) (result *TSequenceFlow) {
	for _, flow := range *sequenceFlows {
		if flow.SourceRef == sourceId && flow.TargetRef == targetId {
			result = &flow
			break
		}
	}
	return result
}

func FindBaseElementsById(processElement ProcessElement, id string) (elements []*BaseElement) {
	appendWhenIdMatches := func(element *BaseElement) {
		if (*element).GetId() == id {
			elements = append(elements, element)
		}
	}

	var be BaseElement = processElement
	appendWhenIdMatches(&be)

	for _, startEvent := range processElement.GetStartEvents() {
		var be BaseElement = startEvent
		appendWhenIdMatches(&be)
	}
	for _, endEvent := range processElement.GetEndEvents() {
		var be BaseElement = endEvent
		appendWhenIdMatches(&be)
	}
	for _, task := range processElement.GetServiceTasks() {
		var be BaseElement = task
		appendWhenIdMatches(&be)
	}
	for _, task := range processElement.GetUserTasks() {
		var be BaseElement = task
		appendWhenIdMatches(&be)
	}
	for _, parallelGateway := range processElement.GetParallelGateway() {
		var be BaseElement = parallelGateway
		appendWhenIdMatches(&be)
	}
	for _, exclusiveGateway := range processElement.GetExclusiveGateway() {
		var be BaseElement = exclusiveGateway
		appendWhenIdMatches(&be)
	}
	for _, eventBasedGateway := range processElement.GetEventBasedGateway() {
		var be BaseElement = eventBasedGateway
		appendWhenIdMatches(&be)
	}
	for _, intermediateCatchEvent := range processElement.GetIntermediateCatchEvent() {
		var be BaseElement = intermediateCatchEvent
		appendWhenIdMatches(&be)
	}
	for _, intermediateCatchEvent := range processElement.GetIntermediateTrowEvent() {
		var be BaseElement = intermediateCatchEvent
		appendWhenIdMatches(&be)
	}
	for _, inclusiveGateway := range processElement.GetInclusiveGateway() {
		var be BaseElement = inclusiveGateway
		appendWhenIdMatches(&be)
	}
	for _, subProcess := range processElement.GetSubProcess() {
		var be BaseElement = subProcess
		appendWhenIdMatches(&be)
		// search recursively for further elements
		elements = append(elements, FindBaseElementsById(subProcess, id)...)
	}
	return elements
}

// HasConditionExpression returns true, if there's exactly 1 expression present (as by the spec)
// and there's some non-whitespace-characters available
func (flow TSequenceFlow) HasConditionExpression() bool {
	return len(flow.ConditionExpression) == 1 && len(strings.TrimSpace(flow.GetConditionExpression())) > 0
}

// GetConditionExpression returns the embedded expression. There will be a panic thrown, in case none exists!
func (flow TSequenceFlow) GetConditionExpression() string {
	return html.UnescapeString(flow.ConditionExpression[0].Text)
}
