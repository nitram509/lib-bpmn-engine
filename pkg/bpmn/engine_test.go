package bpmn

import (
	"log"
	"testing"
	"time"

	"github.com/corbym/gocrest/has"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

type CallPath struct {
	CallPath string
}

func (callPath *CallPath) TaskHandler(job ActivatedJob) {
	if len(callPath.CallPath) > 0 {
		callPath.CallPath += ","
	}
	callPath.CallPath += job.ElementId()
	job.Complete()
}

func Test_BpmnEngine_interfaces_implemented(t *testing.T) {
	var _ BpmnEngine = &BpmnEngineState{}
}

func TestRegisterHandlerByTaskIdGetsCalled(t *testing.T) {
	// setup
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")
	wasCalled := false
	handler := func(job ActivatedJob) {
		wasCalled = true
		job.Complete()
	}

	// given
	bpmnEngine.NewTaskHandler().Id("id").Handler(handler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, wasCalled, is.True())
}

func TestRegisterHandlerByTaskIdGetsCalledAfterLateRegister(t *testing.T) {
	// setup
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")
	wasCalled := false
	handler := func(job ActivatedJob) {
		wasCalled = true
		job.Complete()
	}

	// // given
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	if err != nil {
		log.Fatal(err)
	}
	bpmnEngine.NewTaskHandler().Id("id").Handler(handler)
	bpmnEngine.RunOrContinueInstance(pi.InstanceKey)

	// when
	then.AssertThat(t, wasCalled, is.True())
}

func TestRegisteredHandlerCanMutateVariableContext(t *testing.T) {
	// setup
	bpmnEngine := New()
	variableName := "variable_name"
	taskId := "id"
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")
	variableContext := make(map[string]interface{}, 1)
	variableContext[variableName] = "oldVal"

	handler := func(job ActivatedJob) {
		v := job.Variable(variableName)
		then.AssertThat(t, v, is.EqualTo("oldVal").Reason("one should be able to read variables"))
		job.SetVariable(variableName, "newVal")
		job.Complete()
	}

	// given
	bpmnEngine.NewTaskHandler().Id(taskId).Handler(handler)

	// when
	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, variableContext)

	v := bpmnEngine.persistence.FindProcessInstanceByKey(instance.GetInstanceKey())

	// then
	then.AssertThat(t, v.VariableHolder.GetVariable(variableName), is.EqualTo("newVal"))
}

func TestMetadataIsGivenFromLoadedXmlFile(t *testing.T) {
	// setup
	bpmnEngine := New()
	metadata, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")

	then.AssertThat(t, metadata.Version, is.EqualTo(int32(1)))
	then.AssertThat(t, metadata.ProcessKey, is.GreaterThan(1))
	then.AssertThat(t, metadata.BpmnProcessId, is.EqualTo("Simple_Task_Process"))
}

func TestLoadingTheSameFileWillNotIncreaseTheVersionNorChangeTheProcessKey(t *testing.T) {
	// setup
	bpmnEngine := New()

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
	bpmnEngine := New()

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
	bpmnEngine := New()

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")

	// when
	instance1, _ := bpmnEngine.CreateInstance(process, nil)
	instance2, _ := bpmnEngine.CreateInstance(process, nil)

	// then
	then.AssertThat(t, instance1.CreatedAt.UnixNano(), is.GreaterThanOrEqualTo(beforeCreation.UnixNano()).Reason("make sure we have creation time set"))
	then.AssertThat(t, instance1.ProcessInfo.ProcessKey, is.EqualTo(instance2.ProcessInfo.ProcessKey))
	then.AssertThat(t, instance2.InstanceKey, is.GreaterThan(instance1.InstanceKey).Reason("Because later created"))
}

func TestSimpleAndUncontrolledForkingTwoTasks(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/forked-flow.bpmn")
	bpmnEngine.NewTaskHandler().Id("id-a-1").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("id-b-1").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("id-b-2").Handler(cp.TaskHandler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("id-a-1,id-b-1,id-b-2"))
}

func TestParallelGateWayTwoTasks(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/parallel-gateway-flow.bpmn")
	bpmnEngine.NewTaskHandler().Id("id-a-1").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("id-b-1").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("id-b-2").Handler(cp.TaskHandler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("id-a-1,id-b-1,id-b-2"))
}

func TestMultipleEnginesCanBeCreatedWithoutAName(t *testing.T) {
	// when
	bpmnEngine1 := New()
	bpmnEngine2 := New()

	// then
	then.AssertThat(t, bpmnEngine1.name, is.Not(is.EqualTo(bpmnEngine2.name).Reason("make sure the names are different")))
}

func Test_multiple_engines_create_unique_Ids(t *testing.T) {
	// setup
	bpmnEngine1 := New()
	bpmnEngine2 := New()

	// when
	process1, _ := bpmnEngine1.LoadFromFile("../../test-cases/simple_task.bpmn")
	process2, _ := bpmnEngine2.LoadFromFile("../../test-cases/simple_task.bpmn")

	// then
	then.AssertThat(t, process1.ProcessKey, is.Not(is.EqualTo(process2.ProcessKey)))
}

func Test_CreateInstanceById_uses_latest_process_version(t *testing.T) {
	// setup
	engine := New()

	// when
	v1, err := engine.LoadFromFile("../../test-cases/simple_task.bpmn")
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, v1.definitions.Process.Name, is.EqualTo("aName"))
	// when
	v2, err := engine.LoadFromFile("../../test-cases/simple_task_v2.bpmn")
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, v2.definitions.Process.Name, is.EqualTo("aName"))

	instance, err := engine.CreateInstanceById("Simple_Task_Process", nil)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance, is.Not(is.Nil()))

	// ten
	then.AssertThat(t, instance.ProcessInfo.Version, is.EqualTo(int32(v2.Version)))
}

func Test_CreateAndRunInstanceById_uses_latest_process_version(t *testing.T) {
	// setup
	engine := New()

	// when
	v1, err := engine.LoadFromFile("../../test-cases/simple_task.bpmn")
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, v1.definitions.Process.Name, is.EqualTo("aName"))
	// when
	v2, err := engine.LoadFromFile("../../test-cases/simple_task_v2.bpmn")
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, v2.definitions.Process.Name, is.EqualTo("aName"))

	instance, err := engine.CreateAndRunInstanceById("Simple_Task_Process", nil)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance, is.Not(is.Nil()))

	// then
	then.AssertThat(t, instance.ProcessInfo.Version, is.EqualTo(int32(v2.Version)))
}

func Test_CreateInstanceById_return_error_when_no_ID_found(t *testing.T) {
	// setup
	engine := New()

	// when
	instance, err := engine.CreateInstanceById("Simple_Task_Process_not_existing", nil)

	// then
	then.AssertThat(t, instance, is.Nil())
	then.AssertThat(t, err, is.Not(is.Nil()))
	then.AssertThat(t, err.Error(), has.Prefix("no process with id=Simple_Task_Process_not_existing was found (prior loaded into the engine)"))
}
