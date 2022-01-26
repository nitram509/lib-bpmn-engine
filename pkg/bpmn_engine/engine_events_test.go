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
	bpmnEngine.PublishEventForInstance(pi.GetInstanceKey(), "event-1")
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

func Test_two(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/two-tasks-shared-message-event.bpmn")
	instance, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)
	bpmnEngine.AddTaskHandler("task-a", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("task-b", cp.CallPathHandler)

	// when
	bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "shared-msg")
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-a"))
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.COMPLETED))
}
