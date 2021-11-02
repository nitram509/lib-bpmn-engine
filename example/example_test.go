package example_test

import (
	"fmt"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
)

func ExampleNew() {
	// basic example loading a BPMN from file,
	// register a handler for a service task by Id
	// and execute the process
	bpmnEngine := bpmn_engine.New("a name")
	process, _ := bpmnEngine.LoadFromFile("test.bpmn.xml")
	bpmnEngine.AddTaskHandler("aTaskId", myHandlerGenerator(bpmnEngine))
	bpmnEngine.CreateAndRunInstance(process.ProcessKey)
}

func myHandlerGenerator(state bpmn_engine.BpmnEngineState) func(id string) {
	return func(id string) {
		println("Executing task id=" + id)

		fmt.Printf("Variable context for task: %v",
			state.GetProcessInstances()[0].VariableContext)
	}
}
