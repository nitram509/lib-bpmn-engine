package engine

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/nitram509/golib-bpmn-model/pgk/spec/BPMN/20100501/BPMN20"
	"testing"
)

func TestEngine(t *testing.T) {
	var definitions BPMN20.TDefinitions
	engine := BpmnEngineFromDefinitions(definitions)

	then.AssertThat(t, engine, is.Not(is.Nil()))
}
