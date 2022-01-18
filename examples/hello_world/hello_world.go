// Provides a simple hello world example, which just executes a single service task
// and prints its context variables.
package main

import "github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"

func main() {
	// create a new named engine
	bpmnEngine := bpmn_engine.New("a name")
	// basic example loading a BPMN from file,
	process, err := bpmnEngine.LoadFromFile("simple_task.bpmn")
	if err != nil {
		panic("file \"simple_task.bpmn\" can't be read.")
	}
	// register a handler for a service task by defined task type
	bpmnEngine.AddTaskHandler("hello-world", printContextHandler)
	// setup some variables
	variables := map[string]string{}
	variables["foo"] = "bar"
	// and execute the process
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, variables)
}

func printContextHandler(context bpmn_engine.ProcessInstanceContext) {
	println("Hello World")
	println("Executing: TaskId=" + context.GetTaskId())
	println("Variable:  foo=" + context.GetVariable("foo"))
}
