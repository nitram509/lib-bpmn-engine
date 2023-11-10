package bpmn_engine

import (
	"time"

	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/var_holder"
)

// FIXME: shall this be exported?
type processInstanceInfo struct {
	ProcessInfo    *ProcessInfo              `json:"-"`
	InstanceKey    int64                     `json:"ik"`
	VariableHolder var_holder.VariableHolder `json:"vh,omitempty"`
	CreatedAt      time.Time                 `json:"c"`
	State          ActivityState             `json:"s"`
	CaughtEvents   []catchEvent              `json:"ce,omitempty"`
	activities     []activity
}

type ProcessInstance interface {
	GetProcessInfo() *ProcessInfo
	GetInstanceKey() int64

	// GetVariable from the process instance's variable context
	GetVariable(key string) interface{}

	// SetVariable to the process instance's variable context
	SetVariable(key string, value interface{})

	GetCreatedAt() time.Time
	GetState() ActivityState
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

// GetState returns one of [ Ready, Active, Completed, Failed ]
func (pii *processInstanceInfo) GetState() ActivityState {
	return pii.State
}

func (pii *processInstanceInfo) appendActivity(activity activity) {
	pii.activities = append(pii.activities, activity)
}

func (pii *processInstanceInfo) findActiveActivityByElementId(id string) activity {
	for _, a := range pii.activities {
		if (*a.Element()).GetId() == id && a.State() == Active {
			return a
		}
	}
	return nil
}

func (pii *processInstanceInfo) findActivity(key int64) activity {
	for _, a := range pii.activities {
		if a.Key() == key {
			return a
		}
	}
	return nil
}
