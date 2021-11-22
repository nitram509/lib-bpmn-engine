package bpmn_engine

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"time"
)

type ProcessInfo struct {
	BpmnProcessId string // The ID as defined in the BPMN file
	Version       int32  // A version of the process, default=1, incremented, when another process with the same ID is loaded
	ProcessKey    int64  // The engines key for this given process with version

	definitions   BPMN20.TDefinitions // parsed file content
	checksumBytes [16]byte            // internal checksum to identify different versions
}

type InstanceInfo struct {
	processInfo     *ProcessInfo
	InstanceKey     int64
	VariableContext map[string]string
	createdAt       time.Time
}

type BpmnEngineState struct {
	name              string
	processes         []ProcessInfo
	processInstances  []InstanceInfo
	queue             []BPMN20.BaseElement
	handlers          map[string]func(context ProcessInstanceContext)
	activationCounter map[string]int64
}

type ProcessInstanceContext interface {
	GetTaskId() string
	GetVariable(name string) string
	SetVariable(name string, value string)
}

// GetProcessInstances returns an ordered list instance information.
func (state *BpmnEngineState) GetProcessInstances() []InstanceInfo {
	return state.processInstances
}

func (state *BpmnEngineState) GetName() string {
	return state.name
}
