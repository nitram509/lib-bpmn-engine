package bpmn_engine

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/var_holder"
)

const CurrentSerializerVersion = 1

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

type ProcessInstanceInfoAlias processInstanceInfo
type processInstanceInfoAdapter struct {
	ProcessKey   int64     `json:"pk"`
	CommandQueue []command `json:"cq"`
	*ProcessInstanceInfoAlias
}

type execCommandAdapter struct {
	InboundFlowId   string `json:"fid"`
	BaseElementId   string `json:"eid"`
	BaseElementName string `json:"en"`
}

func (pii *processInstanceInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(&processInstanceInfoAdapter{
		ProcessKey:               pii.ProcessInfo.ProcessKey,
		CommandQueue:             pii.commandQueue,
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
	pii.commandQueue = adapter.CommandQueue
	return nil
}

//func (cmd execCommand) MarshalJSON() ([]byte, error) {
//	cmdAdapter := execCommandAdapter{
//		InboundFlowId:   cmd.inboundFlowId,
//		BaseElementId:   cmd.baseElement.GetId(),
//		BaseElementName: cmd.baseElement.GetName(),
//	}
//	return json.Marshal(&cmdAdapter)
//}
//
//func (cmd execCommand) UnmarshalJSON(data []byte) error {
//	adapter := &execCommandAdapter{}
//	if err := json.Unmarshal(data, &adapter); err != nil {
//		return err
//	}
//	cmd.inboundFlowId = adapter.InboundFlowId
//	// TODO: incorrect use of TStartEvent ... should rather be something more suitable for unmarshalling
//	cmd.baseElement = BPMN20.TStartEvent{
//		Id:   adapter.BaseElementId,
//		Name: adapter.BaseElementName,
//	}
//	return nil
//}

func (state *BpmnEngineState) Marshal() []byte {
	m := serializedBpmnEngine{
		Version:              CurrentSerializerVersion,
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

// Unmarshal loads the data byte array and creates a new instance of the BPMN Engine
// Will return an BpmnEngineUnmarshallingError, if there was an issue AND in case of error,
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
				msg := "Can't load BPMN from serialized data"
				return state, &BpmnEngineUnmarshallingError{
					Msg: msg,
					Err: err,
				}
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
				msg := fmt.Sprintf("Can't find process key %d in current BPMN Engine's processes", pi.ProcessInfo.ProcessKey)
				return state, &BpmnEngineUnmarshallingError{
					Msg: msg,
				}
			}
			m.ProcessInstances[i].ProcessInfo = process
			m.ProcessInstances[i].VariableHolder = var_holder.New(nil, nil)
			err = restoreCommandQueue(process, m.ProcessInstances[i])
			if err != nil {
				return state, err
			}
		}
		state.processInstances = m.ProcessInstances
	}
	if m.Timers != nil {
		state.timers = m.Timers
	}
	return state, nil
}

// restoreCommandQueue post process the commands and restore pointers
func restoreCommandQueue(process *ProcessInfo, instance *processInstanceInfo) (err error) {
	for _, cmd := range instance.commandQueue {
		println(cmd.Type()) // FIXME
		//baseElements := BPMN20.FindBaseElementsById(process.definitions, cmd.baseElement.GetId())
		//found := false
		//for i := 0; i < len(baseElements); i++ {
		//	be := baseElements[i]
		//	found = be.GetId() == cmd.baseElement.GetId() && be.GetName() == cmd.baseElement.GetName()
		//	if found {
		//		cmd.baseElement = be
		//		break
		//	}
		//}
		//if !found {
		//	msg := fmt.Sprintf("Can't restore command queue element with id=%s, name=%s not found in BPMN definitions",
		//		cmd.baseElement.GetId(), cmd.baseElement.GetName())
		//	return &BpmnEngineUnmarshallingError{Msg: msg}
		//}
	}
	return err
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

func (state *BpmnEngineState) findProcess(processKey int64) *ProcessInfo {
	for i := 0; i < len(state.processes); i++ {
		process := state.processes[i]
		if process.ProcessKey == processKey {
			return process
		}
	}
	return nil
}
