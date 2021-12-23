package bpmn_engine

type MessageState struct {
}

type MessageSubscription struct {
	ElementId          string
	ElementInstanceKey int64
	Name               string
	CorrelationKey     string
	state              MessageState
}

func (state *BpmnEngineState) PublishEvent(message string, correlationKey string) {
	//state.messageSubscriptions
}
