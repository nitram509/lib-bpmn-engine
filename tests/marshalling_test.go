package tests

import (
	"os"
	"testing"
	"time"

	"github.com/corbym/gocrest/has"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
)

type CallPath struct {
	CallPath string
}

const EnableJsonDataDump = true

func (callPath *CallPath) CallPathHandler(job bpmn_engine.ActivatedJob) {
	if len(callPath.CallPath) > 0 {
		callPath.CallPath += ","
	}
	callPath.CallPath += job.ElementId()
	job.Complete()
}

func Test_Marshal_Unmarshal_Jobs(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New()

	// given
	pi, err := bpmnEngine.LoadFromFile("../test-cases/simple_task.bpmn")
	then.AssertThat(t, err, is.Nil())

	// when
	instance, err := bpmnEngine.CreateAndRunInstance(pi.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())
	bytes := bpmnEngine.Marshal()
	then.AssertThat(t, len(bytes), is.GreaterThan(32))

	if EnableJsonDataDump {
		os.WriteFile("temp.marshal.jobs.json", bytes, 0644)
	}

	// when
	bpmnEngine, err = bpmn_engine.Unmarshal(bytes)
	then.AssertThat(t, err, is.Nil())

	// then
	instance, err = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance.GetState(), is.EqualTo(bpmn_engine.Active))
}

func Test_Marshal_Unmarshal_partially_executed_jobs_continue_where_left_of_before_marshalling(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New()
	cp := CallPath{}
	bpmnEngine.NewTaskHandler().Id("id-a-1").Handler(cp.CallPathHandler)

	// given
	pi, err := bpmnEngine.LoadFromFile("../test-cases/parallel-gateway-flow.bpmn")
	then.AssertThat(t, err, is.Nil())

	// when
	instance, err := bpmnEngine.CreateAndRunInstance(pi.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, cp.CallPath, is.EqualTo("id-a-1"))

	instance, err = bpmnEngine.RunOrContinueInstance(instance.InstanceKey)
	bytes := bpmnEngine.Marshal()
	then.AssertThat(t, len(bytes), is.GreaterThan(32))

	if EnableJsonDataDump {
		os.WriteFile("temp.marshal.parallel-gateway-flow.json", bytes, 0644)
	}

	// when
	bpmnEngine, err = bpmn_engine.Unmarshal(bytes)
	bpmnEngine.NewTaskHandler().Id("id-b-1").Handler(cp.CallPathHandler)
	bpmnEngine.NewTaskHandler().Id("id-b-2").Handler(cp.CallPathHandler)
	then.AssertThat(t, err, is.Nil())

	// then
	instance, err = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance.GetState(), is.EqualTo(bpmn_engine.Completed))
	then.AssertThat(t, cp.CallPath, is.EqualTo("id-a-1,id-b-1,id-b-2"))

}

func Test_Marshal_Unmarshal_Remain_Handler(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New()
	cp := CallPath{}

	// given
	pi, err := bpmnEngine.LoadFromFile("../test-cases/simple_task.bpmn")
	then.AssertThat(t, err, is.Nil())
	bpmnEngine.NewTaskHandler().Id("id").Handler(cp.CallPathHandler)

	// when
	instance, err := bpmnEngine.CreateInstance(pi.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance.GetState(), is.EqualTo(bpmn_engine.Ready))
	bytes := bpmnEngine.Marshal()

	if EnableJsonDataDump {
		os.WriteFile("temp.marshal.remain.json", bytes, 0644)
	}

	// when
	newEngine, err := bpmn_engine.Unmarshal(bytes)
	then.AssertThat(t, err, is.Nil())
	newEngine.NewTaskHandler().Id("id").Handler(cp.CallPathHandler)

	// then
	instance, err = newEngine.RunOrContinueInstance(instance.GetInstanceKey())
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance.GetState(), is.EqualTo(bpmn_engine.Completed))

	then.AssertThat(t, cp.CallPath, is.EqualTo("id"))
}

