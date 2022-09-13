package main

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
	"time"
)

func main() {
	bpmnEngine := bpmn_engine.New("a name")
	process, err := bpmnEngine.LoadFromFile("timeout-example.bpmn")
	if err != nil {
		panic("file \"timeout-example.bpmn\" can't be read.")
	}
	// just some dummy handler to complete the tasks/jobs
	registerDummyTaskHandlers(bpmnEngine)

	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	println(instance.GetState()) // still ACTIVE at this point

	printScheduledTimerInformation(bpmnEngine.GetTimersScheduled()[0])

	// sleep() for 2 seconds, before trying to continue the process instance
	// this for-loop essentially will block until the process instance has completed OR an error occurred
	for ; instance.GetState() == process_instance.ACTIVE && err == nil; time.Sleep(2 * time.Second) {
		println("tick.")
		// by re-running, the engine will check for active timers and might continue execution,
		// if timer.DueAt has passed
		_, err = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())
	}

	println(instance.GetState()) // finally completed
}
