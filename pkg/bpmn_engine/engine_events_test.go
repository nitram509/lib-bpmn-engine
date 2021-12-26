package bpmn_engine

import (
	"github.com/corbym/gocrest/has"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/activity"
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
	then.AssertThat(t, pi.GetState(), is.EqualTo(BPMN20.ProcessInstanceReady))
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
	then.AssertThat(t, pi.GetState(), is.EqualTo(BPMN20.ProcessInstanceActive).
		Reason("Since the BPMN contains an intermediate catch event, the process instance must be active and can't complete."))
	then.AssertThat(t, procInst.GetState(), is.EqualTo(BPMN20.ProcessInstanceActive))
}

func Test_IntermediateCatchEvent_received_message_completes_the_instance(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event.bpmn")
	pi, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// when
	bpmnEngine.PublishEventForInstance(pi.GetInstanceKey(), "event-1")
	bpmnEngine.RunOrContinueInstance(pi.GetInstanceKey())

	// then
	then.AssertThat(t, pi.GetState(), is.EqualTo(BPMN20.ProcessInstanceCompleted))
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
	then.AssertThat(t, pi.GetState(), is.EqualTo(BPMN20.ProcessInstanceCompleted))
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
	then.AssertThat(t, subscription.State, is.EqualTo(activity.Ready))
}
