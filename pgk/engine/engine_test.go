package engine

import (
	"testing"
)

func TestStartProcess(t *testing.T) {
	core := BpmnEngineState{}
	core.LoadFromFile("../../test/simple_task.xml")
	core.Execute()
}
