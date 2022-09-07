package bpmn_engine

import (
	"github.com/corbym/gocrest/has"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/activity"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
	"testing"
)

func Test_creating_a_process_sets_state_to_READY(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event.bpmn")

	// when
	pi, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)
	// then
	then.AssertThat(t, pi.GetState(), is.EqualTo(process_instance.READY))
}

func Test_running_a_process_sets_state_to_ACTIVE(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event.bpmn")

	// when
	pi, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)
	procInst, _ := bpmnEngine.RunOrContinueInstance(pi.GetInstanceKey())

	// then
	then.AssertThat(t, pi.GetState(), is.EqualTo(process_instance.ACTIVE).
		Reason("Since the BPMN contains an intermediate catch event, the process instance must be active and can't complete."))
	then.AssertThat(t, procInst.GetState(), is.EqualTo(process_instance.ACTIVE))
}

func Test_IntermediateCatchEvent_received_message_completes_the_instance(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event.bpmn")
	pi, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// when
	bpmnEngine.PublishEventForInstance(pi.GetInstanceKey(), "globalMsgRef")
	bpmnEngine.RunOrContinueInstance(pi.GetInstanceKey())

	// then
	then.AssertThat(t, pi.GetState(), is.EqualTo(process_instance.COMPLETED))
}

func Test_IntermediateCatchEvent_message_can_be_published_before_running_the_instance(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event.bpmn")
	pi, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)

	// when
	bpmnEngine.PublishEventForInstance(pi.GetInstanceKey(), "globalMsgRef")
	bpmnEngine.RunOrContinueInstance(pi.GetInstanceKey())

	// then
	then.AssertThat(t, pi.GetState(), is.EqualTo(process_instance.COMPLETED))
}

func Test_IntermediateCatchEvent_a_catch_event_produces_an_active_subscription(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event.bpmn")
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	subscriptions := bpmnEngine.GetMessageSubscriptions()

	then.AssertThat(t, subscriptions, has.Length(1))
	subscription := subscriptions[0]
	then.AssertThat(t, subscription.Name, is.EqualTo("event-1"))
	then.AssertThat(t, subscription.ElementId, is.EqualTo("id-1"))
	then.AssertThat(t, subscription.State, is.EqualTo(activity.Active))
}

func Test_Having_IntermediateCatchEvent_and_ServiceTask_in_parallel_the_process_state_is_maintained(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event-and-parallel-tasks.bpmn")
	instance, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)
	bpmnEngine.AddTaskHandler("task-1", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("task-2", cp.CallPathHandler)

	// when
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.ACTIVE))

	// when
	bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "event-1")
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-2,task-1"))
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.COMPLETED))
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

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task2"))
	// then still active, since there's an implicit fork
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.ACTIVE))
}

func Test_multiple_intermediate_catch_events_implicit_fork_and_merged_COMPLETED(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-multiple-intermediate-catch-events-merged.bpmn")
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-1")
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2")
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-3")
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.COMPLETED))
}

func Test_multiple_intermediate_catch_events_implicit_fork_and_merged_ACTIVE(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-multiple-intermediate-catch-events-merged.bpmn")
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2")
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.ACTIVE))
}

func Test_multiple_intermediate_catch_events_implicit_fork_and_parallel_gateway_COMPLETED(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-multiple-intermediate-catch-events-parallel.bpmn")
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-1")
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2")
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-3")
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.COMPLETED))
}

func Test_multiple_intermediate_catch_events_implicit_fork_and_parallel_gateway_ACTIVE(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-multiple-intermediate-catch-events-parallel.bpmn")
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2")
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.ACTIVE))
}
func Test_multiple_intermediate_catch_events_implicit_fork_and_exclusive_gateway_COMPLETED(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-multiple-intermediate-catch-events-exclusive.bpmn")
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-1")
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2")
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-3")
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.COMPLETED))
}

func Test_multiple_intermediate_catch_events_implicit_fork_and_exclusive_gateway_ACTIVE(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-multiple-intermediate-catch-events-exclusive.bpmn")
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2")
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.ACTIVE))
}

func Test_publishing_a_random_message_does_no_harm(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event.bpmn")
	instance, err := bpmnEngine.CreateInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "random-message")
	then.AssertThat(t, err, is.Nil())
	_, err = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.ACTIVE))
}

func Test_eventBasedGateway_just_fires_one_event_and_instance_COMPLETED(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-EventBasedGateway.bpmn")
	instance, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)
	bpmnEngine.AddTaskHandler("task-a", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("task-b", cp.CallPathHandler)

	// when
	bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-b")
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-b"))
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.COMPLETED))
}
