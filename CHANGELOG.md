
# CHANGELOG lib-bpmn-engine

## v0.3.0-beta3

* add feature to export and import BPMN state, incl. resume capability (#12)
* make explicit engine name optional (#73 BREAKING CHANGE)
* use global ID generator internally, to avoid ID collisions between multiple engine instances 
* refactor `activity.LifecylceState` (BREAKING CHANGE)
* refactor `process_instance.State` (BREAKING CHANGE)
* new ExpressionEvaluationError

### Migration notes for breaking changes

#### New Initializer

Bpmn Engines are anonymous by default now, and shall be initialized by calling `.New()` \
**Example**: replace `bpmn_engine.New("name")` with `bpmn_engine.New()`

**Note**: you might use `.NewWithName("a name")` to assign different names for each engine instance.
This might help in scenarios, where you e.g. assign one engine instance to a thread.

## v0.3.0-beta2

Say "Hello!" to the new mascot \
![](./art/gopher-lib-bpmn-engine-96.png)

* introduce local variable scope for task handlers and do correct variable mapping on successful completion (#48 and #55)

----

## v0.3.0-beta1

* support handlers being registered for (task definition) types (#58 BREAKING CHANGE)
* support handlers for user tasks being registered for assignee or candidate groups (#59)
* improve documentation (#45)

### Migration notes for breaking changes

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
