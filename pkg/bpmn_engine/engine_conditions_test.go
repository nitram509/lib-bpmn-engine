package bpmn_engine

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func Test_exclusive_gateway_with_expressions_selects_one_and_not_the_other(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/exclusive-gateway-with-condition.bpmn")
	bpmnEngine.AddTaskHandler("task-a", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("task-b", cp.CallPathHandler)
	variables := map[string]interface{}{
		"price": -50,
	}

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, variables)

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-b"))
}
