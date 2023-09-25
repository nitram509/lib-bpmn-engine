package bpmn_engine

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/var_holder"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
)

const CurrentSerializerVersion = 1

type serializedBpmnEngine struct {
	Version              int                    `json:"v"`
	Name                 string                 `json:"n"`
	ProcessReferences    []processInfoReference `json:"pr,omitempty"`
	ProcessInstances     []*processInstanceInfo `json:"pi,omitempty"`
	MessageSubscriptions []*MessageSubscription `json:"ms,omitempty"`
	Timers               []*Timer               `json:"t,omitempty"`
	Jobs                 []*job                 `json:"j,omitempty"`
}

type processInfoReference struct {
	BpmnProcessId    string `json:"id"`           // The ID as defined in the BPMN file
	ProcessKey       int64  `json:"pk"`           // The engines key for this given process with version
	BpmnData         string `json:"d"`            // the raw BPMN XML data
	BpmnResourceName string `json:"rn,omitempty"` // the resource's name
	BpmnChecksum     string `json:"crc"`          // internal checksum to identify different versions
}

type ProcessInstanceInfoAlias processInstanceInfo // FIXME: don't export
type processInstanceInfoAdapter struct {
	ProcessKey       int64              `json:"pk"`
	ActivityAdapters []*activityAdapter `json:"a,omitempty"`
	*ProcessInstanceInfoAlias
}

type TimerAlias Timer
type timerAdapter struct {
	OriginActivityKey int64 `json:"oak"`
	*TimerAlias
}

type MessageSubscriptionAlias MessageSubscription
type messageSubscriptionAdapter struct {
	OriginActivityKey int64 `json:"oak"`
	*MessageSubscriptionAlias
}

type activityAdapterType int

const (
	gatewayActivityAdapterType = iota
	eventBasedGatewayActivityAdapterType
)

type activityAdapter struct {
	Type                      activityAdapterType `json:"t"`
	Key                       int64               `json:"k"`
	State                     ActivityState       `json:"s"`
	ElementReference          string              `json:"e"`
	Parallel                  bool                `json:"p,omitempty"` // from gatewayActivity
	InboundFlowIdsCompleted   []string            `json:"i,omitempty"` // from gatewayActivity
	OutboundActivityCompleted string              `json:"o,omitempty"` // from eventBasedGatewayActivity
}

type baseElementPlaceholder struct {
	id string
}

func (b baseElementPlaceholder) GetId() string {
	return b.id
}

func (b baseElementPlaceholder) GetName() string {
	panic("the placeholder does not implement all methods, by intent")
}

func (b baseElementPlaceholder) GetIncomingAssociation() []string {
	panic("the placeholder does not implement all methods, by intent")
}

func (b baseElementPlaceholder) GetOutgoingAssociation() []string {
	panic("the placeholder does not implement all methods, by intent")
}

func (b baseElementPlaceholder) GetType() BPMN20.ElementType {
	panic("the placeholder does not implement all methods, by intent")
}

// ----------------------------------------------------------------------------

type activityPlaceholder struct {
	key int64
}

func (a activityPlaceholder) Key() int64 {
	return a.key
}

func (a activityPlaceholder) State() ActivityState {
	panic("the placeholder does not implement all methods, by intent")
}

func (a activityPlaceholder) Element() *BPMN20.BaseElement {
	panic("the placeholder does not implement all methods, by intent")
}

// ----------------------------------------------------------------------------

func (t *Timer) MarshalJSON() ([]byte, error) {
	ta := &timerAdapter{
		OriginActivityKey: t.originActivity.Key(),
		TimerAlias:        (*TimerAlias)(t),
	}
	return json.Marshal(ta)
}

func (t *Timer) UnmarshalJSON(data []byte) error {
	ta := timerAdapter{
		TimerAlias: (*TimerAlias)(t),
	}
	if err := json.Unmarshal(data, &ta); err != nil {
		return err
	}
	t.originActivity = &activityPlaceholder{key: ta.OriginActivityKey}
	return nil
}

// ----------------------------------------------------------------------------

