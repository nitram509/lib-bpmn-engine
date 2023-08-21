package bpmn_engine

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/var_holder"
	"time"

	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
)

type processInstanceInfo struct {
	ProcessInfo    *ProcessInfo              `json:"-"`
	InstanceKey    int64                     `json:"ik"`
	VariableHolder var_holder.VariableHolder `json:"vh"`
	CreatedAt      time.Time                 `json:"c"`
	State          process_instance.State    `json:"s"`
	CaughtEvents   []catchEvent              `json:"ce"`
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

func (pii *processInstanceInfo) GetProcessInfo() *ProcessInfo {
	return pii.ProcessInfo
}

func (pii *processInstanceInfo) GetInstanceKey() int64 {
	return pii.InstanceKey
}

func (pii *processInstanceInfo) GetVariable(key string) interface{} {
	return pii.VariableHolder.GetVariable(key)
}

func (pii *processInstanceInfo) SetVariable(key string, value interface{}) {
	pii.VariableHolder.SetVariable(key, value)
}

func (pii *processInstanceInfo) GetCreatedAt() time.Time {
	return pii.CreatedAt
}

// GetState returns one of [READY, ACTIVE, COMPLETED, FAILED]
// State diagram:
//
//	 ┌─────┐
//	 │Ready│
//	 └──┬──┘
//	    |
//	┌───▽──┐
//	│Active│
//	└───┬──┘
//	    |
//
// ┌────▽────┐
// │Completed│
// └─────────┘
func (pii *processInstanceInfo) GetState() process_instance.State {
	return pii.State
}
