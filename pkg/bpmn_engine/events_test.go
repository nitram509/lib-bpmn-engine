package bpmn_engine

import (
	"testing"

	"github.com/corbym/gocrest/has"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

func Test_creating_a_process_sets_state_to_READY(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event.bpmn")

	// when
	pi, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)
	// then
	then.AssertThat(t, pi.GetState(), is.EqualTo(Ready))
}

func Test_running_a_process_sets_state_to_ACTIVE(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event.bpmn")

	// when
	pi, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)
	procInst, _ := bpmnEngine.RunOrContinueInstance(pi.GetInstanceKey())

	// then
	then.AssertThat(t, pi.GetState(), is.EqualTo(Active).
		Reason("Since the BPMN contains an intermediate catch event, the process instance must be active and can't complete."))
	then.AssertThat(t, procInst.GetState(), is.EqualTo(Active))
}

func Test_IntermediateCatchEvent_received_message_completes_the_instance(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event.bpmn")
	pi, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// when
	err := bpmnEngine.PublishEventForInstance(pi.GetInstanceKey(), "globalMsgRef", nil)
	then.AssertThat(t, err, is.Nil())
	_, err = bpmnEngine.RunOrContinueInstance(pi.GetInstanceKey())
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, pi.GetState(), is.EqualTo(Completed))
}

func Test_IntermediateCatchEvent_message_can_be_published_before_running_the_instance(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event.bpmn")
	pi, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)

	// when
	bpmnEngine.PublishEventForInstance(pi.GetInstanceKey(), "globalMsgRef", nil)
	bpmnEngine.RunOrContinueInstance(pi.GetInstanceKey())

	// then
	then.AssertThat(t, pi.GetState(), is.EqualTo(Completed))
}

func Test_IntermediateCatchEvent_a_catch_event_produces_an_active_subscription(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event.bpmn")
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	subscriptions := bpmnEngine.GetMessageSubscriptions()

	then.AssertThat(t, subscriptions, has.Length(1))
	subscription := subscriptions[0]
	then.AssertThat(t, subscription.Name, is.EqualTo("event-1"))
	then.AssertThat(t, subscription.ElementId, is.EqualTo("id-1"))
	then.AssertThat(t, subscription.MessageState, is.EqualTo(Active))
}

func Test_Having_IntermediateCatchEvent_and_ServiceTask_in_parallel_the_process_state_is_maintained(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event-and-parallel-tasks.bpmn")
	instance, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)
	bpmnEngine.NewTaskHandler().Id("task-1").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-2").Handler(cp.TaskHandler)

	// when
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(Active))

	// when
	bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "event-1", nil)
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-2,task-1"))
	then.AssertThat(t, instance.GetState(), is.EqualTo(Completed))
}

func Test_multiple_intermediate_catch_events_possible(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-multiple-intermediate-catch-events.bpmn")
	bpmnEngine.NewTaskHandler().Id("task1").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task2").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task3").Handler(cp.TaskHandler)
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2", nil)
	then.AssertThat(t, err, is.Nil())
	_, err = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task2"))
	// then still active, since there's an implicit fork
	then.AssertThat(t, instance.GetState(), is.EqualTo(Active))
}

func Test_multiple_intermediate_catch_events_implicit_fork_and_merged_COMPLETED(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-multiple-intermediate-catch-events-merged.bpmn")
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-1", nil)
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2", nil)
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-3", nil)
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(Completed))
}

func Test_multiple_intermediate_catch_events_implicit_fork_and_merged_ACTIVE(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-multiple-intermediate-catch-events-merged.bpmn")
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2", nil)
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(Active))
}

func Test_multiple_intermediate_catch_events_implicit_fork_and_parallel_gateway_COMPLETED(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-multiple-intermediate-catch-events-parallel.bpmn")
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-1", nil)
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2", nil)
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-3", nil)
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(Completed))
}

func Test_multiple_intermediate_catch_events_implicit_fork_and_parallel_gateway_ACTIVE(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-multiple-intermediate-catch-events-parallel.bpmn")
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2", nil)
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(Active))
}
func Test_multiple_intermediate_catch_events_implicit_fork_and_exclusive_gateway_COMPLETED(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-multiple-intermediate-catch-events-exclusive.bpmn")
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-1", nil)
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2", nil)
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-3", nil)
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(Completed))
}

func Test_multiple_intermediate_catch_events_implicit_fork_and_exclusive_gateway_ACTIVE(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-multiple-intermediate-catch-events-exclusive.bpmn")
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-event-2", nil)
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(Active))
}

func Test_publishing_a_random_message_does_no_harm(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event.bpmn")
	instance, err := bpmnEngine.CreateInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// when
	err = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "random-message", nil)
	then.AssertThat(t, err, is.Nil())
	_, err = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance.GetState(), is.EqualTo(Active))
}

func Test_eventBasedGateway_just_fires_one_event_and_instance_COMPLETED(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-EventBasedGateway.bpmn")
	instance, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)
	bpmnEngine.NewTaskHandler().Id("task-a").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("task-b").Handler(cp.TaskHandler)

	// when
	_ = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg-b", nil)
	_, err := bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-b"))
	then.AssertThat(t, instance.GetState(), is.EqualTo(Completed))
}

func Test_intermediate_message_catch_event_publishes_variables_into_instance(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-intermediate-message-catch-event.bpmn")
	instance, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)

	// when
	vars := map[string]interface{}{"foo": "bar"}
	_ = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg", vars)
	_, _ = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(Completed))
	then.AssertThat(t, instance.GetVariable("foo"), is.EqualTo("bar"))
	then.AssertThat(t, instance.GetVariable("mappedFoo"), is.EqualTo("bar"))
}

func Test_intermediate_message_catch_event_output_mapping_failed(t *testing.T) {
	// setup
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-intermediate-message-catch-event-broken.bpmn")
	instance, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)

	// when
	_ = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg", nil)
	_, _ = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, instance.GetState(), is.EqualTo(Failed))
	then.AssertThat(t, instance.GetVariable("mappedFoo"), is.Nil())
	then.AssertThat(t, bpmnEngine.messageSubscriptions[0].MessageState, is.EqualTo(Failed))
}
