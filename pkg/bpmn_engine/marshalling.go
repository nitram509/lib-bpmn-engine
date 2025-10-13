package bpmn_engine

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
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

type timerAlias Timer
type timerAdapter struct {
	OriginActivitySurrogate activitySurrogate `json:"oas"`
	*timerAlias
}

type messageSubscriptionAlias MessageSubscription
type messageSubscriptionAdapter struct {
	OriginActivitySurrogate activitySurrogate `json:"oas"`
	*messageSubscriptionAlias
}

type variableHolderAdapter struct {
	Parent    *VariableHolder        `json:"p,omitempty"`
	Variables map[string]interface{} `json:"v,omitempty"`
}

func (vh *VariableHolder) MarshalJSON() ([]byte, error) {

	adapter := variableHolderAdapter{
		Parent:    vh.parent,
		Variables: vh.variables,
	}
	return json.Marshal(adapter)
}

func (vh *VariableHolder) UnmarshalJSON(data []byte) error {
	vha := variableHolderAdapter{}
	if err := json.Unmarshal(data, &vha); err != nil {
		return err
	}
	vh.parent = vha.Parent

	vars := vha.Variables
	if vars == nil {
		vars = make(map[string]interface{})
	}

	vh.variables = vars
	return nil
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

// activitySurrogate only exists to have a simple way of marshalling originActivities in MessageSubscription and Timer
// TODO see issue https://github.com/nitram509/lib-bpmn-engine/issues/190
type activitySurrogate struct {
	ActivityKey        int64         `json:"k"`
	ActivityState      ActivityState `json:"s"`
	ElementReferenceId string        `json:"e"`
	elementReference   *BPMN20.BaseElement
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

func (a activityPlaceholder) SetState(state ActivityState) {
	panic("the placeholder does not implement all methods, by intent")
}

func (a activityPlaceholder) Element() *BPMN20.BaseElement {
	panic("the placeholder does not implement all methods, by intent")
}

// ----------------------------------------------------------------------------

func (t *Timer) MarshalJSON() ([]byte, error) {
	ta := &timerAdapter{
		timerAlias: (*timerAlias)(t),
	}
	// TODO see issue https://github.com/nitram509/lib-bpmn-engine/issues/190
	ta.OriginActivitySurrogate = activitySurrogate{
		ActivityKey:        t.originActivity.Key(),
		ActivityState:      t.originActivity.State(),
		ElementReferenceId: (*t.originActivity.Element()).GetId(),
	}
	return json.Marshal(ta)
}

func (t *Timer) UnmarshalJSON(data []byte) error {
	ta := timerAdapter{
		timerAlias: (*timerAlias)(t),
	}
	if err := json.Unmarshal(data, &ta); err != nil {
		return err
	}
	t.originActivity = ta.OriginActivitySurrogate
	return nil
}

// ----------------------------------------------------------------------------

func (m *MessageSubscription) MarshalJSON() ([]byte, error) {
	msa := &messageSubscriptionAdapter{
		messageSubscriptionAlias: (*messageSubscriptionAlias)(m),
	}
	// TODO see issue https://github.com/nitram509/lib-bpmn-engine/issues/190
	msa.OriginActivitySurrogate = activitySurrogate{
		ActivityKey:        m.originActivity.Key(),
		ActivityState:      m.originActivity.State(),
		ElementReferenceId: (*m.originActivity.Element()).GetId(),
	}
	return json.Marshal(msa)
}

func (m *MessageSubscription) UnmarshalJSON(data []byte) error {
	msa := messageSubscriptionAdapter{
		messageSubscriptionAlias: (*messageSubscriptionAlias)(m),
	}
	if err := json.Unmarshal(data, &msa); err != nil {
		return err
	}
	m.originActivity = msa.OriginActivitySurrogate
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
			return nil, fmt.Errorf("[invariant check] missing activity adapter for the type %T", a)
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
	pii.VariableHolder = adapter.VariableHolder
	return recoverProcessInstanceActivitiesPart1(pii, adapter)
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

func (a activitySurrogate) Key() int64 {
	return a.ActivityKey
}

func (a activitySurrogate) State() ActivityState {
	return a.ActivityState
}

func (a activitySurrogate) SetState(state ActivityState) {
	a.ActivityState = state
}

func (a activitySurrogate) Element() *BPMN20.BaseElement {
	return a.elementReference
}

// ----------------------------------------------------------------------------

// VariableWrapFunc function to wrap variables before marshalling
//
// Parameters:
// - key: Variables key
// - value: Wrapped value of the variable
type VariableWrapFunc func(string, any) (any, error)

// marshalOptions Options that will be used while marshalling the engine
type marshalOptions struct {
	marshalVariablesFunc VariableWrapFunc
}

// MarshalOption is a function that modifies the marshalOptions
type MarshalOption func(*marshalOptions) error

// WithMarshalVariableFunc sets a function that will be called for each variable in the engine's VarHolder
// This allows you to customize variables before they are marshalled, e.g. to convert them to a different type
func WithMarshalVariableFunc(fun VariableWrapFunc) MarshalOption {
	return func(opts *marshalOptions) error {
		opts.marshalVariablesFunc = fun
		return nil
	}
}

// applyMarshalOptions Applies the given options and returns the applied marshalOptions
func applyMarshalOptions(options ...MarshalOption) (*marshalOptions, error) {
	opts := &marshalOptions{}
	for _, o := range options {
		err := o(opts)
		if err != nil {
			return nil, fmt.Errorf("could not apply option: %w", err)
		}
	}

	return opts, nil
}

// Marshal marshals the engine into a byte array.
// Options may be provided to configure the marshalling process.
// It returns a byte array containing the marshalled engine state.
// If there is an error applying the options, it will panic.
//
// Example:
//
//	```go
//	// Marshal with default options
//	data, err := bpmn_engine.Marshal()
//
//	// Marshal with type information for complex variables
//	data, err := bpmn_engine.Marshal(WithMarshalComplexTypes())
//	```
func (state *BpmnEngineState) Marshal(options ...MarshalOption) ([]byte, error) {
	opts, err := applyMarshalOptions(options...)
	if err != nil {
		return nil, err
	}
	pis, err := createProcessInstances(state.processInstances, opts)
	if err != nil {
		return nil, err
	}

	m := serializedBpmnEngine{
		Version:              CurrentSerializerVersion,
		Name:                 state.name,
		MessageSubscriptions: state.messageSubscriptions,
		ProcessReferences:    createReferences(state.processes),
		ProcessInstances:     pis,
		Timers:               state.timers,
		Jobs:                 state.jobs,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// wrapVariables takes a variable holder and wraps each variable with a complexVariable if the variable is a
// struct or pointer
func wrapVariables(vh VariableHolder, f VariableWrapFunc) (VariableHolder, error) {
	for k, v := range vh.variables {
		val, err := f(k, v)
		if err != nil {
			return vh, err
		}
		vh.variables[k] = val
	}
	// If there is a parent, create complex variables for it as well
	if vh.parent != nil {
		parent, err := wrapVariables(*vh.parent, f)
		if err != nil {
			return VariableHolder{}, err
		}
		vh.parent = &parent
	}
	return vh, nil
}

// createProcessInstances Creates process instances that can be marshalled to JSON
func createProcessInstances(pii []*processInstanceInfo, opts *marshalOptions) ([]*processInstanceInfo, error) {

	// If exporting types is not enable, there is nothing extra to do
	if opts.marshalVariablesFunc == nil {
		return pii, nil
	}

	// Create complex variables for each process instance
	for _, pi := range pii {
		cvs, err := wrapVariables(pi.VariableHolder, opts.marshalVariablesFunc)
		if err != nil {
			return nil, err
		}
		pi.VariableHolder = cvs
	}
	return pii, nil
}

// VariableUnwrapFunc function to unwrap variables from marshalled state
//
// Parameters:
// - key: Variables key
// - value: Wrapped value of the variable
type VariableUnwrapFunc func(key string, value any) (any, error)

// unmarshalOptions Options that will be used while unmarshalling the engine
type unmarshalOptions struct {
	// variableUnwrapFunc function that can be called to restore a variable from a marshalled state
	variableUnwrapFunc VariableUnwrapFunc
}

// UnmarshalOption is a function that modifies the unmarshalOptions
type UnmarshalOption func(*unmarshalOptions) error

// WithUnmarshalVariableFunc sets a function that will be called for each variable in the engine's VarHolder
// This allows you to customize variables after they are unmarshalled, e.g. to convert them to a different type
func WithUnmarshalVariableFunc(fun VariableUnwrapFunc) UnmarshalOption {
	return func(opts *unmarshalOptions) error {
		opts.variableUnwrapFunc = fun
		return nil
	}
}

// applyUnmarshalOptions Applies the given options and returns the applied unmarshalOptions
func applyUnmarshalOptions(options ...UnmarshalOption) (*unmarshalOptions, error) {
	opts := &unmarshalOptions{}
	for _, o := range options {
		err := o(opts)
		if err != nil {
			return nil, fmt.Errorf("could not apply option: %w", err)
		}
	}

	return opts, nil
}

// Unmarshal loads the data byte array and creates a new instance of the BPMN Engine
// Will return an BpmnEngineUnmarshallingError, if there was an issue AND in case of error,
// the engine return object is only partially initialized and likely not usable
func Unmarshal(data []byte, opts ...UnmarshalOption) (BpmnEngineState, error) {

	// Build an unmarshalOptions object from the provided options
	options, err := applyUnmarshalOptions(opts...)
	if err != nil {
		return BpmnEngineState{}, &BpmnEngineUnmarshallingError{
			Msg: "Failed to apply unmarshalling options",
			Err: err,
		}
	}

	eng := serializedBpmnEngine{}
	err = json.Unmarshal(data, &eng)
	if err != nil {
		return BpmnEngineState{}, &BpmnEngineUnmarshallingError{
			Msg: "Failed to unmarshall engine data",
			Err: err,
		}
	}
	state := New()
	state.name = eng.Name
	if eng.ProcessReferences != nil {
		for _, pir := range eng.ProcessReferences {
			xmlData, err := decodeAndDecompress(pir.BpmnData)
			if err != nil {
				msg := "Can't decode nor decompress serialized BPMN data"
				return state, &BpmnEngineUnmarshallingError{
					Msg: msg,
					Err: err,
				}
			}
			process, err := state.load(xmlData, pir.BpmnResourceName)
			if err != nil {
				msg := "Can't load BPMN from serialized data"
				return state, &BpmnEngineUnmarshallingError{
					Msg: msg,
					Err: err,
				}
			}
			process.ProcessKey = pir.ProcessKey
		}
	}
	if eng.ProcessInstances != nil {
		state.processInstances = eng.ProcessInstances
		err = recoverProcessInstances(&state, options)
		if err != nil {
			return state, err
		}
	}
	err = recoverProcessInstanceActivitiesPart2(&state)
	if err != nil {
		return BpmnEngineState{}, err
	}
	if eng.MessageSubscriptions != nil {
		state.messageSubscriptions = eng.MessageSubscriptions
		err = recoverMessageSubscriptions(&state)
		if err != nil {
			return state, err
		}
	}
	if eng.Timers != nil {
		state.timers = eng.Timers
		err = recoverTimers(&state)
		if err != nil {
			return state, err
		}
	}
	if eng.Jobs != nil {
		state.jobs = eng.Jobs
		err = recoverJobs(&state)
		if err != nil {
			return state, err
		}
	}
	return state, nil
}

func recoverProcessInstanceActivitiesPart1(pii *processInstanceInfo, adapter *processInstanceInfoAdapter) error {
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
			return fmt.Errorf("[invariant check] missing recovery code for actictyAdapter.Type=%d", aa.Type)
		}
	}
	return nil
}

func recoverProcessInstanceActivitiesPart2(state *BpmnEngineState) error {
	for _, pi := range state.processInstances {
		for _, a := range pi.activities {
			switch activity := a.(type) {
			case *eventBasedGatewayActivity:
				activity.element = BPMN20.FindBaseElementsById(pi.ProcessInfo.definitions.Process, (*a.Element()).GetId())[0]
			case *gatewayActivity:
				activity.element = BPMN20.FindBaseElementsById(pi.ProcessInfo.definitions.Process, (*a.Element()).GetId())[0]
			default:
				return fmt.Errorf("[invariant check] missing case for activity type=%T", a)
			}
		}
	}
	return nil
}

// recoverVariableInstances recovers the variable instances from the given VariableHolder
func recoverVariableInstances(vh VariableHolder, opts *unmarshalOptions) (VariableHolder, error) {
	if opts.variableUnwrapFunc == nil {
		// Nothing additional to do
		return vh, nil
	}

	for k, v := range vh.variables {
		val, err := opts.variableUnwrapFunc(k, v)
		if err != nil {
			return vh, err
		}

		// Replace the variable with the proper instance
		vh.variables[k] = val
	}
	return vh, nil
}

func recoverProcessInstances(state *BpmnEngineState, opts *unmarshalOptions) error {
	for i, pi := range state.processInstances {
		process := state.findProcess(pi.ProcessInfo.ProcessKey)
		if process == nil {
			msg := fmt.Sprintf("Can't find process key %d in current BPMN Engine's processes", pi.ProcessInfo.ProcessKey)
			return &BpmnEngineUnmarshallingError{
				Msg: msg,
			}
		}
		state.processInstances[i].ProcessInfo = process
		vars, err := recoverVariableInstances(pi.VariableHolder, opts)
		if err != nil {
			return err
		}
		state.processInstances[i].VariableHolder = vars
	}
	return nil
}

func recoverJobs(state *BpmnEngineState) error {
	for _, j := range state.jobs {
		pi := state.FindProcessInstance(j.ProcessInstanceKey)
		if pi == nil {
			return &BpmnEngineUnmarshallingError{
				Msg: fmt.Sprintf("can't find process instannce with key %d; "+
					"the marshalled JSON was likely corrupt", j.ProcessInstanceKey),
			}
		}
		definitions := pi.ProcessInfo.definitions
		element := BPMN20.FindBaseElementsById(definitions.Process, j.ElementId)[0]
		j.baseElement = element
	}
	return nil
}

func recoverTimers(state *BpmnEngineState) error {
	for _, t := range state.timers {
		pi := state.FindProcessInstance(t.ProcessInstanceKey)
		if pi == nil {
			return &BpmnEngineUnmarshallingError{
				Msg: fmt.Sprintf("can't find process instannce with key %d; "+
					"the marshalled JSON was likely corrupt", t.ProcessInstanceKey),
			}
		}
		t.baseElement = BPMN20.FindBaseElementsById(pi.ProcessInfo.definitions.Process, t.ElementId)[0]
		availableOriginActivity := pi.findActivity(t.originActivity.Key())
		if availableOriginActivity != nil {
			t.originActivity = availableOriginActivity
		} else {
			originActivitySurrogate := t.originActivity.(activitySurrogate)
			originActivitySurrogate.elementReference = BPMN20.FindBaseElementsById(pi.ProcessInfo.definitions.Process, originActivitySurrogate.ElementReferenceId)[0]
			t.originActivity = originActivitySurrogate
		}
	}
	return nil
}

func recoverMessageSubscriptions(state *BpmnEngineState) error {
	for _, ms := range state.messageSubscriptions {
		pi := state.FindProcessInstance(ms.ProcessInstanceKey)
		if pi == nil {
			return &BpmnEngineUnmarshallingError{
				Msg: fmt.Sprintf("can't find process instannce with key %d; "+
					"the marshalled JSON was likely corrupt", ms.ProcessInstanceKey),
			}
		}
		ms.baseElement = BPMN20.FindBaseElementsById(pi.ProcessInfo.definitions.Process, ms.ElementId)[0]
		availableOriginActivity := pi.findActivity(ms.originActivity.Key())
		if availableOriginActivity != nil {
			ms.originActivity = availableOriginActivity
		} else {
			originActivitySurrogate := ms.originActivity.(activitySurrogate)
			originActivitySurrogate.elementReference = BPMN20.FindBaseElementsById(pi.ProcessInfo.definitions.Process, originActivitySurrogate.ElementReferenceId)[0]
			ms.originActivity = originActivitySurrogate
		}
	}
	return nil
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
