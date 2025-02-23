package bpmn_engine

import (
	"time"

	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/var_holder"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
)

type process struct {
	ElementId      string
	Instance       *processInstanceInfo
	ProcessId      int64
	CreatedAt      time.Time
	processState   ActivityState
	variableHolder var_holder.VariableHolder
	baseElement    *BPMN20.BaseElement
}

func (p process) Key() int64 {
	return p.ProcessId
}

func (p process) State() ActivityState {
	return p.processState
}

func (p process) Element() *BPMN20.BaseElement {
	return p.baseElement
}
