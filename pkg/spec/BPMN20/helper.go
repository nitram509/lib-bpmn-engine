package BPMN20

import (
	"html"
	"strings"
)

func _FindSequenceFlows(sequenceFlows *[]TSequenceFlow, ids []string) (ret []TSequenceFlow) {
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
func _FindFirstSequenceFlow(sequenceFlows *[]TSequenceFlow, sourceId string, targetId string) (result *TSequenceFlow) {
	for _, flow := range *sequenceFlows {
		if flow.SourceRef == sourceId && flow.TargetRef == targetId {
			result = &flow
			break
		}
	}
	return result
}

func _FindBaseElementsById(process ProcessElement, id string) (elements []*BaseElement) {
	appender := func(element *BaseElement) {
		if (*element).GetId() == id {
			elements = append(elements, element)
		}
	}
	for _, startEvent := range process.GetStartEvents() {
		var be BaseElement = startEvent
		appender(&be)
	}
	for _, endEvent := range process.GetEndEvents() {
		var be BaseElement = endEvent
		appender(&be)
	}
	for _, task := range process.GetServiceTasks() {
		var be BaseElement = task
		appender(&be)
	}
	for _, task := range process.GetUserTasks() {
		var be BaseElement = task
		appender(&be)
	}
	for _, parallelGateway := range process.GetParallelGateway() {
		var be BaseElement = parallelGateway
		appender(&be)
	}
	for _, exclusiveGateway := range process.GetExclusiveGateway() {
		var be BaseElement = exclusiveGateway
		appender(&be)
	}
	for _, eventBasedGateway := range process.GetEventBasedGateway() {
		var be BaseElement = eventBasedGateway
		appender(&be)
	}
	for _, intermediateCatchEvent := range process.GetIntermediateCatchEvent() {
		var be BaseElement = intermediateCatchEvent
		appender(&be)
	}
	for _, intermediateCatchEvent := range process.GetIntermediateTrowEvent() {
		var be BaseElement = intermediateCatchEvent
		appender(&be)
	}
	for _, inclusiveGateway := range process.GetInclusiveGateway() {
		var be BaseElement = inclusiveGateway
		appender(&be)
	}
	for _, subProcess := range process.GetSubProcess() {
		var be BaseElement = subProcess
		appender(&be)
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
