package example_test

import (
	"fmt"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
)

func ExampleNew() {
	// create a new named engine
	bpmnEngine := bpmn_engine.New("a name")
	// basic example loading a BPMN from file,
	process, _ := bpmnEngine.LoadFromFile("test.bpmn.xml")
	// register a handler for a service task by Id
	bpmnEngine.AddTaskHandler("aTaskId", myHandlerGenerator(bpmnEngine))
	// and execute the process
	bpmnEngine.CreateAndRunInstance(process.ProcessKey)
}

func myHandlerGenerator(state bpmn_engine.BpmnEngineState) func(id string) {
	return func(id string) {
		println("Executing task id=" + id)
		fmt.Printf("Variable context for task: %v",
			state.GetProcessInstances()[0].VariableContext)
	}
}
