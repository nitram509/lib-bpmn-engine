package BPMN20

type BaseElement interface {
	GetId() string
	GetIncoming() []string
	GetOutgoing() []string
	//Type     BaseElementType
}

func (startEvent TStartEvent) GetId() string {
	return startEvent.Id
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

func (endEvent TEndEvent) GetIncoming() []string {
	return endEvent.IncomingAssociation
}

func (endEvent TEndEvent) GetOutgoing() []string {
	return endEvent.OutgoingAssociation
}

func (serviceTask TServiceTask) GetId() string {
	return serviceTask.Id
}

func (serviceTask TServiceTask) GetIncoming() []string {
	return serviceTask.IncomingAssociation
}

func (serviceTask TServiceTask) GetOutgoing() []string {
	return serviceTask.OutgoingAssociation
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
