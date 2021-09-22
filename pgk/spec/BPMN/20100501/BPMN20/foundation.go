package BPMN20

type BaseElement struct {
	Id       string
	Incoming []string
	Outgoing []string
	Type     BaseElementType
}

type BaseElementType int8

const (
	NotYetSupportedType BaseElementType = 0
	ServiceTaskType     BaseElementType = 1
)

func FindTargetRefs(sequenceFlows []TSequenceFlow, withId func(string) bool) (ret []string) {
	for _, flow := range sequenceFlows {
		if withId(flow.Id) {
			ret = append(ret, flow.TargetRef)
		}
	}
	return
}
