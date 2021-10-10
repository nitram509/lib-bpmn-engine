package bpmn_engine

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func TestRegisteredHandlerGetsCalled(t *testing.T) {
	// setup
	bpmnEngine := BpmnEngineState{
		states: map[string]*BpmnEngineNamedResourceState{},
	}
	simpleTask := "simple_task"
	bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml", simpleTask)
	var wasCalled = false
	handler := func(id string) {
		wasCalled = true
	}

	// given
	bpmnEngine.AddTaskHandler(simpleTask, "Activity_1yyow37", handler)

	// when
	bpmnEngine.Execute(simpleTask)

	then.AssertThat(t, wasCalled, is.True())
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

	metadata1, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml", simpleTask)
	metadata2, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task_modified_taskId.xml", simpleTask)
	metadata3, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml", simpleTask2)

	then.AssertThat(t, metadata1.BpmnProcessId, is.EqualTo(metadata2.BpmnProcessId))
	then.AssertThat(t, metadata2.ProcessKey, is.GreaterThan(metadata1.ProcessKey))
	then.AssertThat(t, metadata3.ProcessKey, is.GreaterThan(metadata2.ProcessKey))

	then.AssertThat(t, metadata1.Version, is.EqualTo(int32(1)))
	then.AssertThat(t, metadata2.Version, is.EqualTo(int32(2)))
	then.AssertThat(t, metadata3.Version, is.EqualTo(int32(1)))

	then.AssertThat(t, metadata1.ProcessKey, is.Not(is.EqualTo(metadata2.ProcessKey)))
	then.AssertThat(t, metadata1.ResourceName, is.EqualTo(metadata2.ResourceName))
}
