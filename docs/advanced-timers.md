
## Timers

The lib-bpmn-engine supports *timer intermediate catch events*,
which are very useful to model typical timeout scenarios.

![timeout example bpmn](./examples/timers/timeout-example.png)

The one above "ask $1 million question" demonstrates a 10 seconds timeout
to give the correct answer or lose the whole "game".
This is a best-practice example, for how to model (business) timeouts or deadlines.

### How to model timer/scheduler?

In BPMN processes, Timer Intermediate Catch events can be used in combination with
Event Based Gateway, to exclusively select one execution path in the process.
When a timer event happens before a message event, then the example $1 million question game is lost. 

### Architectural choices for timer/scheduler

**The Problem:** implementing a timer/scheduler very much depends on your context or non-functional requirements.
E.g. you might run lib-bpmn-engine as part of a single batch job instance OR you have a web service 
implement which is running with 3 instances. Both scenarios do require different implementation approaches,
how to deal with long-running processes.

**Choices:** Depending on your scenario/use case, you might implement a trivial blocking loop,
like in the example code below. 
In multi-instance environments, you might better use a central scheduler, to avoid each instance of the
application (using lib-bpmn-engine) is doing its own un-coordinated timing/scheduling.

##### lib-bpmn-engine design

* lib-bpmn-engine is not aware of how it's deployed
* lib-bpmn-engine will not block, when timers are set == this is like an implicit pause
* lib-bpmn-engine delegates scheduler/timer responsibility to the developer (==you)

**In a nutshell:** 
The lib-bpmn-engine does create such timer event objects and will pause the process execution.
This means, an external ticker/scheduler is required, to continue the process instance.


### Trivial example, blocking local execution

The code snippet below demonstrates a trivial example, how to execute processes with timers.
Here, the execution is blocking until the due time is reached.
This might fit in a scenario, where you have a single instance running in a batch-job like environment.

Depending on your context, you might choose some external ticker/scheduler,
to check for active scheduled timer events.

<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/timers/timers.go) -->
<!-- The below code snippet is automatically added from ./examples/timers/timers.go -->
```go
package main

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"time"
)

func main() {
	bpmnEngine := bpmn_engine.New()
	process, err := bpmnEngine.LoadFromFile("timeout-example.bpmn")
	if err != nil {
		panic("file \"timeout-example.bpmn\" can't be read.")
	}
	// just some dummy handler to complete the tasks/jobs
	registerDummyTaskHandlers(&bpmnEngine)

	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	println(instance.GetState()) // still ACTIVE at this point

	printScheduledTimerInformation(bpmnEngine.GetTimersScheduled()[0])

	// sleep() for 2 seconds, before trying to continue the process instance
	// this for-loop essentially will block until the process instance has completed OR an error occurred
	for ; instance.GetState() == bpmn_engine.Active && err == nil; time.Sleep(2 * time.Second) {
		println("tick.")
		// by re-running, the engine will check for active timers and might continue execution,
		// if timer.DueAt has passed
		_, err = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())
	}

	println(instance.GetState()) // finally completed
}
```
<!-- MARKDOWN-AUTO-DOCS:END -->

To get the snippet compile, see the full sources in the
[examples/timers/](./examples/timers/) folder.
