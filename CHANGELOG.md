
# CHANGELOG lib-bpmn-engine

## v0.3.0-beta1

- support handlers being registered for (task definition) types (#58 BREAKING CHANGE)
- support handlers for user tasks being registered for assignee or candidate groups (#59)

## Migration notes for breaking changes

- replace ```AddTaskHandler("id", handlerFunc)``` with ```NewTaskHandler.Id("id").Handler(handlerFunc)```

----

## v0.2.4

* support input/output for service task and user task (#2)
   * breaking change: ```ActivatedJob``` type is no more using fields, but only function interface
* support for user tasks (BPMN) (#32)
* document how to use timers (#37)
* support adding variables along with publishing messages (#41)
   * breaking change in method signature: ```PublishEventForInstance(processInstanceKey int64, messageName string, variables map[string]interface{})``` now requires a variable parameter
* fix two issues with not finding/handling the correct messages (#31)

----
