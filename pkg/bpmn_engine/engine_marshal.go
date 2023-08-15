package bpmn_engine

import (
	"encoding/json"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/var_holder"
)

type serializedBpmnEngine struct {
	Version              int                    `json:"version"`
	Name                 string                 `json:"name"`
	Processes            []ProcessInfo          `json:"Processes,omitempty"`
	ProcessInstances     []*processInstanceInfo `json:"ProcessInstances,omitempty"`
	MessageSubscriptions []*MessageSubscription `json:"MessageSubscriptions,omitempty"`
	Timers               []*Timer               `json:"Timers,omitempty"`
}

func (pii *processInstanceInfo) MarshalJSON() ([]byte, error) {
	type Alias processInstanceInfo
	return json.Marshal(&struct {
		ProcessKey int64 `json:"ProcessKey"`
		*Alias
	}{
		ProcessKey: pii.ProcessInfo.ProcessKey,
		Alias:      (*Alias)(pii),
	})
}

func (pii *processInstanceInfo) UnmarshalJSON(data []byte) error {
	type Alias processInstanceInfo
	aux := &struct {
		ProcessKey int64 `json:"ProcessKey"`
		*Alias
	}{
		Alias: (*Alias)(pii),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	pii.ProcessInfo = &ProcessInfo{ProcessKey: aux.ProcessKey}
	return nil
}

func (state *BpmnEngineState) Marshal() []byte {
	m := serializedBpmnEngine{
		Name:                 state.name,
		MessageSubscriptions: state.messageSubscriptions,
		Processes:            state.processes,
		ProcessInstances:     state.processInstances,
		Timers:               state.timers,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return bytes
}

func Unmarshal(data []byte) BpmnEngineState {
	m := serializedBpmnEngine{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		panic(err)
	}
	state := New(m.Name)
	if m.Processes != nil {
		state.processes = m.Processes
	}
	if m.MessageSubscriptions != nil {
		state.messageSubscriptions = m.MessageSubscriptions
	}
	if m.ProcessInstances != nil {
		for i, pi := range m.ProcessInstances {
			process := state.findProcess(pi.ProcessInfo.ProcessKey)
			if process == nil {
				panic("TODO") // TODO, do proper error handling
			}
			m.ProcessInstances[i].ProcessInfo = process
			m.ProcessInstances[i].VariableHolder = var_holder.New(nil, nil)
		}
		state.processInstances = m.ProcessInstances
	}
	if m.Timers != nil {
		state.timers = m.Timers
	}
	return state
}

func (state *BpmnEngineState) findProcess(processKey int64) *ProcessInfo {
	for i := 0; i < len(state.processes); i++ {
		process := &state.processes[i]
		if process.ProcessKey == processKey {
			return process
		}
	}
	return nil
}
