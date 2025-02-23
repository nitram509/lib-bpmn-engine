package main

import (
	"fmt"
	"time"

	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
)

func initBpmnEngine() {
	bpmnEngine = bpmn_engine.New()
	process, _ = bpmnEngine.LoadFromBytes(OrderingItemsWorkflowBpmn)
	bpmnEngine.NewTaskHandler().Id("validate-order").Handler(printHandler)
	bpmnEngine.NewTaskHandler().Id("send-bill").Handler(printHandler)
	bpmnEngine.NewTaskHandler().Id("send-friendly-reminder").Handler(printHandler)
	bpmnEngine.NewTaskHandler().Id("update-accounting").Handler(updateAccountingHandler)
	bpmnEngine.NewTaskHandler().Id("package-and-deliver").Handler(printHandler)
	bpmnEngine.NewTaskHandler().Id("send-cancellation").Handler(printHandler)
}

func printHandler(job bpmn_engine.ActivatedJob) {
	// do important stuff here
	println(fmt.Sprintf("%s >>> Executing job '%s'", time.Now(), job.ElementId()))
	job.Complete()
}

func updateAccountingHandler(job bpmn_engine.ActivatedJob) {
	println(fmt.Sprintf("%s >>> Executing job '%s'", time.Now(), job.ElementId()))
	println(fmt.Sprintf("%s >>> update ledger revenue account with amount=%s", time.Now(), job.Variable("amount")))
	job.Complete()
}
