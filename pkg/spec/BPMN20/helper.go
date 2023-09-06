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

// FindSequenceFlow TODO: warning: one can define multiple flows from one element to another
func FindSequenceFlow(sequenceFlows *[]TSequenceFlow, sourceId string, targetId string) (result *TSequenceFlow) {
	for _, flow := range *sequenceFlows {
		if flow.SourceRef == sourceId && flow.TargetRef == targetId {
			result = &flow
			break
		}
	}
	return result
}

func FindBaseElementsById(definitions *TDefinitions, id string) (elements []*BaseElement) {
	appender := func(element *BaseElement) {
		if (*element).GetId() == id {
			elements = append(elements, element)
		}
	}
	for _, startEvent := range definitions.Process.StartEvents {
		var be BaseElement = startEvent
		appender(&be)
	}
	for _, endEvent := range definitions.Process.EndEvents {
		var be BaseElement = endEvent
		appender(&be)
	}
	for _, task := range definitions.Process.ServiceTasks {
		var be BaseElement = task
		appender(&be)
	}
	for _, task := range definitions.Process.UserTasks {
		var be BaseElement = task
		appender(&be)
	}
	for _, parallelGateway := range definitions.Process.ParallelGateway {
		var be BaseElement = parallelGateway
		appender(&be)
	}
	for _, exclusiveGateway := range definitions.Process.ExclusiveGateway {
		var be BaseElement = exclusiveGateway
		appender(&be)
	}
	for _, eventBasedGateway := range definitions.Process.EventBasedGateway {
		var be BaseElement = eventBasedGateway
		appender(&be)
	}
	for _, intermediateCatchEvent := range definitions.Process.IntermediateCatchEvent {
		var be BaseElement = intermediateCatchEvent
		appender(&be)
	}
	for _, intermediateCatchEvent := range definitions.Process.IntermediateTrowEvent {
		var be BaseElement = intermediateCatchEvent
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
