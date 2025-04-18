package bpmn_engine

import (
	"time"

	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/var_holder"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
)

type subProcessInfo struct {
	ElementId       string
	ProcessInstance *processInstanceInfo
	ProcessId       int64
	CreatedAt       time.Time
	processState    ActivityState
	variableHolder  var_holder.VariableHolder
	baseElement     *BPMN20.BaseElement
}

func (sb *subProcessInfo) Key() int64 {
	return sb.ProcessId
}

func (sb *subProcessInfo) State() ActivityState {
	return sb.processState
}

func (sb *subProcessInfo) SetState(state ActivityState) {
	sb.processState = state
}

func (sb *subProcessInfo) Element() *BPMN20.BaseElement {
	return sb.baseElement
}
