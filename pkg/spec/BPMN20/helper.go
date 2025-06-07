package BPMN20

import (
	"html"
	"strings"
)

// FindSequenceFlows returns all TSequenceFlow with any given `id`
func FindSequenceFlows(processElement ProcessElement, ids []string) (ret []TSequenceFlow) {
	for _, flow := range processElement.GetSequenceFlows() {
		for _, id := range ids {
			if id == flow.Id {
				ret = append(ret, flow)
			}
		}
	}
	for _, subprocess := range processElement.GetSubProcess() {
		ret = append(ret, FindSequenceFlows(subprocess, ids)...)
	}
	return ret
}

// FindFirstSequenceFlow returns the first flow definition for any given source and target element ID
func FindFirstSequenceFlow(processElement ProcessElement, sourceId string, targetId string) (result *TSequenceFlow) {
	for _, flow := range processElement.GetSequenceFlows() {
		if flow.SourceRef == sourceId && flow.TargetRef == targetId {
			result = &flow
			return result
		}
	}
	for _, subProcess := range processElement.GetSubProcess() {
		result = FindFirstSequenceFlow(subProcess, sourceId, targetId)
		if result != nil {
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

	appendWhenIdMatches(Ptr[BaseElement](processElement))

	for _, startEvent := range processElement.GetStartEvents() {
		appendWhenIdMatches(Ptr[BaseElement](startEvent))
	}
	for _, endEvent := range processElement.GetEndEvents() {
		appendWhenIdMatches(Ptr[BaseElement](endEvent))
	}
	for _, task := range processElement.GetServiceTasks() {
		appendWhenIdMatches(Ptr[BaseElement](task))
	}
	for _, task := range processElement.GetUserTasks() {
		appendWhenIdMatches(Ptr[BaseElement](task))
	}
	for _, parallelGateway := range processElement.GetParallelGateway() {
		appendWhenIdMatches(Ptr[BaseElement](parallelGateway))
	}
	for _, exclusiveGateway := range processElement.GetExclusiveGateway() {
		appendWhenIdMatches(Ptr[BaseElement](exclusiveGateway))
	}
	for _, eventBasedGateway := range processElement.GetEventBasedGateway() {
		appendWhenIdMatches(Ptr[BaseElement](eventBasedGateway))
	}
	for _, intermediateCatchEvent := range processElement.GetIntermediateCatchEvent() {
		appendWhenIdMatches(Ptr[BaseElement](intermediateCatchEvent))
	}
	for _, intermediateCatchEvent := range processElement.GetIntermediateTrowEvent() {
		appendWhenIdMatches(Ptr[BaseElement](intermediateCatchEvent))
	}
	for _, inclusiveGateway := range processElement.GetInclusiveGateway() {
		appendWhenIdMatches(Ptr[BaseElement](inclusiveGateway))
	}
	for _, boundaryEvent := range processElement.GetBoundaryEvent() {
		appendWhenIdMatches(Ptr[BaseElement](boundaryEvent))
	}
	for _, subProcess := range processElement.GetSubProcess() {
		appendWhenIdMatches(Ptr[BaseElement](subProcess))
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

func Ptr[T any](v T) *T {
	return &v
}
