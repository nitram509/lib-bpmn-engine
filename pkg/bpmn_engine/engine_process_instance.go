package bpmn_engine

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/var_holder"
	"time"

	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
)

type ProcessInstanceInfo struct {
	processInfo    *ProcessInfo
	instanceKey    int64
	createdAt      time.Time
	state          process_instance.State
	caughtEvents   []catchEvent
	variableHolder var_holder.VariableHolder
}

type ProcessInstance interface {
	GetProcessInfo() *ProcessInfo
	GetInstanceKey() int64

	// GetVariable from the process instance's variable context
	GetVariable(key string) interface{}

	// SetVariable to the process instance's variable context
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
	return pii.variableHolder.GetVariable(key)
}

func (pii *ProcessInstanceInfo) SetVariable(key string, value interface{}) {
	pii.variableHolder.SetVariable(key, value)
}

func (pii *ProcessInstanceInfo) GetCreatedAt() time.Time {
	return pii.createdAt
}

// GetState returns one of [READY, ACTIVE, COMPLETED, FAILED]
// State diagram:
//   ┌─────┐
//   │Ready│
//   └──┬──┘
//      |
//  ┌───▽──┐
//  │Active│
//  └───┬──┘
//      |
// ┌────▽────┐
// │Completed│
// └─────────┘
func (pii *ProcessInstanceInfo) GetState() process_instance.State {
	return pii.state
}
