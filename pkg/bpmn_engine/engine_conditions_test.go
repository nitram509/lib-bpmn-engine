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

func Test_exclusive_gateway_with_expressions_selects_default(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/exclusive-gateway-with-condition-and-default.bpmn")
	bpmnEngine.AddTaskHandler("task-a", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("task-b", cp.CallPathHandler)
	variables := map[string]interface{}{
		"price": -1,
	}

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, variables)

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-b"))
}

func Test_boolean_expression_evaluates(t *testing.T) {
	variables := map[string]interface{}{
		"aValue": 3,
	}

	result, err := evaluateExpression("aValue > 1", variables)

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, result, is.True())
}

func Test_boolean_expression_with_equalsign_evaluates(t *testing.T) {
	variables := map[string]interface{}{
		"aValue": 3,
	}

	result, err := evaluateExpression("= aValue > 1", variables)

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, result, is.True())
}
