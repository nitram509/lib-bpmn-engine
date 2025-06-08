package bpmn_engine

import (
	"time"
)

// ActivatedJob is a struct to provide information for registered task handler
type activatedJob struct {
	processInstanceInfo      *processInstanceInfo
	completeHandler          func()
	failHandler              func(reason string)
	errorHandler             func(errorCode string)
	key                      int64
	processInstanceKey       int64
	bpmnProcessId            string
	processDefinitionVersion int32
	processDefinitionKey     int64
	elementId                string
	createdAt                time.Time
	variableHolder           VariableHolder
}

// ActivatedJob represents an abstraction for the activated job
// don't forget to call Fail or Complete when your task worker job is complete or not.
type ActivatedJob interface {
	// Key the key, a unique identifier for the job
	Key() int64

	// ProcessInstanceKey the job's process instance key
	ProcessInstanceKey() int64

	// BpmnProcessId Retrieve id of the job process definition
	BpmnProcessId() string

	// ProcessDefinitionVersion Retrieve version of the job process definition
	ProcessDefinitionVersion() int32

	// ProcessDefinitionKey Retrieve key of the job process definition
	ProcessDefinitionKey() int64

	// ElementId Get element id of the job
	ElementId() string

	// Variable from the process instance's variable context
	Variable(key string) interface{}

	// SetVariable in the variables context of the given process instance
	SetVariable(key string, value interface{})

	// InstanceKey get instance key from ProcessInfo
	InstanceKey() int64

	// CreatedAt when the job was created
	CreatedAt() time.Time

	// Fail does set the State the worker missed completing the job
	// ThrowError, Fail and Complete mutual exclude each other
	Fail(reason string)

	// ThrowError throws an error event
	// ThrowError, Fail and Complete mutual exclude each other
	ThrowError(errorCode string)

	// Complete does set the State the worker successfully completing the job
	// ThrowError, Fail and Complete mutual exclude each other
	Complete()
}

// CreatedAt implements ActivatedJob
func (aj *activatedJob) CreatedAt() time.Time {
	return aj.createdAt
}

// InstanceKey implements ActivatedJob
func (aj *activatedJob) InstanceKey() int64 {
	return aj.processInstanceInfo.GetInstanceKey()
}

// ElementId implements ActivatedJob
func (aj *activatedJob) ElementId() string {
	return aj.elementId
}

// Key implements ActivatedJob
func (aj *activatedJob) Key() int64 {
	return aj.key
}

// BpmnProcessId implements ActivatedJob
func (aj *activatedJob) BpmnProcessId() string {
	return aj.bpmnProcessId
}

// ProcessDefinitionKey implements ActivatedJob
func (aj *activatedJob) ProcessDefinitionKey() int64 {
	return aj.processDefinitionKey
}

// ProcessDefinitionVersion implements ActivatedJob
func (aj *activatedJob) ProcessDefinitionVersion() int32 {
	return aj.processDefinitionVersion
}

// ProcessInstanceKey implements ActivatedJob
func (aj *activatedJob) ProcessInstanceKey() int64 {
	return aj.processInstanceKey
}

// Variable implements ActivatedJob
func (aj *activatedJob) Variable(key string) interface{} {
	return aj.variableHolder.GetVariable(key)
}

// SetVariable implements ActivatedJob
func (aj *activatedJob) SetVariable(key string, value interface{}) {
	aj.variableHolder.SetVariable(key, value)
}

// Fail implements ActivatedJob
func (aj *activatedJob) Fail(reason string) {
	aj.failHandler(reason)
}

// ThrowError implements ActivatedJob
func (aj *activatedJob) ThrowError(errorCode string) {
	aj.errorHandler(errorCode)
}

// Complete implements ActivatedJob
func (aj *activatedJob) Complete() {
	aj.completeHandler()
}
