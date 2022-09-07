package bpmn_engine

import (
	"github.com/corbym/gocrest/has"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/activity"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
	"testing"
)

const varCounter = "counter"
const varEngineValidationAttempts = "engineValidationAttempts"
const varFoobar = "foobar"

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
