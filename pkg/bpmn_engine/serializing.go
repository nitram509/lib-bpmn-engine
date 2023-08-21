package bpmn_engine

import (
	"encoding/hex"
	"encoding/json"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/var_holder"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
	"time"
)

type serializedBpmnEngine struct {
	Version              int                    `json:"v"`
	Name                 string                 `json:"n"`
	ProcessReferences    []processInfoReference `json:"pr,omitempty"`
	ProcessInstances     []*processInstanceInfo `json:"pi,omitempty"`
	MessageSubscriptions []*MessageSubscription `json:"ms,omitempty"`
	Timers               []*Timer               `json:"t,omitempty"`
}

type processInfoReference struct {
	BpmnProcessId    string `json:"id"`           // The ID as defined in the BPMN file
	ProcessKey       int64  `json:"pk"`           // The engines key for this given process with version
	BpmnData         string `json:"d"`            // the raw BPMN XML data
	BpmnResourceName string `json:"rn,omitempty"` // the resource's name
	BpmnChecksum     string `json:"crc"`          // internal checksum to identify different versions
}

type processInstanceInfoReference struct {
	Processkey     int64                     `json:"pk"`
	InstanceKey    int64                     `json:"ik"`
	VariableHolder var_holder.VariableHolder `json:"vh"`
	CreatedAt      time.Time                 `json:"c"`
	State          process_instance.State    `json:"s"`
	CaughtEvents   []catchEvent              `json:"ce"`
}

type ProcessInstanceInfoAlias processInstanceInfo
type processInstanceInfoAdapter struct {
	ProcessKey int64 `json:"pk"`
	*ProcessInstanceInfoAlias
}

func (pii *processInstanceInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(&processInstanceInfoAdapter{
		ProcessKey:               pii.ProcessInfo.ProcessKey,
		ProcessInstanceInfoAlias: (*ProcessInstanceInfoAlias)(pii),
	})
}

func (pii *processInstanceInfo) UnmarshalJSON(data []byte) error {
	adapter := &processInstanceInfoAdapter{
		ProcessInstanceInfoAlias: (*ProcessInstanceInfoAlias)(pii),
	}
	if err := json.Unmarshal(data, &adapter); err != nil {
		return err
	}
	pii.ProcessInfo = &ProcessInfo{ProcessKey: adapter.ProcessKey}
	return nil
}

func (state *BpmnEngineState) Marshal() []byte {
	m := serializedBpmnEngine{
		Name:                 state.name,
		MessageSubscriptions: state.messageSubscriptions,
		ProcessReferences:    createReferences(state.processes),
		ProcessInstances:     state.processInstances,
		Timers:               state.timers,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return bytes
}

func createReferences(processes []*ProcessInfo) (result []processInfoReference) {
	for _, pi := range processes {
		ref := processInfoReference{
			BpmnProcessId:    pi.BpmnProcessId,
			ProcessKey:       pi.ProcessKey,
			BpmnData:         pi.bpmnData,
			BpmnResourceName: pi.bpmnResourceName,
			BpmnChecksum:     hex.EncodeToString(pi.bpmnChecksum[:]),
		}
		result = append(result, ref)
	}
	return result
}

// Unmarshal loads the data byte array and creates a new instance of the BPMN Engine
// Will return an error, if there was an issue AND in case of error,
// the engine return object is only partially initialized and likely not usable
func Unmarshal(data []byte) (BpmnEngineState, error) {
	m := serializedBpmnEngine{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		panic(err)
	}
	state := New()
	if m.ProcessReferences != nil {
		for _, pir := range m.ProcessReferences {
			xmlData := decodeAndDecompress(pir.BpmnData)
			process, err := state.load(xmlData, pir.BpmnResourceName)
			process.ProcessKey = pir.ProcessKey
			if err != nil {
				return state, err
			}
		}
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
	return state, nil
}

func (state *BpmnEngineState) findProcess(processKey int64) *ProcessInfo {
	for i := 0; i < len(state.processes); i++ {
		process := state.processes[i]
		if process.ProcessKey == processKey {
			return process
		}
	}
	return nil
}
