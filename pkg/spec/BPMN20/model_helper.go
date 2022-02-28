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

func FindSourceRefs(sequenceFlows []TSequenceFlow, id string) (ret []string) {
	for _, flow := range sequenceFlows {
		if id == flow.Id {
			ret = append(ret, flow.SourceRef)
		}
	}
	return
}

func FindBaseElementsById(definitions TDefinitions, id string) (elements []BaseElement) {
	appender := func(element BaseElement) {
		if element.GetId() == id {
			elements = append(elements, element)
		}
	}
	for _, task := range definitions.Process.ServiceTasks {
		appender(task)
	}
	for _, endEvent := range definitions.Process.EndEvents {
		appender(endEvent)
	}
	for _, parallelGateway := range definitions.Process.ParallelGateway {
		appender(parallelGateway)
	}
	for _, exclusiveGateway := range definitions.Process.ExclusiveGateway {
		appender(exclusiveGateway)
	}
	for _, intermediateCatchEvent := range definitions.Process.IntermediateCatchEvent {
		appender(intermediateCatchEvent)
	}
	for _, eventBasedGateway := range definitions.Process.EventBasedGateway {
		appender(eventBasedGateway)
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
