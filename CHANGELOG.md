
# CHANGELOG lib-bpmn-engine

## v0.2.4

* support input/output for service task and user task (#2)
   * breaking change: ```ActivatedJob``` type is no more using fields, but only function interface
* support for user tasks (BPMN) (#32)
* document how to use timers (#37)
* support adding variables along with publishing messages (#41)
   * breaking change in method signature: ```PublishEventForInstance(processInstanceKey int64, messageName string, variables map[string]interface{})``` now requires a variable parameter
* fix two issues with not finding/handling the correct messages (#31)

----
