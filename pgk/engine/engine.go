package engine

import (
	"github.com/nitram509/golib-bpmn-model/pgk/spec/BPMN/20100501/BPMN20NEW"
)

type BpmnEngine interface {
}

func BpmnEngineFromDefinitions(definitions BPMN20NEW.TDefinitions) BpmnEngine {
	return nil
}
