package main

import (
	"fmt"

	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
)

func main() {
	// create a new named engine
	bpmnEngine := bpmn_engine.New()
	// basic example loading a BPMN from file,
	process, err := bpmnEngine.LoadFromFile("simple_task.bpmn")
	if err != nil {
		panic("file \"simple_task.bpmn\" can't be read.")
	}
	// register a handler for a service task by defined task id (there are alternative options for types as well)
	bpmnEngine.NewTaskHandler().Id("hello-world").Handler(printContextHandler)
	// setup some variables
	variables := map[string]interface{}{}
	variables["foo"] = "bar"
	// and execute the process
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, variables)
}

func printContextHandler(job bpmn_engine.ActivatedJob) {
	println("< Hello World >")
	println(fmt.Sprintf("ElementId                = %s", job.ElementId()))
	println(fmt.Sprintf("BpmnProcessId            = %s", job.BpmnProcessId()))
	println(fmt.Sprintf("ProcessDefinitionKey     = %d", job.ProcessDefinitionKey()))
	println(fmt.Sprintf("ProcessDefinitionVersion = %d", job.ProcessDefinitionVersion()))
	println(fmt.Sprintf("CreatedAt                = %s", job.CreatedAt()))
	println(fmt.Sprintf("Variable 'foo'           = %s", job.Variable("foo")))
	job.Complete() // don't forget this one, or job.Fail("foobar")
}
