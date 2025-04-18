package bpmn_engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/corbym/gocrest/has"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

func Test_EventBasedGateway_selects_path_where_timer_occurs(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-timer-event.bpmn")
	bpmnEngine.NewTaskHandler().Id("task-for-message").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-for-timer").Handler(cp.TaskHandler)
	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// when
	bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "message", nil)
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-for-message"))
}

func Test_InvalidTimer_will_stop_execution_and_return_err(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-invalid-timer-event.bpmn")
	bpmnEngine.NewTaskHandler().Id("task-for-timer").Handler(cp.TaskHandler)
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, instance.ActivityState, is.EqualTo(Failed))
	then.AssertThat(t, err, is.Not(is.Nil()))
	then.AssertThat(t, err.Error(), has.Prefix("Error evaluating expression in intermediate timer cacht event element id="))
	then.AssertThat(t, cp.CallPath, is.EqualTo(""))
}

func Test_EventBasedGateway_selects_path_where_message_received(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-timer-event.bpmn")
	bpmnEngine.NewTaskHandler().Id("task-for-message").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-for-timer").Handler(cp.TaskHandler)
	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// when
	time.Sleep((1 * time.Second) + (1 * time.Millisecond))
	_, err := bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-for-timer"))
}

func Test_EventBasedGateway_selects_just_one_path(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-timer-event.bpmn")
	bpmnEngine.NewTaskHandler().Id("task-for-message").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-for-timer").Handler(cp.TaskHandler)
	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// when
	time.Sleep((1 * time.Second) + (1 * time.Millisecond))
	err := bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "message", nil)
	then.AssertThat(t, err, is.Nil())
	_, err = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.AllOf(
		has.Prefix("task-for"),
		is.Not(is.ValueContaining(","))),
	)
}

func Test_mapping_timer_state_can_always_be_mapped_to_activity_state(t *testing.T) {
	tests := []struct {
		ts TimerState
	}{
		{TimerCreated},
		{TimerCancelled},
		{TimerTriggered},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.ts), func(t *testing.T) {
			timer := &Timer{
				TimerState: test.ts,
			}
			activityState := timer.State()
			timer.TimerState = "" // delete state, to see if mapping works
			timer.SetState(activityState)
			then.AssertThat(t, timer.TimerState, is.EqualTo(test.ts))
		})
	}
}
