package bpmn_engine

import "github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"

func (state *BpmnEngineState) handleUserTask(process BPMN20.ProcessElement, instance *processInstanceInfo, element *BPMN20.TaskElement) *job {
	// TODO consider different handlers, since Service Tasks are different in their definition than user tasks
	_, j := state.handleServiceTask(process, instance, element)
	return j
}
