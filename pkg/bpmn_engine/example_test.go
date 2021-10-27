package bpmn_engine_test

import (
	"fmt"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
)

func ExampleNew() {
	// basic example loading a BPMN from file,
	// register a handler for a service task by Id
	// and execute the process
	bpmnEngine := bpmn_engine.New()
	simpleTask := "simple_task"
	bpmnEngine.LoadFromFile("test.bpmn.xml", simpleTask)
	bpmnEngine.AddTaskHandler(simpleTask, "aTaskId", myHandlerGenerator(bpmnEngine))
	bpmnEngine.Execute(simpleTask)
}

func myHandlerGenerator(state bpmn_engine.BpmnEngineState) func(id string) {
	return func(id string) {
		println("Executing task id=" + id)

		fmt.Printf("Variable context for task: %v",
			state.GetProcesses("simple_task")[0].VariableContext)
	}
}
