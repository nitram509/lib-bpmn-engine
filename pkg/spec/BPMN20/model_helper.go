package BPMN20

func FindTargetRefs(sequenceFlows []TSequenceFlow, id string) (ret []string) {
	for _, flow := range sequenceFlows {
		if id == flow.Id {
			ret = append(ret, flow.TargetRef)
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
