
# CHANGELOG lib-bpmn-engine

## v0.3.0-beta5

- rename FindProcessInstanceById() -> FindProcessInstance()

## v0.3.0-beta4

* WiP serialization and deserialization ... please don't try yet ... it's not working in this beta3 version
* make explicit engine name optional (#73 BREAKING CHANGE)
* use global ID generator internally, to avoid ID collisions between multiple engine instances 
* refactor `activity.LifecylceState` (BREAKING CHANGE)
* refactor `process_instance.State` (BREAKING CHANGE)
   * new return type for e.g. `instance.GetState()` --is--> `ActivityState` (BREAKING CHANGE)
* new ExpressionEvaluationError
  * improved errors for intermediate timer catch events (#38, #69)
  * improved error handling for intermediate message catch events
  * variable mapping and expression evaluation errors
  * return proper BpmnEngineError, at creation time (BREAKING CHANGE)
* new support intermediate link throw & catch element (#141)
* new `CreateInstanceById()` and `CreateAndRunInstanceById()` functions to ease handling with multiple versions
* use Go Getter idiomatic (BREAKING CHANGE) (#144)

### Migration notes for breaking changes

#### New Initializer

Bpmn Engines are anonymous by default now, and shall be initialized by calling `.New()` \
**Example**: replace `bpmn_engine.New("name")` with `bpmn_engine.New()`

**Note**: you might use `.NewWithName("a name")` to assign different names for each engine instance.
This might help in scenarios, where you e.g. assign one engine instance to a thread.

#### `activity.LifecylceState` and `process_instance.State`

Both are consolidated towards `bpmn_engine.ActivityState`, which you can simply use in the same manner.

#### Use Go Getter idiomatic

For some interfaces, the prior code looked like e.g. `engine.GetName()`.
According to https://go.dev/doc/effective_go#Getters, this getter should better be written as `engine.Name()`.

----

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
