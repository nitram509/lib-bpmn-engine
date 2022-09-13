package main

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
)

func registerDummyTaskHandlers(bpmnEngine bpmn_engine.BpmnEngineState) {
	var justCompleteHandler = func(job bpmn_engine.ActivatedJob) {
		job.Complete()
	}
	bpmnEngine.AddTaskHandler("ask", justCompleteHandler)
	bpmnEngine.AddTaskHandler("win", justCompleteHandler)
	bpmnEngine.AddTaskHandler("lose", justCompleteHandler)
}
