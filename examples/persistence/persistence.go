package main

import "github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"

func main() {
	bpmnEngine := bpmn_engine.New()
	process, _ := bpmnEngine.LoadFromFile("simple-user-task.bpmn")
	bpmnEngine.NewTaskHandler().Assignee("assignee").Handler(doNothingHandler())
	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	instanceKey := instance.InstanceKey // remember the key for later continuation

	// export the whole engine state as bytes
	// the export format is valid JSON and can be stored however you want
	bytes := bpmnEngine.Marshal()

	// debug print ...
	println(string(bytes))

	resumeWorkflow(bytes, instanceKey)
}

func resumeWorkflow(bytes []byte, processInstanceKey int64) {
	// import the bytes
	newBpmnEngine, _ := bpmn_engine.Unmarshal(bytes)

	// and resume the workflow for the give process instance key
	_, _ = newBpmnEngine.RunOrContinueInstance(processInstanceKey)
}

func doNothingHandler() func(job bpmn_engine.ActivatedJob) {
	return func(job bpmn_engine.ActivatedJob) {
		println("Do nothing, which keeps the job active.")
		// HINT: to complete a job, the handler must call `job.Complete()`
	}
}
