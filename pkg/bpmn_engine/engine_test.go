package bpmn_engine

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
	"time"
)

func TestRegisteredHandlerGetsCalled(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml")
	var wasCalled = false
	handler := func(id string) {
		wasCalled = true
	}

	// given
	bpmnEngine.AddTaskHandler("Activity_1yyow37", handler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey)

	then.AssertThat(t, wasCalled, is.True())
}

func TestRegisteredHandlerCanMutateVariableContext(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	variableName := "variable_name"
	taskId := "Activity_1yyow37"
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml")
	bpmnEngine.CreateInstance(process.ProcessKey)
	bpmnEngine.GetProcessInstances()[0].VariableContext[variableName] = 3

	var wasCalled = false

	handler := func(id string) {
		md := bpmnEngine.GetProcessInstances()
		md[0].VariableContext[variableName] = md[0].VariableContext[variableName].(int) + 2
		wasCalled = true
	}

	// given
	bpmnEngine.AddTaskHandler(taskId, handler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey)

	then.AssertThat(t, wasCalled, is.True())
	then.AssertThat(t, bpmnEngine.processInstances[0].VariableContext[variableName], is.EqualTo(5))
}

func TestMetadataIsGivenFromLoadedXmlFile(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	metadata, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml")

	then.AssertThat(t, metadata.Version, is.EqualTo(int32(1)))
	then.AssertThat(t, metadata.ProcessKey, is.GreaterThan(1))
	then.AssertThat(t, metadata.BpmnProcessId, is.EqualTo("Simple_Task_Process"))
}

func TestLoadingTheSameFileWillNotIncreaseTheVersionNorChangeTheProcessKey(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	metadata, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml")
	keyOne := metadata.ProcessKey
	then.AssertThat(t, metadata.Version, is.EqualTo(int32(1)))

	metadata, _ = bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml")
	keyTwo := metadata.ProcessKey
	then.AssertThat(t, metadata.Version, is.EqualTo(int32(1)))

	then.AssertThat(t, keyOne, is.EqualTo(keyTwo))
}

func TestLoadingTheSameProcessWithModificationWillCreateNewVersion(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	process1, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml")
	process2, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task_modified_taskId.xml")
	process3, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml")

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
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml")

	// when
	instance1, _ := bpmnEngine.CreateInstance(process.ProcessKey)
	instance2, _ := bpmnEngine.CreateInstance(process.ProcessKey)

	// then
	then.AssertThat(t, instance1.createdAt.UnixNano(), is.GreaterThanOrEqualTo(beforeCreation.UnixNano()).Reason("make sure we have creation time set"))
	then.AssertThat(t, instance1.processInfo.ProcessKey, is.EqualTo(instance2.processInfo.ProcessKey))
	then.AssertThat(t, instance2.InstanceKey, is.GreaterThan(instance1.InstanceKey).Reason("Because later created"))
}
