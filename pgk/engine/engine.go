package engine

import (
	"github.com/nitram509/golib-bpmn-model/pgk/spec/BPMN/20100501/BPMN20"
)

type BpmnEngine interface {
}

func BpmnEngineFromDefinitions(definitions BPMN20.TDefinitions) BpmnEngine {
	return nil
}
