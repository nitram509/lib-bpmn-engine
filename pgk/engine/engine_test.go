package engine

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
	bpmnEngine.AddHandler("Activity_1yyow37", handler)

	// when
	bpmnEngine.Execute()

	then.AssertThat(t, wasCalled, is.True())
}
