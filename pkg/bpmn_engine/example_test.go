package bpmn_engine_test

import "github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"

func ExampleNew() {
	// basic example loading a BPMN from file,
	// register a handler for a service task by Id
	// and execute the process
	bpmnEngine := bpmn_engine.New()
	simpleTask := "simple_task"
	bpmnEngine.LoadFromFile("test.bpmn.xml", simpleTask)
	bpmnEngine.AddTaskHandler(simpleTask, "aTaskId", myHandler)
	bpmnEngine.Execute(simpleTask)
}

func myHandler(id string) {
	println("Executing task id=" + id)
}
