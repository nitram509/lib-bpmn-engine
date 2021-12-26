package bpmn_engine

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"time"
)

type ProcessInstanceContextData struct {
	taskId       string
	processInfo  *ProcessInfo
	instanceInfo *ProcessInstanceInfo
}

type ProcessInstanceContext interface {
	GetTaskId() string
	GetVariable(name string) string
	SetVariable(name string, value string)
}

type ProcessInstanceInfo struct {
	processInfo     *ProcessInfo
	instanceKey     int64
	variableContext map[string]string
	createdAt       time.Time
	state           BPMN20.ProcessInstanceState
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
	GetState() BPMN20.ProcessInstanceState
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

func (pii *ProcessInstanceInfo) GetState() BPMN20.ProcessInstanceState {
	return pii.state
}

func (data *ProcessInstanceContextData) GetTaskId() string {
	return data.taskId
}

func (data *ProcessInstanceContextData) GetVariable(name string) string {
	return data.instanceInfo.variableContext[name]
}

func (data *ProcessInstanceContextData) SetVariable(name string, value string) {
	data.instanceInfo.variableContext[name] = value
}
