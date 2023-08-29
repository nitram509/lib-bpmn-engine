package bpmn_engine

import "fmt"

type BpmnEngineError struct {
	Msg string
}

func (e *BpmnEngineError) Error() string {
	return e.Msg
}

// newEngineErrorf uses fmt.Sprintf(format, a...) to format the message
func newEngineErrorf(format string, a ...interface{}) error {
	return &BpmnEngineError{
		Msg: fmt.Sprintf(format, a...),
	}
}

type BpmnEngineUnmarshallingError struct {
	Msg string
	Err error
}

func (e *BpmnEngineUnmarshallingError) Error() string {
	return e.Msg + ": " + e.Err.Error()
}

type ExpressionEvaluationError struct {
	Msg string
	Err error
}

func (e *ExpressionEvaluationError) Error() string {
	return e.Msg + "\nerror: " + e.Err.Error()
}
