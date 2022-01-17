package bpmn_engine

import (
	"github.com/corbym/gocrest/has"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
	"time"
)

func TestEventBasedGatewaySelectsPathWhereTimerOccurs(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-timer-event.bpmn")
	bpmnEngine.AddTaskHandler("task-for-message", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("task-for-timer", cp.CallPathHandler)
	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// when
	bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "message")
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-for-message"))
}

func TestEventBasedGatewaySelectsPathWhereMessageReceived(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-timer-event.bpmn")
	bpmnEngine.AddTaskHandler("task-for-message", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("task-for-timer", cp.CallPathHandler)
	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// when
	time.Sleep((1 * time.Second) + (1 * time.Millisecond))
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-for-timer"))
}

func TestEventBasedGatewaySelectsJustOnePath(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-timer-event.bpmn")
	bpmnEngine.AddTaskHandler("task-for-message", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("task-for-timer", cp.CallPathHandler)
	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// when
	time.Sleep((1 * time.Second) + (1 * time.Millisecond))
	bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "message")
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())

	// then
	then.AssertThat(t, cp.CallPath, is.AllOf(
		has.Prefix("task-for"),
		is.Not(is.ValueContaining(","))),
	)
}
