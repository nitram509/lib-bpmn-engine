package bpmn_engine

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
)

type ProcessInfo struct {
	BpmnProcessId string // The ID as defined in the BPMN file
	Version       int32  // A version of the process, default=1, incremented, when another process with the same ID is loaded
	ProcessKey    int64  // The engines key for this given process with version

	definitions   BPMN20.TDefinitions // parsed file content
	checksumBytes [16]byte            // internal checksum to identify different versions
}

type BpmnEngineState struct {
	name                 string
	processes            []ProcessInfo
	processInstances     []*ProcessInstanceInfo
	messageSubscriptions []*MessageSubscription
	jobs                 []*Job
	timers               []*Timer
	scheduledFlows       []string
	handlers             map[string]func(job ActivatedJob)
}

func (state *BpmnEngineState) GetProcessInstances() []*ProcessInstanceInfo {
	return state.processInstances
}

func (state *BpmnEngineState) FindProcessInstanceById(processInstanceKey int64) *ProcessInstanceInfo {
	for _, instance := range state.processInstances {
		if instance.instanceKey == processInstanceKey {
			return instance
		}
	}
	return nil
}

func (state *BpmnEngineState) GetName() string {
	return state.name
}
