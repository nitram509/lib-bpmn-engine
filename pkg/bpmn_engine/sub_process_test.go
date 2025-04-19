package bpmn_engine

import (
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

func Test_subprocess(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/subprocess.bpmn")
	bpmnEngine.NewTaskHandler().Id("sub-process-a").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-in-sub-a").Handler(cp.TaskHandler)

	// when
	_, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-in-sub-a"))
}

func Test_subprocess_with_gateways(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/subprocess-with-gateways.bpmn")
	bpmnEngine.NewTaskHandler().Id("random_generator").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("print_10").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("print_20").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("print_25").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("print_30").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("say_goodbye").Handler(cp.TaskHandler)

	variables := map[string]interface{}{
		"COUNT": 28,
	}

	// when
	_, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, variables)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("random_generator,print_30,say_goodbye"))
}

func Test_subprocess_multiple_intermediate_catch_events_implicit_fork_and_exclusive_gateway_COMPLETED(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, err := bpmnEngine.LoadFromFile("../../test-cases/subprocess-message-multiple-intermediate-catch-events-exclusive.bpmn")
	if err != nil {
		t.Fatal(err)
	}
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-1", nil)
	then.AssertThat(t, err, is.Nil())
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2", nil)
	then.AssertThat(t, err, is.Nil())
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-3", nil)
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(Completed))
}

func Test_subprocess_link_events_are_thrown_and_caught_and_flow_continued(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/subprocess-simple-link-events.bpmn")
	bpmnEngine.NewTaskHandler().Type("task").Handler(cp.TaskHandler)
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance.ActivityState, is.EqualTo(Completed))
	then.AssertThat(t, cp.CallPath, is.EqualTo("Task-A,Sub-Task-A,Sub-Task-B,Task-B"))
}

func Test_subprocess_ForkControlledExclusiveJoin(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/subprocess-fork-controlled-exclusive-join.bpmn")
	bpmnEngine.NewTaskHandler().Type("task").Handler(cp.TaskHandler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("Sub-A-Task-A1,Sub-A-Task-A2,Sub-A-Task-B1,Sub-A-Task-B1,Sub-B-Task-A1,Task-C,Task-C"))
}
func Test_subprocess_ForkControlledParallelJoin(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/subprocess-fork-controlled-parallel-join.bpmn")
	bpmnEngine.NewTaskHandler().Type("task").Handler(cp.TaskHandler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("Sub-A-Task-A1,Sub-A-Task-A2,Sub-A-Task-B1,Sub-B-Task-A1,Task-C"))
}
