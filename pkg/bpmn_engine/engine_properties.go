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

type BpmnEngineState struct {
	name              string
	processes         []ProcessInfo
	processInstances  []ProcessInstanceInfo
	queue             []BPMN20.BaseElement
	handlers          map[string]func(context ProcessInstanceContext)
	activationCounter map[string]int64
}

// GetProcessInstances returns an ordered list instance information.
func (state *BpmnEngineState) GetProcessInstances() []ProcessInstanceInfo {
	return state.processInstances
}

func (state *BpmnEngineState) GetName() string {
	return state.name
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
}

type ProcessInstance interface {
	GetProcessInfo() *ProcessInfo
	GetInstanceKey() int64
	GetVariableContext() map[string]string
	GetCreatedAt() time.Time
	// GetState returns one of [ProcessInstanceReady,ProcessInstanceActive,ProcessInstanceCompleted]
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
	return BPMN20.ProcessInstanceReady
}
