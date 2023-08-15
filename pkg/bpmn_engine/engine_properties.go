package bpmn_engine

import (
	"github.com/bwmarrin/snowflake"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/exporter"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
)

type ProcessInfo struct {
	BpmnProcessId string `json:"BpmnProcessId"` // The ID as defined in the BPMN file
	Version       int32  `json:"Version"`       // A version of the process, default=1, incremented, when another process with the same ID is loaded
	ProcessKey    int64  `json:"ProcessKey"`    // The engines key for this given process with version

	// TODO: make them private again?
	Definitions   BPMN20.TDefinitions `json:"definitions"`   // parsed file content
	ChecksumBytes [16]byte            `json:"checksumBytes"` // internal checksum to identify different versions
}

type BpmnEngineState struct {
	name                 string
	processes            []ProcessInfo
	processInstances     []*processInstanceInfo
	messageSubscriptions []*MessageSubscription
	jobs                 []*job
	timers               []*Timer
	scheduledFlows       []string
	taskHandlers         []*taskHandler
	exporters            []exporter.EventExporter
	snowflake            *snowflake.Node
}

// GetProcessInstances returns a list of instance information.
func (state *BpmnEngineState) GetProcessInstances() []*processInstanceInfo {
	return state.processInstances
}

// FindProcessInstanceById searches for a give processInstanceKey
// and returns the corresponding processInstanceInfo otherwise nil
func (state *BpmnEngineState) FindProcessInstanceById(processInstanceKey int64) *processInstanceInfo {
	for _, instance := range state.processInstances {
		if instance.InstanceKey == processInstanceKey {
			return instance
		}
	}
	return nil
}

// GetName returns the name of the engine, only useful in case you control multiple ones
func (state *BpmnEngineState) GetName() string {
	return state.name
}
