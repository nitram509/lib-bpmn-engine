package example_test

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
)

func ExampleNew() {
	// create a new named engine
	bpmnEngine := bpmn_engine.New("a name")
	// basic example loading a BPMN from file,
	process, _ := bpmnEngine.LoadFromFile("test.bpmn")
	// register a handler for a service task by Id
	bpmnEngine.AddTaskHandler("aTaskId", myHandler)
	// and execute the process
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
}

func myHandler(context bpmn_engine.ProcessInstanceContext) {
	println("Executing task id=" + context.GetTaskId())
	println("Variable foo=" + context.GetVariable("foo"))
}
