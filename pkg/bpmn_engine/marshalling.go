package bpmn_engine

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"github.com/pbinitiative/feel"
	"reflect"
	"strings"
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

func (a activitySurrogate) Element() *BPMN20.BaseElement {
	return a.elementReference
}

// ----------------------------------------------------------------------------

// marshalOptions Options that will be used while marshalling the engine
type marshalOptions struct {
	exportTypes bool
}

// MarshalOption is a function that modifies the marshalOptions
type MarshalOption func(*marshalOptions) error

// WithMarshalComplexTypes if added as an option the marshaller will export variables with their specific types.
// When this is used, calls to Unmarshal will need to use RegisterType to configure the types that can be
// unmarshalled into actual instances
// .
// This is useful when you have complex types in your variables that you want to preserve.
//
//	Example:
//	```go
//	// Marshal with type information
//	data := engine.Marshal(WithMarshalComplexTypes())
//
//	// Unmarshal with type mapping
//	engine, _ = Unmarshal(data, RegisterType(MyStruct{}))
//	```
func WithMarshalComplexTypes() MarshalOption {
	return func(opts *marshalOptions) error {
		opts.exportTypes = true
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

// complexVariable this struct is used when marshalling with WithMarshalComplexTypes to export complex types
type complexVariable struct {
	Type  string `json:"_t"`
	Value string `json:"v"`
}

// newComplexVariable creates a new complexVariable from the given value.
// The type is derived from the type of the given value. If the type is a pointer, it will be prefixed with "*".
// The value is the value itself.
func newComplexVariable(v any) (*complexVariable, error) {
	t := reflect.TypeOf(v)
	prefix := ""
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		prefix = "*"
	}

	valStr, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return &complexVariable{
		Type:  fmt.Sprintf("%s%s.%s", prefix, t.PkgPath(), t.Name()),
		Value: string(valStr),
	}, nil
}

// createComplexVariables takes a variable holder and wraps each variable with a complexVariable if the variable is a
// struct or pointer
func createComplexVariables(vh VariableHolder) (VariableHolder, error) {
	for k, v := range vh.variables {
		kind := reflect.ValueOf(v).Kind()
		if kind == reflect.Struct || kind == reflect.Ptr {
			vars, err := newComplexVariable(v)
			if err != nil {
				return VariableHolder{}, err
			}
			vh.variables[k] = vars
		}
	}
	// If there is a parent, create complex variables for it as well
	if vh.parent != nil {
		parent, err := createComplexVariables(*vh.parent)
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
	if !opts.exportTypes {
		return pii, nil
	}

	// Create complex variables for each process instance
	for _, pi := range pii {
		cvs, err := createComplexVariables(pi.VariableHolder)
		if err != nil {
			return nil, err
		}
		pi.VariableHolder = cvs
	}
	return pii, nil
}

// variableTypeMapping defines a type that can be used to map type keys to actual relection types
type variableTypeMapping map[string]reflect.Type

// unmarshalOptions Options that will be used while unmarshalling the engine
type unmarshalOptions struct {
	exportTypes bool
	typeMapping variableTypeMapping
}

// UnmarshalOption is a function that modifies the unmarshalOptions
type UnmarshalOption func(*unmarshalOptions) error

// registerTypeOption is a function that registers a type in the type mapping
type registerTypeOption func(map[string]reflect.Type) error

// WithUnmarshalComplexTypes enables type unmarshalling.
// Additional types can be registered using the RegisterType function.
func WithUnmarshalComplexTypes(at ...registerTypeOption) UnmarshalOption {
	// All default complex types that the engine may produce must be registered by default here
	mappingOptions := []registerTypeOption{
		RegisterType(feel.FEELDuration{}),
		RegisterType(feel.FEELDatetime{}),
		RegisterType(feel.FEELDate{}),
		RegisterType(feel.FEELTime{}),
		RegisterType(feel.NullValue{}),
		RegisterType(feel.Number{}),
	}
	mappingOptions = append(mappingOptions, at...)

	return func(opts *unmarshalOptions) error {
		opts.exportTypes = true

		for _, mo := range mappingOptions {
			err := mo(opts.typeMapping)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// RegisterType registers a type that can be unmarshalled into an instance of the given type.
func RegisterType(instance any) func(map[string]reflect.Type) error {
	return func(m map[string]reflect.Type) error {
		t := reflect.TypeOf(instance)

		// Do not allow pointers or any other basic types to be passed in as an instance type
		// Marshalling and Unmarshalling will take care of pointers
		if t.Kind() != reflect.Struct {
			return errors.New("only instance of structs should be used")
		}

		typeKey := fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
		m[typeKey] = t
		return nil
	}
}

// applyUnmarshalOptions Applies the given options and returns the applied unmarshalOptions
func applyUnmarshalOptions(options ...UnmarshalOption) (*unmarshalOptions, error) {
	opts := &unmarshalOptions{
		typeMapping: make(map[string]reflect.Type),
	}
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
				activity.element = BPMN20.FindBaseElementsById(&pi.ProcessInfo.definitions, (*a.Element()).GetId())[0]
			case *gatewayActivity:
				activity.element = BPMN20.FindBaseElementsById(&pi.ProcessInfo.definitions, (*a.Element()).GetId())[0]
			default:
				return fmt.Errorf("[invariant check] missing case for activity type=%T", a)
			}
		}
	}
	return nil
}

// ----------------------------------------------------------------------------

// newInstance Create a new instance given a type name
func newInstance(typeName string, opts *unmarshalOptions) any {
	if strings.HasPrefix(typeName, "*") {
		// Remove pointer from type name
		typeName = typeName[1:]
	}

	t, exists := opts.typeMapping[typeName]
	if !exists {
		fmt.Printf("could not find %s\n", typeName)
		return nil // Type not found
	}
	return reflect.New(t).Interface() // Create a new instance (as pointer)
}

// removePointer Function to remove pointer from an `any` type variable
func removePointer(v any) (any, bool) {
	// Use type assertion to check if it's a pointer
	if ptr, ok := v.(interface{ Elem() any }); ok {
		return ptr.Elem(), true
	}

	// Use reflection as a fallback for generic cases
	return removePointerReflect(v)
}

// removePointerReflect Function to remove pointer from an `any` type variable
// using reflection
func removePointerReflect(v any) (any, bool) {
	// Use reflection to handle arbitrary pointer types
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		return rv.Elem().Interface(), true
	}
	return v, false
}

// recoverVariableInstances recovers the variable instances from the given VariableHolder
func recoverVariableInstances(vh VariableHolder, opts *unmarshalOptions) (VariableHolder, error) {
	if len(opts.typeMapping) == 0 {
		// Nothing additional to do
		return vh, nil
	}

	for k, v := range vh.variables {
		if m, ok := v.(map[string]any); ok {
			typeKey, typeKeyOk := m["_t"]
			value, valueOk := m["v"]
			var valueBytes []byte

			if strVal, isString := value.(string); isString {
				valueBytes = []byte(strVal)
			} else {
				valueOk = false
			}

			// if map has a "_t" key, it's a wrapped variable
			if typeKeyOk && valueOk {
				isPointer := false
				instanceType := typeKey.(string)
				if strings.HasPrefix(instanceType, "*") {
					// Remove pointer from type name
					instanceType = instanceType[1:]
					isPointer = true
				}

				// creates a new instance with the type from the map
				instance := newInstance(instanceType, opts)
				if instance == nil {
					return vh, fmt.Errorf("unmarshalling unknown type %s", instanceType)
				}

				err := json.Unmarshal(valueBytes, instance)

				if err != nil {
					return vh, err
				}

				if !isPointer {
					if val, ok := removePointer(instance); ok {
						instance = val
					} else {
						return vh, errors.New("could not remove pointer")
					}
				}

				// Replace the variable with the proper instance
				vh.variables[k] = instance
			}

		}
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
		element := BPMN20.FindBaseElementsById(&definitions, j.ElementId)[0]
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
		t.baseElement = BPMN20.FindBaseElementsById(&pi.ProcessInfo.definitions, t.ElementId)[0]
		availableOriginActivity := pi.findActivity(t.originActivity.Key())
		if availableOriginActivity != nil {
			t.originActivity = availableOriginActivity
		} else {
			originActivitySurrogate := t.originActivity.(activitySurrogate)
			originActivitySurrogate.elementReference = BPMN20.FindBaseElementsById(&pi.ProcessInfo.definitions, originActivitySurrogate.ElementReferenceId)[0]
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
		ms.baseElement = BPMN20.FindBaseElementsById(&pi.ProcessInfo.definitions, ms.ElementId)[0]
		availableOriginActivity := pi.findActivity(ms.originActivity.Key())
		if availableOriginActivity != nil {
			ms.originActivity = availableOriginActivity
		} else {
			originActivitySurrogate := ms.originActivity.(activitySurrogate)
			originActivitySurrogate.elementReference = BPMN20.FindBaseElementsById(&pi.ProcessInfo.definitions, originActivitySurrogate.ElementReferenceId)[0]
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
