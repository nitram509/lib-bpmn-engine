package bpmn_engine

type BpmnEngineError struct {
	Msg string
}

func (e *BpmnEngineError) Error() string {
	return e.Msg
}

type BpmnEngineUnmarshallingError struct {
	Msg string
	Err error
}

func (e *BpmnEngineUnmarshallingError) Error() string {
	return e.Msg + ": " + e.Err.Error()
}
