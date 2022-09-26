package bpmn_engine

import (
	"testing"

	"github.com/corbym/gocrest/has"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/activity"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
)

const (
	varCounter                  = "counter"
	varEngineValidationAttempts = "engineValidationAttempts"
	varFoobar                   = "foobar"
)

func increaseCounterHandler(job ActivatedJob) {
	counter := job.GetVariable(varCounter).(int)
	counter++
	job.SetVariable(varCounter, counter)
	job.Complete()
}

func jobFailHandler(job ActivatedJob) {
	job.Fail("just because I can")
}

func jobCompleteHandler(job ActivatedJob) {
	job.Complete()
}

func Test_a_job_can_fail_and_keeps_the_instance_in_active_state(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")
	bpmnEngine.NewTaskHandler().Id("id").Handler(jobFailHandler)

	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	then.AssertThat(t, instance.state, is.EqualTo(process_instance.ACTIVE))
}

func Test_simple_count_loop(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-count-loop.bpmn")
	bpmnEngine.NewTaskHandler().Id("id-increaseCounter").Handler(increaseCounterHandler)

	vars := map[string]interface{}{}
	vars[varCounter] = 0
	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, vars)

	then.AssertThat(t, instance.GetVariable(varCounter), is.EqualTo(4))
	then.AssertThat(t, instance.state, is.EqualTo(process_instance.COMPLETED))
}

func Test_simple_count_loop_with_message(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-count-loop-with-message.bpmn")

	vars := map[string]interface{}{}
	vars[varEngineValidationAttempts] = 0
	bpmnEngine.NewTaskHandler().Id("do-nothing").Handler(jobCompleteHandler)
	bpmnEngine.NewTaskHandler().Id("validate").Handler(func(job ActivatedJob) {
		attempts := job.GetVariable(varEngineValidationAttempts).(int)
		foobar := attempts >= 1
		attempts++
		job.SetVariable(varEngineValidationAttempts, attempts)
		job.SetVariable(varFoobar, foobar)
		job.Complete()
	})

	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, vars) // should stop at the intermediate message catch event

	_ = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg", nil)
	_, _ = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey()) // again, should stop at the intermediate message catch event
	// validation happened
	_ = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg", nil)
	_, _ = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey()) // should finish
	// validation happened

	then.AssertThat(t, instance.GetVariable(varFoobar), is.EqualTo(true))
	then.AssertThat(t, instance.GetVariable(varEngineValidationAttempts), is.EqualTo(2))
	then.AssertThat(t, instance.state, is.EqualTo(process_instance.COMPLETED))

	// internal state expected
	then.AssertThat(t, bpmnEngine.GetMessageSubscriptions(), has.Length(2))
	then.AssertThat(t, bpmnEngine.GetMessageSubscriptions()[0].State, is.EqualTo(activity.Completed))
	then.AssertThat(t, bpmnEngine.GetMessageSubscriptions()[1].State, is.EqualTo(activity.Completed))
}

func Test_activated_job_data(t *testing.T) {
	bpmnEngine := New("name")
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")
	bpmnEngine.NewTaskHandler().Id("id").Handler(func(aj ActivatedJob) {
		then.AssertThat(t, aj.GetElementId(), is.Not(is.Empty()))
		then.AssertThat(t, aj.GetCreatedAt(), is.Not(is.Nil()))
		then.AssertThat(t, aj.GetKey(), is.Not(is.EqualTo(int64(0))))
		then.AssertThat(t, aj.GetBpmnProcessId(), is.Not(is.Empty()))
		then.AssertThat(t, aj.GetProcessDefinitionKey(), is.Not(is.EqualTo(int64(0))))
		then.AssertThat(t, aj.GetProcessDefinitionVersion(), is.Not(is.EqualTo(int32(0))))
		then.AssertThat(t, aj.GetProcessInstanceKey(), is.Not(is.EqualTo(int64(0))))
	})

	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	then.AssertThat(t, instance.state, is.EqualTo(process_instance.ACTIVE))
}

