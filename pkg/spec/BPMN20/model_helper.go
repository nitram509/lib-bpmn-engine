package BPMN20

func FindTargetRefs(sequenceFlows []TSequenceFlow, withId func(string) bool) (ret []string) {
	for _, flow := range sequenceFlows {
		if withId(flow.Id) {
			ret = append(ret, flow.TargetRef)
		}
	}
	return
}

func FindSourceRefs(sequenceFlows []TSequenceFlow, withId func(string) bool) (ret []string) {
	for _, flow := range sequenceFlows {
		if withId(flow.Id) {
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
	return elements
}

func FindSourceBaseElements(definitions TDefinitions, refIds []string) []BaseElement {
	sourceRefs := make([]string, 0)
	for _, id := range refIds {
		withId := func(s string) bool { return s == id }
		sourceRefs = append(sourceRefs, FindSourceRefs(definitions.Process.SequenceFlows, withId)...)
	}

	elements := make([]BaseElement, 0)
	for _, sourceRef := range sourceRefs {
		elements = append(elements, FindBaseElementsById(definitions, sourceRef)...)
	}
	return elements
}
