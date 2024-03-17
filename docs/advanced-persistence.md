
## Persistence

The lib-bpmn-engine supports persistence (a.k.a. marshalling or serialization),
which can be used to pause workflows, store them on disk and resume later.
The data format is plain JSON, which you as the user of the lib must store and load.
By design, the lib will not support any specific database technology.

When calling `bpmnEngine.Marshal()`, the whole engine including all process instances is exported.
When you have a large amount of process instances, it's recommended to rather use multiple
engine instances, one per process instance, to keep the exported data small and efficient.

#### Example

For this example, we're just using a simple human task, which is supposed to be stored on disk.

![simple-user-task_bpmn](./examples/persistence/simple-user-task.png)


<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/persistence/persistence.go) -->
<!-- The below code snippet is automatically added from ./examples/persistence/persistence.go -->
```go
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

	// debug print ... in real-live the data is stored to disk/database
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
```
<!-- MARKDOWN-AUTO-DOCS:END -->

To get the snippet compile, see the full sources in the
[./examples/persistence/](./examples/persistence/) folder.