func Test_task_InputOutput_mapping_happy_path(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// give
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/service-task-input-output.bpmn")
	bpmnEngine.NewTaskHandler().Id("service-task-1").Handler(cp.CallPathHandler)
	bpmnEngine.NewTaskHandler().Id("user-task-2").Handler(cp.CallPathHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	for _, job := range bpmnEngine.jobs {
		then.AssertThat(t, job.State, is.EqualTo(activity.Completed))
	}
	then.AssertThat(t, cp.CallPath, is.EqualTo("service-task-1,user-task-2"))
	// id from input should not exist in instance variableContext
	then.AssertThat(t, pi.GetVariable("id"), is.EqualTo(nil))
	// output should exist in instance variableContext
	then.AssertThat(t, pi.GetVariable("dstcity"), is.EqualTo("beijing"))
	then.AssertThat(t, pi.GetVariable("order"), is.EqualTo(map[string]interface{}{
		"name": "order1",
		"id":   "1234",
	}))
	then.AssertThat(t, pi.GetVariable("orderId"), is.EqualTo(1234))
	then.AssertThat(t, pi.GetVariable("orderName"), is.EqualTo(nil))
}

func Test_instance_fails_on_Invalid_Input_mapping(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// give
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/service-task-invalid-input.bpmn")
	bpmnEngine.NewTaskHandler().Id("invalid-input").Handler(cp.CallPathHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo(""))
	then.AssertThat(t, pi.GetVariable("id"), is.EqualTo(nil))
	then.AssertThat(t, bpmnEngine.jobs[0].State, is.EqualTo(activity.Failed))
	then.AssertThat(t, pi.GetState(), is.EqualTo(process_instance.FAILED))
}

func Test_job_fails_on_Invalid_Output_mapping(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// give
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/service-task-invalid-output.bpmn")
	bpmnEngine.NewTaskHandler().Id("invalid-output").Handler(cp.CallPathHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("invalid-output"))
	then.AssertThat(t, pi.GetVariable("order"), is.EqualTo(nil))
	then.AssertThat(t, bpmnEngine.jobs[0].State, is.EqualTo(activity.Failed))
	then.AssertThat(t, pi.GetState(), is.EqualTo(process_instance.FAILED))
}

func Test_task_type_handler(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// give
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-task-with-type.bpmn")
	bpmnEngine.NewTaskHandler().Type("foobar").Handler(cp.CallPathHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("id"))
	then.AssertThat(t, pi.GetState(), is.EqualTo(process_instance.COMPLETED))
}

func Test_task_type_handler_ID_handler_has_precedence(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	calledHandler := "none"
	idHandler := func(job ActivatedJob) {
		calledHandler = "ID"
		job.Complete()
	}
	typeHandler := func(job ActivatedJob) {
		calledHandler = "TYPE"
		job.Complete()
	}
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-task-with-type.bpmn")

	// given reverse order of definition, means 'type:foobar' before 'id'
	bpmnEngine.NewTaskHandler().Type("foobar").Handler(typeHandler)
	bpmnEngine.NewTaskHandler().Id("id").Handler(idHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, calledHandler, is.EqualTo("ID"))
	then.AssertThat(t, pi.GetState(), is.EqualTo(process_instance.COMPLETED))
}

func Test_just_one_handler_called(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-task-with-type.bpmn")

	// given multiple matching handlers executed
	bpmnEngine.NewTaskHandler().Id("id").Handler(cp.CallPathHandler)
	bpmnEngine.NewTaskHandler().Id("id").Handler(cp.CallPathHandler)
	bpmnEngine.NewTaskHandler().Type("foobar").Handler(cp.CallPathHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("id").Reason("just one execution"))
	then.AssertThat(t, pi.GetState(), is.EqualTo(process_instance.COMPLETED))
}

func Test_assignee_and_candidate_groups_are_assigned_to_handler(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/user-tasks-with-assignments.bpmn")

	// given multiple matching handlers executed
	bpmnEngine.NewTaskHandler().Assignee("john.doe").Handler(cp.CallPathHandler)
	bpmnEngine.NewTaskHandler().CandidateGroups("marketing", "support").Handler(cp.CallPathHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("assignee-task,group-task"))
	then.AssertThat(t, pi.GetState(), is.EqualTo(process_instance.COMPLETED))
}
