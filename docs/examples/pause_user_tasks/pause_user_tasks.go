package main

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
)

func main() {
	bpmnEngine := bpmn_engine.New("name")
	process, _ := bpmnEngine.LoadFromFile("simple-user-task.bpmn")
	bpmnEngine.NewTaskHandler().Assignee("assignee").Handler(userTaskHandler())
	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	// ... just wait for the human completed his/her task
	bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())
}

var externalEvent = "none"

func userTaskHandler() func(job bpmn_engine.ActivatedJob) {
	return func(job bpmn_engine.ActivatedJob) {
		if externalEvent == "none" {
			// send notification to user
		}
		if externalEvent == "user is done" {
			job.Complete()
		}
		if externalEvent == "user is done but wrong response" {
			job.Fail("error in user task")
		}
		// just return and so 'pause' the process instance
	}
}
