package bpmn

import (
	"testing"

	"github.com/corbym/gocrest/has"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

func Test_exclusive_gateway_with_expressions_selects_one_and_not_the_other(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/exclusive-gateway-with-condition.bpmn")
	bpmnEngine.NewTaskHandler().Id("task-a").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-b").Handler(cp.TaskHandler)
	variables := map[string]interface{}{
		"price": -50,
	}

	// when
	_, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, variables)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-b"))
}

func Test_exclusive_gateway_with_expressions_selects_default(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/exclusive-gateway-with-condition-and-default.bpmn")
	bpmnEngine.NewTaskHandler().Id("task-a").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-b").Handler(cp.TaskHandler)
	variables := map[string]interface{}{
		"price": -1,
	}

	// when
	_, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, variables)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-b"))
}

func Test_exclusive_gateway_executes_just_one_matching_path(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/exclusive-gateway-multiple-tasks.bpmn")
	bpmnEngine.NewTaskHandler().Id("task-a").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-b").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-default").Handler(cp.TaskHandler)
	variables := map[string]interface{}{
		"price": 0,
	}

	// when
	_, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, variables)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-a"))
}

func Test_exclusive_gateway_executes_just_no_matching_path_default_is_used(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/exclusive-gateway-multiple-tasks.bpmn")
	bpmnEngine.NewTaskHandler().Id("task-a").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-b").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-default").Handler(cp.TaskHandler)
	variables := map[string]interface{}{
		"price": -99,
	}

	// when
	_, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, variables)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-default"))
}

func Test_exclusive_gateway_executes_just_no_matching_no_default_error_thrown(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/exclusive-gateway-multiple-tasks-no-default.bpmn")
	bpmnEngine.NewTaskHandler().Id("task-a").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-b").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-default").Handler(cp.TaskHandler)
	variables := map[string]interface{}{
		"price": -99,
	}

	// when
	_, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, variables)

	// then
	then.AssertThat(t, err, is.Not(is.Nil()))
	then.AssertThat(t, cp.CallPath, is.EqualTo(""))
}

func Test_boolean_expression_evaluates(t *testing.T) {
	variables := map[string]interface{}{
		"aValue": 3,
	}

	result, err := evaluateExpression("aValue > 1", variables)

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, result, is.True())
}

func Test_boolean_expression_with_equal_sign_evaluates(t *testing.T) {
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
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/exclusive-gateway-with-condition.bpmn")

	// when
	// don't provide variables, for execution to get an evaluation error
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, instance.State, is.EqualTo(Failed))
	then.AssertThat(t, err, is.Not(is.Nil()))
	then.AssertThat(t, err.Error(), has.Prefix("Error evaluating expression in flow element id="))
}

func Test_inclusive_gateway_with_expressions_selects_one_and_not_the_other(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/inclusive-gateway-with-condition.bpmn")
	bpmnEngine.NewTaskHandler().Id("task-a").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-b").Handler(cp.TaskHandler)
	variables := map[string]interface{}{
		"price": -50,
	}

	// when
	_, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, variables)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-b"))
}

func Test_inclusive_gateway_with_expressions_selects_default(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/inclusive-gateway-with-condition-and-default.bpmn")
	bpmnEngine.NewTaskHandler().Id("task-a").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-b").Handler(cp.TaskHandler)
	variables := map[string]interface{}{
		"price": -1,
	}

	// when
	_, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, variables)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-b"))
}

func Test_inclusive_gateway_executes_all_paths(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/inclusive-gateway-multiple-tasks.bpmn")
	bpmnEngine.NewTaskHandler().Id("task-a").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-b").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-default").Handler(cp.TaskHandler)
	variables := map[string]interface{}{
		"price": 0,
	}

	// when
	_, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, variables)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-a,task-b,task-default"))
}
