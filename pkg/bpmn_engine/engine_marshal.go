package bpmn_engine

import "encoding/json"

type serializedBpmnEngine struct {
	Version              int                    `json:"version"`
	Name                 string                 `json:"name"`
	MessageSubscriptions []*MessageSubscription `json:"MessageSubscriptions,omitempty"`
	Processes            []ProcessInfo          `json:"Processes,omitempty"`
}

func (state *BpmnEngineState) Marshal() []byte {
	m := serializedBpmnEngine{
		Name:                 state.name,
		MessageSubscriptions: state.messageSubscriptions,
		Processes:            state.processes,
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
	state := New("foo")
	state.name = m.Name
	if m.Processes != nil {
		state.processes = m.Processes
	}
	if m.MessageSubscriptions != nil {
		state.messageSubscriptions = m.MessageSubscriptions
	}
	return state
}
