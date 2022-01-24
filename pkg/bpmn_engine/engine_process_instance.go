package bpmn_engine

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
	"time"
)

type ProcessInstanceInfo struct {
	processInfo     *ProcessInfo
	instanceKey     int64
	variableContext map[string]string
	createdAt       time.Time
	state           process_instance.State
	caughtEvents    []CatchEvent
}

type ProcessInstance interface {
	GetProcessInfo() *ProcessInfo
	GetInstanceKey() int64
	GetVariableContext() map[string]string
	GetCreatedAt() time.Time
	// GetState returns one of [ProcessInstanceReady,ProcessInstanceActive,ProcessInstanceCompleted]
	//  ┌─────┐
	//  │Ready│
	//  └──┬──┘
	// ┌───▽──┐
	// │Active│
	// └───┬──┘
	//┌────▽────┐
	//│Completed│
	//└─────────┘
	GetState() process_instance.State
}

func (pii *ProcessInstanceInfo) GetProcessInfo() *ProcessInfo {
	return pii.processInfo
}

func (pii *ProcessInstanceInfo) GetInstanceKey() int64 {
	return pii.instanceKey
}

func (pii *ProcessInstanceInfo) GetVariableContext() map[string]string {
	return pii.variableContext
}

func (pii *ProcessInstanceInfo) GetCreatedAt() time.Time {
	return pii.createdAt
}

func (pii *ProcessInstanceInfo) GetState() process_instance.State {
	return pii.state
}
