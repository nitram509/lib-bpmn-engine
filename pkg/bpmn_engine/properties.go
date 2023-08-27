package bpmn_engine

import (
	"github.com/bwmarrin/snowflake"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/exporter"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"sort"
)

type BpmnEngineState struct {
	name                 string
	processes            []*ProcessInfo
	processInstances     []*processInstanceInfo
	messageSubscriptions []*MessageSubscription
	jobs                 []*job
	timers               []*Timer
	scheduledFlows       []string // Deprecated: FIXME is this correct per ENGINE STATE? or should it rather be on per INSTANCE level?
	taskHandlers         []*taskHandler
	exporters            []exporter.EventExporter
	snowflake            *snowflake.Node
}

type ProcessInfo struct {
	BpmnProcessId    string              // The ID as defined in the BPMN file
	Version          int32               // A version of the process, default=1, incremented, when another process with the same ID is loaded
	ProcessKey       int64               // The engines key for this given process with version
	definitions      BPMN20.TDefinitions // parsed file content
	bpmnData         string              // the raw source data, compressed and encoded via ascii85
	bpmnResourceName string              // some name for the resource
	bpmnChecksum     [16]byte            // internal checksum to identify different versions
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

// FindProcessesById returns all registered processes with given ID
// result array is ordered by version number, from 1 (first) and largest version (last)
func (state *BpmnEngineState) FindProcessesById(id string) (infos []*ProcessInfo) {
	for _, p := range state.processes {
		if p.BpmnProcessId == id {
			infos = append(infos, p)
		}
	}
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Version < infos[j].Version
	})
	return infos
}
