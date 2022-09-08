package bpmn_engine

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/activity"
	"testing"
	"time"
)

type CallPath struct {
	CallPath string
}

func (callPath *CallPath) CallPathHandler(job ActivatedJob) {
	if len(callPath.CallPath) > 0 {
		callPath.CallPath += ","
	}
	callPath.CallPath += job.ElementId
	job.Complete()
}

func TestAllInterfacesImplemented(t *testing.T) {
	var _ BpmnEngine = &BpmnEngineState{}
}

func TestRegisterHandlerByTaskIdGetsCalled(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")
	var wasCalled = false
	handler := func(job ActivatedJob) {
		wasCalled = true
		job.Complete()
	}

	// given
	bpmnEngine.AddTaskHandler("id", handler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	then.AssertThat(t, wasCalled, is.True())
}

func TestRegisteredHandlerCanMutateVariableContext(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	variableName := "variable_name"
	taskId := "id"
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")
	variableContext := make(map[string]interface{}, 1)
	variableContext[variableName] = "oldVal"

	handler := func(job ActivatedJob) {
		v := job.GetVariable(variableName)
		then.AssertThat(t, v, is.EqualTo("oldVal").Reason("one should be able to read variables"))
		job.SetVariable(variableName, "newVal")
		job.Complete()
	}

	// given
	bpmnEngine.AddTaskHandler(taskId, handler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, variableContext)

	// then
	then.AssertThat(t, bpmnEngine.processInstances[0].variableContext[variableName], is.EqualTo("newVal"))
}

func TestMetadataIsGivenFromLoadedXmlFile(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	metadata, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")

	then.AssertThat(t, metadata.Version, is.EqualTo(int32(1)))
	then.AssertThat(t, metadata.ProcessKey, is.GreaterThan(1))
	then.AssertThat(t, metadata.BpmnProcessId, is.EqualTo("Simple_Task_Process"))
}

func TestLoadingTheSameFileWillNotIncreaseTheVersionNorChangeTheProcessKey(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	metadata, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")
	keyOne := metadata.ProcessKey
	then.AssertThat(t, metadata.Version, is.EqualTo(int32(1)))

	metadata, _ = bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")
	keyTwo := metadata.ProcessKey
	then.AssertThat(t, metadata.Version, is.EqualTo(int32(1)))

	then.AssertThat(t, keyOne, is.EqualTo(keyTwo))
}

func TestLoadingTheSameProcessWithModificationWillCreateNewVersion(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	process1, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")
	process2, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task_modified_taskId.bpmn")
	process3, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")

	then.AssertThat(t, process1.BpmnProcessId, is.EqualTo(process2.BpmnProcessId).Reason("both prepared files should have equal IDs"))
	then.AssertThat(t, process2.ProcessKey, is.GreaterThan(process1.ProcessKey).Reason("Because later created"))
	then.AssertThat(t, process3.ProcessKey, is.EqualTo(process1.ProcessKey).Reason("Same processKey return for same input file, means already registered"))

	then.AssertThat(t, process1.Version, is.EqualTo(int32(1)))
	then.AssertThat(t, process2.Version, is.EqualTo(int32(2)))
	then.AssertThat(t, process3.Version, is.EqualTo(int32(1)))

	then.AssertThat(t, process1.ProcessKey, is.Not(is.EqualTo(process2.ProcessKey)))
}

func TestMultipleInstancesCanBeCreated(t *testing.T) {
	// setup
	beforeCreation := time.Now()
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")

	// when
	instance1, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)
	instance2, _ := bpmnEngine.CreateInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, instance1.createdAt.UnixNano(), is.GreaterThanOrEqualTo(beforeCreation.UnixNano()).Reason("make sure we have creation time set"))
	then.AssertThat(t, instance1.processInfo.ProcessKey, is.EqualTo(instance2.processInfo.ProcessKey))
	then.AssertThat(t, instance2.instanceKey, is.GreaterThan(instance1.instanceKey).Reason("Because later created"))
}

func TestSimpleAndUncontrolledForkingTwoTasks(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/forked-flow.bpmn")
	bpmnEngine.AddTaskHandler("id-a-1", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("id-b-1", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("id-b-2", cp.CallPathHandler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("id-a-1,id-b-1,id-b-2"))
}

func TestParallelGateWayTwoTasks(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/parallel-gateway-flow.bpmn")
	bpmnEngine.AddTaskHandler("id-a-1", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("id-b-1", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("id-b-2", cp.CallPathHandler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("id-a-1,id-b-1,id-b-2"))
}

func TestTaskInputOutput(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// give
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/service-task-input-output.bpmn")
	bpmnEngine.AddTaskHandler("input-task-1", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("input-task-2", cp.CallPathHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	if err != nil {
		panic(err)
	}

	// then
	for _, job := range bpmnEngine.jobs {
		then.AssertThat(t, job.State, is.EqualTo(activity.Completed))
	}
	then.AssertThat(t, cp.CallPath, is.EqualTo("input-task-1,input-task-2"))
	then.AssertThat(t, pi.GetVariable("id"), is.EqualTo(1))
	then.AssertThat(t, pi.GetVariable("orderId"), is.EqualTo(1234))
	then.AssertThat(t, pi.GetVariable("order"), is.EqualTo(map[string]interface{}{
		"name": "order1",
		"id":   "1234",
	}))
	then.AssertThat(t, pi.GetVariable("orderName").(string), is.EqualTo("order1"))
}

func TestInvalidTaskInput(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// give
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/service-task-invalid-input.bpmn")
	bpmnEngine.AddTaskHandler("invalid-input", cp.CallPathHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	if err != nil {
		panic(err)
	}
	// then
	then.AssertThat(t, pi.GetVariable("id"), is.EqualTo(nil))
	then.AssertThat(t, cp.CallPath, is.EqualTo(""))
}

func TestInvalidTaskOutput(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// give
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/service-task-invalid-output.bpmn")
	bpmnEngine.AddTaskHandler("invalid-output", cp.CallPathHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	if err != nil {
		panic(err)
	}
	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("invalid-output"))
	then.AssertThat(t, pi.GetVariable("order"), is.EqualTo(nil))
}