func Test_Marshal_Unmarshal_IntermediateCatchEvents(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New()

	// given
	pi, err := bpmnEngine.LoadFromFile("../test-cases/simple-intermediate-message-catch-event.bpmn")
	then.AssertThat(t, err, is.Nil())

	// when
	_, err = bpmnEngine.CreateAndRunInstance(pi.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())
	bytes := bpmnEngine.Marshal()
	then.AssertThat(t, len(bytes), is.GreaterThan(32))

	if EnableJsonDataDump {
		os.WriteFile("temp.marshal.intermediatecatchevent.json", bytes, 0644)
	}

	// when
	bpmnEngine, err = bpmn_engine.Unmarshal(bytes)
	then.AssertThat(t, err, is.Nil())

	// then
	subscriptions := bpmnEngine.GetMessageSubscriptions()
	then.AssertThat(t, subscriptions, has.Length(1))
}

func Test_Marshal_Unmarshal_IntermediateTimerEvents_timer_is_completing(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New()
	cp := CallPath{}

	// given
	pi, err := bpmnEngine.LoadFromFile("../test-cases/message-intermediate-timer-event.bpmn")
	then.AssertThat(t, err, is.Nil())

	// when
	instance, err := bpmnEngine.CreateAndRunInstance(pi.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())
	bytes := bpmnEngine.Marshal()
	then.AssertThat(t, len(bytes), is.GreaterThan(32))

	if EnableJsonDataDump {
		os.WriteFile("temp.marshal.message-intermediate-timer-event.json", bytes, 0644)
	}

	// when
	bpmnEngine, err = bpmn_engine.Unmarshal(bytes)
	then.AssertThat(t, err, is.Nil())

	// then
	timers := bpmnEngine.GetTimersScheduled()
	then.AssertThat(t, timers, has.Length(1))
	pii := bpmnEngine.FindProcessInstance(instance.InstanceKey)
	then.AssertThat(t, pii, is.Not(is.Nil()))

	// when
	bpmnEngine.NewTaskHandler().Id("task-for-timer").Handler(cp.CallPathHandler)
	bpmnEngine.NewTaskHandler().Id("task-for-message").Handler(cp.CallPathHandler)
	time.Sleep(1 * time.Second)
	pii, err = bpmnEngine.RunOrContinueInstance(pii.InstanceKey)
	then.AssertThat(t, pii, is.Not(is.Nil()))
	then.AssertThat(t, pii.State, is.EqualTo(bpmn_engine.Completed))
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-for-timer"))
}

func Test_Marshal_Unmarshal_IntermediateTimerEvents_message_is_completing(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New()
	cp := CallPath{}

	// given
	pi, err := bpmnEngine.LoadFromFile("../test-cases/message-intermediate-timer-event.bpmn")
	then.AssertThat(t, err, is.Nil())

	// when
	instance, err := bpmnEngine.CreateAndRunInstance(pi.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())
	bytes := bpmnEngine.Marshal()
	then.AssertThat(t, len(bytes), is.GreaterThan(32))

	// when
	bpmnEngine, err = bpmn_engine.Unmarshal(bytes)
	then.AssertThat(t, err, is.Nil())

	// then
	subscriptions := bpmnEngine.GetMessageSubscriptions()
	then.AssertThat(t, subscriptions, has.Length(1))
	pii := bpmnEngine.FindProcessInstance(instance.InstanceKey)
	then.AssertThat(t, pii, is.Not(is.Nil()))

	// when
	bpmnEngine.NewTaskHandler().Id("task-for-timer").Handler(cp.CallPathHandler)
	bpmnEngine.NewTaskHandler().Id("task-for-message").Handler(cp.CallPathHandler)
	err = bpmnEngine.PublishEventForInstance(pii.InstanceKey, "message", nil)
	then.AssertThat(t, err, is.Nil())
	pii, err = bpmnEngine.RunOrContinueInstance(pii.InstanceKey)
	then.AssertThat(t, pii, is.Not(is.Nil()))
	then.AssertThat(t, pii.State, is.EqualTo(bpmn_engine.Completed))
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-for-message"))
}

func Test_Unmarshal_restores_processKey(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New()

	// given
	piBefore, err := bpmnEngine.LoadFromFile("../test-cases/simple_task.bpmn")
	then.AssertThat(t, err, is.Nil())

	// when
	bytes := bpmnEngine.Marshal()

	// when
	bpmnEngine, err = bpmn_engine.Unmarshal(bytes)
	then.AssertThat(t, err, is.Nil())
	processes := bpmnEngine.FindProcessesById("Simple_Task_Process")

	// then
	then.AssertThat(t, processes, has.Length(1))
	then.AssertThat(t, processes[0].ProcessKey, is.EqualTo(piBefore.ProcessKey))
}
