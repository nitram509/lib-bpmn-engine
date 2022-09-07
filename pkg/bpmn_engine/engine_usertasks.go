package bpmn_engine

import "github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"

func (state *BpmnEngineState) handleUserTask(process *ProcessInfo, instance *ProcessInstanceInfo, element *BPMN20.BaseElement) bool {
	// TODO consider different handlers, since Service Tasks are different in their definition than user tasks
	return state.handleServiceTask(process, instance, element)
}
