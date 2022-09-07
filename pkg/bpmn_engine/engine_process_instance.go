package bpmn_engine

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
	"time"
)

type ProcessInstanceInfo struct {
	processInfo     *ProcessInfo
	instanceKey     int64
	variableContext map[string]interface{}
	createdAt       time.Time
	state           process_instance.State
	caughtEvents    []CatchEvent
}

type ProcessInstance interface {
	GetProcessInfo() *ProcessInfo
	GetInstanceKey() int64
	GetVariable(key string) interface{}
	SetVariable(key string, value interface{})
	GetCreatedAt() time.Time
	GetState() process_instance.State
}

func (pii *ProcessInstanceInfo) GetProcessInfo() *ProcessInfo {
	return pii.processInfo
}

func (pii *ProcessInstanceInfo) GetInstanceKey() int64 {
	return pii.instanceKey
}

func (pii *ProcessInstanceInfo) GetVariable(key string) interface{} {
	return pii.variableContext[key]
}

func (pii *ProcessInstanceInfo) SetVariable(key string, value interface{}) {
	pii.variableContext[key] = value
}

func (pii *ProcessInstanceInfo) GetCreatedAt() time.Time {
	return pii.createdAt
}

// GetState returns one of [ProcessInstanceReady,ProcessInstanceActive,ProcessInstanceCompleted]
//
//	┌─────┐
//	│Ready│
//	└──┬──┘
//
// ┌───▽──┐
// │Active│
// └───┬──┘
// ┌────▽────┐
// │Completed│
// └─────────┘
func (pii *ProcessInstanceInfo) GetState() process_instance.State {
	return pii.state
}
