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

	then.AssertThat(t, metadata.version, is.EqualTo(int32(1)))
	then.AssertThat(t, metadata.processKey, is.GreaterThan(1))
	then.AssertThat(t, metadata.resourceName, is.EqualTo(fileName))
	then.AssertThat(t, metadata.bpmnProcessId, is.EqualTo("Simple_Task_Process"))
}

func TestLoadingTheSameFileWillNotIncreaseTheVersionNorChangeTheProcessKey(t *testing.T) {
	// setup
	bpmnEngine := New()

	metadata, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml")
	keyOne := metadata.processKey
	then.AssertThat(t, metadata.version, is.EqualTo(int32(1)))

	metadata, _ = bpmnEngine.LoadFromFile("../../test-cases/simple_task.xml")
	keyTwo := metadata.processKey
	then.AssertThat(t, metadata.version, is.EqualTo(int32(1)))

	then.AssertThat(t, keyOne, is.EqualTo(keyTwo))
}
