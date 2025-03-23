package main

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
)

func main() {
	// create a new named engine
	bpmnEngine := bpmn_engine.New()
	// basic example loading a v1 BPMN from file,
	_, err := bpmnEngine.LoadFromFile("simple_task.bpmn")
	if err != nil {
		panic("file \"simple_task.bpmn\" can't be read. " + err.Error())
	}
	// now loading v2, basically with the same process ID
	_, err = bpmnEngine.LoadFromFile("simple_task_v2.bpmn")
	if err != nil {
		panic("file \"simple_task.bpmn\" can't be read. " + err.Error())
	}

	// register a handler for a service task by defined task type
	bpmnEngine.NewTaskHandler().Type("hello-world").Handler(printElementIdHandler)
	// and execute the process, means we will use v2
	bpmnEngine.CreateAndRunInstanceById("hello-world-process-id", nil)
}

func printElementIdHandler(job bpmn_engine.ActivatedJob) {
	println(job.ElementId())
	job.Complete() // don't forget this one, or job.Fail("foobar")
}
