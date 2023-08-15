package tests

import (
	"github.com/corbym/gocrest/has"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
	"testing"
)

type CallPath struct {
	CallPath string
}

func (callPath *CallPath) CallPathHandler(job bpmn_engine.ActivatedJob) {
	if len(callPath.CallPath) > 0 {
		callPath.CallPath += ","
	}
	callPath.CallPath += job.GetElementId()
	job.Complete()
}

func Test_Marshal_Unmarshal_Jobs(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New("name")

	// given
	pi, err := bpmnEngine.LoadFromFile("../test-cases/simple_task.bpmn")
	then.AssertThat(t, err, is.Nil())

	// when
	instance, err := bpmnEngine.CreateAndRunInstance(pi.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())
	bytes := bpmnEngine.Marshal()
	then.AssertThat(t, len(bytes), is.GreaterThan(32))

	// when
	bpmnEngine = bpmn_engine.Unmarshal(bytes)

	// then
	instance, err = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.ACTIVE))
}

func Test_Marshal_Unmarshal_Remain_Handler(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New("name")
	cp := CallPath{}

	// given
	pi, err := bpmnEngine.LoadFromFile("../test-cases/simple_task.bpmn")
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.NewTaskHandler().Id("id").Handler(cp.CallPathHandler)

	// when
	instance, err := bpmnEngine.CreateInstance(pi.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.READY))
	bytes := bpmnEngine.Marshal()

	// when
	newEngine := bpmn_engine.Unmarshal(bytes)
	newEngine.NewTaskHandler().Id("id").Handler(cp.CallPathHandler)

	// then
	instance, err = newEngine.RunOrContinueInstance(instance.GetInstanceKey())
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance.GetState(), is.EqualTo(process_instance.COMPLETED))

	then.AssertThat(t, cp.CallPath, is.EqualTo("id"))
}

func Test_Marshal_Unmarshal_IntermediateCatchEvents(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New("name")

	// given
	pi, err := bpmnEngine.LoadFromFile("../test-cases/simple-intermediate-message-catch-event.bpmn")
	then.AssertThat(t, err, is.Nil())

	// when
	_, err = bpmnEngine.CreateAndRunInstance(pi.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())
	bytes := bpmnEngine.Marshal()
	then.AssertThat(t, len(bytes), is.GreaterThan(32))

	// when
	bpmnEngine = bpmn_engine.Unmarshal(bytes)

	// then
	subscriptions := bpmnEngine.GetMessageSubscriptions()
	then.AssertThat(t, subscriptions, has.Length(1))
}

func Test_Marshal_Unmarshal_IntermediateTimerEvents(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New("name")

	// given
	pi, err := bpmnEngine.LoadFromFile("../test-cases/message-intermediate-timer-event.bpmn")
	then.AssertThat(t, err, is.Nil())

	// when
	_, err = bpmnEngine.CreateAndRunInstance(pi.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())
	bytes := bpmnEngine.Marshal()
	then.AssertThat(t, len(bytes), is.GreaterThan(32))

	// when
	bpmnEngine = bpmn_engine.Unmarshal(bytes)

	// then
	timers := bpmnEngine.GetTimersScheduled()
	then.AssertThat(t, timers, has.Length(1))
}