func (m *MessageSubscription) MarshalJSON() ([]byte, error) {
	msa := &messageSubscriptionAdapter{
		OriginActivityKey:        m.originActivity.Key(),
		MessageSubscriptionAlias: (*MessageSubscriptionAlias)(m),
	}
	return json.Marshal(msa)
}

func (m *MessageSubscription) UnmarshalJSON(data []byte) error {
	msa := messageSubscriptionAdapter{
		MessageSubscriptionAlias: (*MessageSubscriptionAlias)(m),
	}
	if err := json.Unmarshal(data, &msa); err != nil {
		return err
	}
	m.originActivity = &activityPlaceholder{key: msa.OriginActivityKey}
	return nil
}

// ----------------------------------------------------------------------------

func (pii *processInstanceInfo) MarshalJSON() ([]byte, error) {
	piia := &processInstanceInfoAdapter{
		ProcessKey:               pii.ProcessInfo.ProcessKey,
		ProcessInstanceInfoAlias: (*ProcessInstanceInfoAlias)(pii),
	}
	for _, a := range pii.activities {
		switch activity := a.(type) {
		case *gatewayActivity:
			piia.ActivityAdapters = append(piia.ActivityAdapters, createGatewayActivityAdapter(activity))
		case *eventBasedGatewayActivity:
			piia.ActivityAdapters = append(piia.ActivityAdapters, createEventBasedGatewayActivityAdapter(activity))
		default:
			panic(fmt.Sprintf("[invariant check] missing activity adapter for the type %T", a))
		}
	}
	return json.Marshal(piia)
}

func (pii *processInstanceInfo) UnmarshalJSON(data []byte) error {
	adapter := &processInstanceInfoAdapter{
		ProcessInstanceInfoAlias: (*ProcessInstanceInfoAlias)(pii),
	}
	if err := json.Unmarshal(data, &adapter); err != nil {
		return err
	}
	pii.ProcessInfo = &ProcessInfo{ProcessKey: adapter.ProcessKey}
	recoverProcessInstanceActivitiesPart1(pii, adapter)
	return nil
}

func createEventBasedGatewayActivityAdapter(ebga *eventBasedGatewayActivity) *activityAdapter {
	aa := &activityAdapter{
		Type:                      eventBasedGatewayActivityAdapterType,
		Key:                       ebga.key,
		State:                     ebga.state,
		ElementReference:          (*ebga.element).GetId(),
		OutboundActivityCompleted: ebga.OutboundActivityCompleted,
	}
	return aa
}

func createGatewayActivityAdapter(ga *gatewayActivity) *activityAdapter {
	aa := &activityAdapter{
		Type:                    gatewayActivityAdapterType,
		Key:                     ga.key,
		State:                   ga.state,
		ElementReference:        (*ga.element).GetId(),
		Parallel:                ga.parallel,
		InboundFlowIdsCompleted: ga.inboundFlowIdsCompleted,
	}
	return aa
}

// ----------------------------------------------------------------------------

