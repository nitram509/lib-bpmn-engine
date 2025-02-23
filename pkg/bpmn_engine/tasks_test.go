package bpmn_engine

import (
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

func Test_user_tasks_can_be_handled(t *testing.T) {
	// setup
	bpmnEngine := New()
	process, err := bpmnEngine.LoadFromFile("../../test-cases/simple-user-task.bpmn")
	then.AssertThat(t, err, is.Nil())
	cp := CallPath{}
	bpmnEngine.NewTaskHandler().Id("user-task").Handler(cp.TaskHandler)

	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	then.AssertThat(t, instance.State, is.EqualTo(Completed))
	then.AssertThat(t, cp.CallPath, is.EqualTo("user-task"))
}

func Test_user_tasks_can_be_continue(t *testing.T) {
	// setup
	bpmnEngine := New()
	process, err := bpmnEngine.LoadFromFile("../../test-cases/simple-user-task.bpmn")
	then.AssertThat(t, err, is.Nil())
	cp := CallPath{}

	instance, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)

	userConfirm := false
	bpmnEngine.NewTaskHandler().Id("user-task").Handler(func(job ActivatedJob) {
		if userConfirm {
			cp.TaskHandler(job)
		}
	})
	_, err = bpmnEngine.RunOrContinueInstance(instance.InstanceKey)
	then.AssertThat(t, err, is.Nil())

	userConfirm = true

	_, err = bpmnEngine.RunOrContinueInstance(instance.InstanceKey)
	then.AssertThat(t, err, is.Nil())

	then.AssertThat(t, instance.State, is.EqualTo(Completed))
	then.AssertThat(t, cp.CallPath, is.EqualTo("user-task"))
}
