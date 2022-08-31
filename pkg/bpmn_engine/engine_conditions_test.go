package bpmn_engine

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
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

func Test_multiple_intermediate_catch_events_possible(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-multiple-intermediate-catch-events.bpmn")
	bpmnEngine.AddTaskHandler("task1", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("task2", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("task3", cp.CallPathHandler)
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2")
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	//then
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.COMPLETED))
	then.AssertThat(t, cp.CallPath, is.EqualTo("task2"))
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

func Test_mathematical_expression_evaluates(t *testing.T) {
	variables := map[string]interface{}{
		"foo": 3,
		"bar": 7,
		"sum": 10,
	}

	result, err := evaluateExpression("sum >= foo + bar", variables)

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, result, is.True())
}

func Test_evaluation_error_percolates_up(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/exclusive-gateway-with-condition.bpmn")

	// when
	// don't provide variables, for execution
	_, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, err, is.Not(is.Nil()))
}
