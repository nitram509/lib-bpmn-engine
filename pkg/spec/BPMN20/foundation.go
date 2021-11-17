package BPMN20

type BaseElement interface {
	GetId() string
	GetName() string
	GetIncoming() []string
	GetOutgoing() []string
}

func (startEvent TStartEvent) GetId() string {
	return startEvent.Id
}

func (startEvent TStartEvent) GetName() string {
	return startEvent.Name
}

func (startEvent TStartEvent) GetIncoming() []string {
	return startEvent.IncomingAssociation
}

func (startEvent TStartEvent) GetOutgoing() []string {
	return startEvent.OutgoingAssociation
}

func (endEvent TEndEvent) GetId() string {
	return endEvent.Id
}

func (endEvent TEndEvent) GetName() string {
	return endEvent.Name
}

func (endEvent TEndEvent) GetIncoming() []string {
	return endEvent.IncomingAssociation
}

func (endEvent TEndEvent) GetOutgoing() []string {
	return endEvent.OutgoingAssociation
}

func (serviceTask TServiceTask) GetId() string {
	return serviceTask.Id
}

func (serviceTask TServiceTask) GetName() string {
	return serviceTask.Name
}

func (serviceTask TServiceTask) GetIncoming() []string {
	return serviceTask.IncomingAssociation
}

func (serviceTask TServiceTask) GetOutgoing() []string {
	return serviceTask.OutgoingAssociation
}

func (parallelGateway TParallelGateway) GetId() string {
	return parallelGateway.Id
}

func (parallelGateway TParallelGateway) GetName() string {
	return parallelGateway.Name
}

func (parallelGateway TParallelGateway) GetIncoming() []string {
	return parallelGateway.IncomingAssociation
}

func (parallelGateway TParallelGateway) GetOutgoing() []string {
	return parallelGateway.OutgoingAssociation
}

//type BaseElementType int8
//
//const (
//	NotYetSupportedType BaseElementType = 0
//	ServiceTaskType     BaseElementType = 1
//)

func FindTargetRefs(sequenceFlows []TSequenceFlow, withId func(string) bool) (ret []string) {
	for _, flow := range sequenceFlows {
		if withId(flow.Id) {
			ret = append(ret, flow.TargetRef)
		}
	}
	return
}
