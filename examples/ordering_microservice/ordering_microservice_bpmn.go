package main

import (
	"fmt"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"time"
)

func initBpmnEngine() {
	bpmnEngine = bpmn_engine.New("Ordering-Microservice")
	process, _ = bpmnEngine.LoadFromBytes(OrderingItemsWorkflowBpmn)
	bpmnEngine.AddTaskHandler("validate-order", printHandler)
	bpmnEngine.AddTaskHandler("send-bill", printHandler)
	bpmnEngine.AddTaskHandler("send-friendly-reminder", printHandler)
	bpmnEngine.AddTaskHandler("update-accounting", updateAccountingHandler)
	bpmnEngine.AddTaskHandler("package-and-deliver", printHandler)
	bpmnEngine.AddTaskHandler("send-cancellation", printHandler)
}

func printHandler(job bpmn_engine.ActivatedJob) {
	// do important stuff here
	println(fmt.Sprintf("%s >>> Executing job '%s'", time.Now(), job.ElementId))
	job.Complete()
}

func updateAccountingHandler(job bpmn_engine.ActivatedJob) {
	println(fmt.Sprintf("%s >>> Executing job '%s'", time.Now(), job.ElementId))
	println(fmt.Sprintf("%s >>> update ledger revenue account with amount=%s", time.Now(), job.GetVariable("amount")))
	job.Complete()
}
