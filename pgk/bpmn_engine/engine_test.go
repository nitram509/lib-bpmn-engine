package bpmn_engine

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func TestRegisteredHandlerGetsCalled(t *testing.T) {
	// setup
	bpmnEngine := BpmnEngineState{}
	bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml")
	var wasCalled = false
	handler := func(id string) {
		wasCalled = true
	}

	// given
	bpmnEngine.AddTaskHandler("Activity_1yyow37", handler)

	// when
	bpmnEngine.Execute()

	then.AssertThat(t, wasCalled, is.True())
}

func TestMetadataIsGivenFromLoadedXmlFile(t *testing.T) {
	// setup
	bpmnEngine := New()
	fileName := "../../test-cases/simple_task.xml"
	metadata, _ := bpmnEngine.LoadFromFile(fileName)

	then.AssertThat(t, metadata.Version, is.EqualTo(int32(1)))
	then.AssertThat(t, metadata.ProcessKey, is.GreaterThan(1))
	then.AssertThat(t, metadata.ResourceName, is.EqualTo(fileName))
	then.AssertThat(t, metadata.BpmnProcessId, is.EqualTo("Simple_Task_Process"))
}

func TestLoadingTheSameFileWillNotIncreaseTheVersionNorChangeTheProcessKey(t *testing.T) {
	// setup
	bpmnEngine := New()

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
	bpmnEngine := New()

	metadata1, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml")
	metadata2, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task_modified_taskId.xml")

	then.AssertThat(t, metadata1.BpmnProcessId, is.EqualTo(metadata2.BpmnProcessId))

	then.AssertThat(t, metadata1.Version, is.EqualTo(int32(1)))
	then.AssertThat(t, metadata2.Version, is.EqualTo(int32(2)))

	then.AssertThat(t, metadata1.ProcessKey, is.Not(is.EqualTo(metadata2.ProcessKey)))
	then.AssertThat(t, metadata1.ResourceName, is.Not(is.EqualTo(metadata2.ResourceName)))
}
