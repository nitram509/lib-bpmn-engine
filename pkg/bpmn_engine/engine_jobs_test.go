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

func Test_a_job_can_fail_and_keeps_the_instance_in_active_state(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")
	bpmnEngine.AddTaskHandler("id", jobFailHandler)

	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	then.AssertThat(t, instance.state, is.EqualTo(process_instance.ACTIVE))
}

func Test_simple_count_loop(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-count-loop.bpmn")
	bpmnEngine.AddTaskHandler("id-increaseCounter", increaseCounterHandler)

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
	bpmnEngine.AddTaskHandler("do-nothing", jobCompleteHandler)
	bpmnEngine.AddTaskHandler("validate", func(job ActivatedJob) {
		attempts := job.GetVariable(varEngineValidationAttempts).(int)
		foobar := attempts >= 1
		attempts++
		job.SetVariable(varEngineValidationAttempts, attempts)
		job.SetVariable(varFoobar, foobar)
		job.Complete()
	})

	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, vars) // should stop at the intermediate message catch event

	bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg")
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey()) // again, should stop at the intermediate message catch event
	// validation happened
	bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg")
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey()) // should finish
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
	bpmnEngine.AddTaskHandler("id", func(aj ActivatedJob) {
		then.AssertThat(t, aj.GetElementId(), is.Not(is.Empty()))
		then.AssertThat(t, aj.GetCreatedAt(), is.Not(is.Nil()))
		then.AssertThat(t, aj.GetState(), is.Not(is.EqualTo(activity.Active)))
		then.AssertThat(t, aj.GetInstanceKey(), is.Not(is.EqualTo(int64(0))))
		then.AssertThat(t, aj.GetKey(), is.Not(is.EqualTo(int64(0))))
		then.AssertThat(t, aj.GetBpmnProcessId(), is.Not(is.Empty()))
		then.AssertThat(t, aj.GetProcessDefinitionKey(), is.Not(is.EqualTo(int64(0))))
		then.AssertThat(t, aj.GetProcessDefinitionVersion(), is.Not(is.EqualTo(int32(0))))
		then.AssertThat(t, aj.GetProcessInstanceKey(), is.Not(is.EqualTo(int64(0))))
	})

	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	then.AssertThat(t, instance.state, is.EqualTo(process_instance.ACTIVE))
}

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


func TestTaskInputOutput(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// give
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/service-task-input-output.bpmn")
	bpmnEngine.AddTaskHandler("input-task-1", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("input-task-2", cp.CallPathHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	if err != nil {
		panic(err)
	}

	// then
	for _, job := range bpmnEngine.jobs {
		then.AssertThat(t, job.State, is.EqualTo(activity.Completed))
	}
	then.AssertThat(t, cp.CallPath, is.EqualTo("input-task-1,input-task-2"))
	then.AssertThat(t, pi.GetVariable("id"), is.EqualTo(1))
	then.AssertThat(t, pi.GetVariable("orderId"), is.EqualTo(1234))
	then.AssertThat(t, pi.GetVariable("order"), is.EqualTo(map[string]interface{}{
		"name": "order1",
		"id":   "1234",
	}))
	then.AssertThat(t, pi.GetVariable("orderName").(string), is.EqualTo("order1"))
}

func TestInvalidTaskInput(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// give
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/service-task-invalid-input.bpmn")
	bpmnEngine.AddTaskHandler("invalid-input", cp.CallPathHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	if err != nil {
		panic(err)
	}
	// then
	then.AssertThat(t, pi.GetVariable("id"), is.EqualTo(nil))
	then.AssertThat(t, cp.CallPath, is.EqualTo(""))
}

func TestInvalidTaskOutput(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// give
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/service-task-invalid-output.bpmn")
	bpmnEngine.AddTaskHandler("invalid-output", cp.CallPathHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	if err != nil {
		panic(err)
	}
	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("invalid-output"))
	then.AssertThat(t, pi.GetVariable("order"), is.EqualTo(nil))
}
