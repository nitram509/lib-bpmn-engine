package bpmn_engine

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func TestRegisteredHandlerGetsCalled(t *testing.T) {
	// setup
	bpmnEngine := New()
	simpleTask := "simple_task"
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml", simpleTask)
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
	bpmnEngine := New()
	simpleTask := "simple_task"
	variableName := "variable_name"
	taskId := "Activity_1yyow37"
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml", simpleTask)
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
	bpmnEngine := New()
	fileName := "../../test-cases/simple_task.xml"
	metadata, _ := bpmnEngine.LoadFromFile(fileName, "simple_task")

	then.AssertThat(t, metadata.Version, is.EqualTo(int32(1)))
	then.AssertThat(t, metadata.ProcessKey, is.GreaterThan(1))
	then.AssertThat(t, metadata.ResourceName, is.EqualTo("simple_task"))
	then.AssertThat(t, metadata.BpmnProcessId, is.EqualTo("Simple_Task_Process"))
}

func TestLoadingTheSameFileWillNotIncreaseTheVersionNorChangeTheProcessKey(t *testing.T) {
	// setup
	bpmnEngine := New()
	simpleTask := "simple_task"

	metadata, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml", simpleTask)
	keyOne := metadata.ProcessKey
	then.AssertThat(t, metadata.Version, is.EqualTo(int32(1)))

	metadata, _ = bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml", simpleTask)
	keyTwo := metadata.ProcessKey
	then.AssertThat(t, metadata.Version, is.EqualTo(int32(1)))

	then.AssertThat(t, keyOne, is.EqualTo(keyTwo))
}

func TestLoadingTheSameProcessWithModificationWillCreateNewVersion(t *testing.T) {
	// setup
	bpmnEngine := New()
	simpleTask := "simple_task"
	simpleTask2 := "simple_task_2"

	process1, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml", simpleTask)
	process2, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task_modified_taskId.xml", simpleTask)
	process3, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml", simpleTask2)

	then.AssertThat(t, process1.BpmnProcessId, is.EqualTo(process2.BpmnProcessId).Reason("both prepared files should have equal IDs"))
	then.AssertThat(t, process2.ProcessKey, is.GreaterThan(process1.ProcessKey).Reason("Because later created"))
	then.AssertThat(t, process3.ProcessKey, is.EqualTo(process1.ProcessKey).Reason("Same processKey return for same input file, means already registered"))

	then.AssertThat(t, process1.Version, is.EqualTo(int32(1)))
	then.AssertThat(t, process2.Version, is.EqualTo(int32(2)))
	then.AssertThat(t, process3.Version, is.EqualTo(int32(1)))

	then.AssertThat(t, process1.ProcessKey, is.Not(is.EqualTo(process2.ProcessKey)))
	then.AssertThat(t, process1.ResourceName, is.EqualTo(process2.ResourceName))
}
