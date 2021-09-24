package bpmn_engine_test

import "github.com/nitram509/golib-bpmn-model/pgk/bpmn_engine"

func ExampleNew() {
	// basic example loading a BPMN from file,
	// register a handler for a service task by Id
	// and execute the process
	bpmnEngine := bpmn_engine.New()
	bpmnEngine.LoadFromFile("test.bpmn.xml")
	bpmnEngine.AddTaskHandler("aTaskId", myHandler)
	bpmnEngine.Execute()
}

func myHandler(id string) {

}
