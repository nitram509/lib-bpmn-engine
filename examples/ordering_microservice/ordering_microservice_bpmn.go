package main

import (
	"fmt"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"time"
)

func initBpmnEngine() {
	bpmnEngine = bpmn_engine.New("Ordering-Microservice")
	process, _ = bpmnEngine.LoadFromBytes(OrderingItemsWorkflowBpmn)
	bpmnEngine.AddTaskHandler("validate-order", businessActionHandler)
	bpmnEngine.AddTaskHandler("send-bill", businessActionHandler)
	bpmnEngine.AddTaskHandler("send-friendly-reminder", businessActionHandler)
	bpmnEngine.AddTaskHandler("update-accounting", businessActionHandler)
	bpmnEngine.AddTaskHandler("package-and-deliver", businessActionHandler)
	bpmnEngine.AddTaskHandler("send-cancellation", businessActionHandler)
}

func businessActionHandler(job bpmn_engine.ActivatedJob) {
	// do important stuff here
	msg := fmt.Sprintf("%s >>> Executing job '%s", time.Now(), job.ElementId)
	println(msg)
}
