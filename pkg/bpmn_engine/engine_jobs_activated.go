package bpmn_engine

import (
	"encoding/xml"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/var_holder"
	"time"
)

// ActivatedJob is a struct to provide information for registered task handler
type activatedJob struct {
	processInstanceInfo      *ProcessInstanceInfo
	completeHandler          func()
	failHandler              func(reason string)
	key                      int64
	processInstanceKey       int64
	bpmnProcessId            string
	processDefinitionVersion int32
	processDefinitionKey     int64
	elementId                string
	createdAt                time.Time
	variableHolder           var_holder.VariableHolder
	attributes               []xml.Attr
}

// ActivatedJob represents an abstraction for the activated job
// don't forget to call Fail or Complete when your task worker job is complete or not.
type ActivatedJob interface {
	// GetKey the key, a unique identifier for the job
	GetKey() int64

	// GetProcessInstanceKey the job's process instance key
	GetProcessInstanceKey() int64

	// GetBpmnProcessId Retrieve id of the job process definition
	GetBpmnProcessId() string

	// GetProcessDefinitionVersion Retrieve version of the job process definition
	GetProcessDefinitionVersion() int32

	// GetProcessDefinitionKey Retrieve key of the job process definition
	GetProcessDefinitionKey() int64

	// GetElementId Get element id of the job
	GetElementId() string

	// GetVariable from the process instance's variable context
	GetVariable(key string) interface{}

	// SetVariable in the variables context of the given process instance
	SetVariable(key string, value interface{})

	// GetInstanceKey get instance key from processInfo
	GetInstanceKey() int64

	// GetCreatedAt when the job was created
	GetCreatedAt() time.Time

	// Fail does set the state the worker missed completing the job
	// Fail and Complete mutual exclude each other
	Fail(reason string)

	// Complete does set the state the worker successfully completing the job
	// Fail and Complete mutual exclude each other
	Complete()

	// GetAttributes Get attributes id of the job
	GetAttributes() []xml.Attr

	// GetAttribute Get attribute id of the job
	GetAttribute(name xml.Name) *xml.Attr
}

// GetCreatedAt implements ActivatedJob
func (aj *activatedJob) GetCreatedAt() time.Time {
	return aj.createdAt
}

// GetInstanceKey implements ActivatedJob
func (aj *activatedJob) GetInstanceKey() int64 {
	return aj.processInstanceInfo.GetInstanceKey()
}

// GetElementId implements ActivatedJob
func (aj *activatedJob) GetElementId() string {
	return aj.elementId
}

// GetKey implements ActivatedJob
func (aj *activatedJob) GetKey() int64 {
	return aj.key
}

// GetBpmnProcessId implements ActivatedJob
func (aj *activatedJob) GetBpmnProcessId() string {
	return aj.bpmnProcessId
}

// GetProcessDefinitionKey implements ActivatedJob
func (aj *activatedJob) GetProcessDefinitionKey() int64 {
	return aj.processDefinitionKey
}

// GetProcessDefinitionVersion implements ActivatedJob
func (aj *activatedJob) GetProcessDefinitionVersion() int32 {
	return aj.processDefinitionVersion
}

// GetProcessInstanceKey implements ActivatedJob
func (aj *activatedJob) GetProcessInstanceKey() int64 {
	return aj.processInstanceKey
}

// GetVariable implements ActivatedJob
func (aj *activatedJob) GetVariable(key string) interface{} {
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

// Complete implements ActivatedJob
func (aj *activatedJob) Complete() {
	aj.completeHandler()
}

// GetAttributes Get attributes id of the job
func (aj *activatedJob) GetAttributes() []xml.Attr {
	return aj.attributes
}

// GetAttribute Get attribute id of the job
func (aj *activatedJob) GetAttribute(name xml.Name) *xml.Attr {
	for _, attr := range aj.attributes {
		if attr.Name == name {
			return &attr
		}
	}
	return nil
}