func (state *BpmnEngineState) Marshal() []byte {
	m := serializedBpmnEngine{
		Version:              CurrentSerializerVersion,
		Name:                 state.name,
		MessageSubscriptions: state.messageSubscriptions,
		ProcessReferences:    createReferences(state.processes),
		ProcessInstances:     state.processInstances,
		Timers:               state.timers,
		Jobs:                 state.jobs,
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
	eng := serializedBpmnEngine{}
	err := json.Unmarshal(data, &eng)
	if err != nil {
		panic(err)
	}
	state := New()
	if eng.ProcessReferences != nil {
		for _, pir := range eng.ProcessReferences {
			xmlData := decodeAndDecompress(pir.BpmnData)
			process, err := state.load(xmlData, pir.BpmnResourceName)
			process.ProcessKey = pir.ProcessKey
			if err != nil {
				msg := "Can't load BPMN from serialized data"
				return state, &BpmnEngineUnmarshallingError{
					Msg: msg,
					Err: err,
				}
			}
		}
	}
	if eng.ProcessInstances != nil {
		state.processInstances = eng.ProcessInstances
		err := recoverProcessInstances(&state)
		if err != nil {
			return state, err
		}
	}
	recoverProcessInstanceActivitiesPart2(&state)
	if eng.MessageSubscriptions != nil {
		state.messageSubscriptions = eng.MessageSubscriptions
		recoverMessageSubscriptions(&state)
	}
	if eng.Timers != nil {
		state.timers = eng.Timers
		recoverTimers(&state)
	}
	if eng.Jobs != nil {
		state.jobs = eng.Jobs
		recoverJobs(&state)
	}
	return state, nil
}

func recoverProcessInstanceActivitiesPart1(pii *processInstanceInfo, adapter *processInstanceInfoAdapter) {
	for _, aa := range adapter.ActivityAdapters {
		switch aa.Type {
		case gatewayActivityAdapterType:
			var elementPlaceholder BPMN20.BaseElement = &baseElementPlaceholder{id: aa.ElementReference}
			pii.activities = append(pii.activities, &gatewayActivity{
				key:                     aa.Key,
				state:                   aa.State,
				element:                 &elementPlaceholder,
				parallel:                aa.Parallel,
				inboundFlowIdsCompleted: aa.InboundFlowIdsCompleted,
			})
		case eventBasedGatewayActivityAdapterType:
			var elementPlaceholder BPMN20.BaseElement = &baseElementPlaceholder{id: aa.ElementReference}
			pii.activities = append(pii.activities, &eventBasedGatewayActivity{
				key:                       aa.Key,
				state:                     aa.State,
				element:                   &elementPlaceholder,
				OutboundActivityCompleted: aa.OutboundActivityCompleted,
			})
		default:
			panic(fmt.Sprintf("[invariant check] missing recovery code for actictyAdapter.Type=%d", aa.Type))
		}
	}
}

func recoverProcessInstanceActivitiesPart2(state *BpmnEngineState) {
	for _, pi := range state.processInstances {
		for _, a := range pi.activities {
			switch activity := a.(type) {
			case *eventBasedGatewayActivity:
				activity.element = BPMN20.FindBaseElementsById(&pi.ProcessInfo.definitions, (*a.Element()).GetId())[0]
			case *gatewayActivity:
				activity.element = BPMN20.FindBaseElementsById(&pi.ProcessInfo.definitions, (*a.Element()).GetId())[0]
			default:
				panic(fmt.Sprintf("[invariant check] missing case for activity type=%T", a))
			}
		}
	}
}

// ----------------------------------------------------------------------------

func recoverProcessInstances(state *BpmnEngineState) error {
	for i, pi := range state.processInstances {
		process := state.findProcess(pi.ProcessInfo.ProcessKey)
		if process == nil {
			msg := fmt.Sprintf("Can't find process key %d in current BPMN Engine's processes", pi.ProcessInfo.ProcessKey)
			return &BpmnEngineUnmarshallingError{
				Msg: msg,
			}
		}
		state.processInstances[i].ProcessInfo = process
		state.processInstances[i].VariableHolder = var_holder.New(nil, nil)
	}
	return nil
}

func recoverJobs(state *BpmnEngineState) {
	for _, j := range state.jobs {
		definitions := state.FindProcessInstance(j.ProcessInstanceKey).ProcessInfo.definitions
		element := BPMN20.FindBaseElementsById(&definitions, j.ElementId)[0]
		j.baseElement = element
	}
}

func recoverTimers(state *BpmnEngineState) {
	for _, t := range state.timers {
		pi := state.FindProcessInstance(t.ProcessInstanceKey)
		if pi == nil {
			// FIXME: return error
		}
		t.baseElement = BPMN20.FindBaseElementsById(&pi.ProcessInfo.definitions, t.ElementId)[0]
		t.originActivity = pi.findActivity(t.originActivity.Key())
	}
}

func recoverMessageSubscriptions(state *BpmnEngineState) {
	for _, ms := range state.messageSubscriptions {
		pi := state.FindProcessInstance(ms.ProcessInstanceKey)
		if pi == nil {
			// FIXME: return error
		}
		ms.baseElement = BPMN20.FindBaseElementsById(&pi.ProcessInfo.definitions, ms.ElementId)[0]
		ms.originActivity = pi.findActivity(ms.originActivity.Key())
	}
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
