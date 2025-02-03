package bpmn

import (
	"encoding/json"

	"github.com/pbinitiative/zenbpm/pkg/bpmn/model/bpmn20"
)

// ActivityState as per BPMN 2.0 spec, section 13.2.2 Activity, page 428
// State diagram (just partially shown):
//
//	 ┌─────┐
//	 │Ready│
//	 └──┬──┘
//	    |
//	┌───▽──┐
//	│Active│
//	└───┬──┘
//	    |
//	┌───▽──────┐
//	│Completing│
//	└────┬─────┘
//	     |
//	┌────▽────┐
//	│Completed│
//	└─────────┘
type ActivityState string

const (
	Active       ActivityState = "ACTIVE"
	Compensated  ActivityState = "COMPENSATED"
	Compensating ActivityState = "COMPENSATING"
	Completed    ActivityState = "COMPLETED"
	Completing   ActivityState = "COMPLETING"
	Failed       ActivityState = "FAILED"
	Failing      ActivityState = "FAILING"
	Ready        ActivityState = "READY"
	Terminated   ActivityState = "TERMINATED"
	Terminating  ActivityState = "TERMINATING"
	WithDrawn    ActivityState = "WITHDRAWN"
)

type activity interface {
	Key() int64
	State() ActivityState
	Element() *bpmn20.BaseElement
}

type elementActivity struct {
	key     int64         `json:"k"`
	state   ActivityState `json:"s"`
	element *bpmn20.BaseElement
}

func (a elementActivity) Key() int64 {
	return a.key
}

func (a elementActivity) State() ActivityState {
	return a.state
}

func (a elementActivity) Element() *bpmn20.BaseElement {
	return a.element
}

// -------------------------------------------------------------------------

type gatewayActivity struct {
	key                     int64         `json:"k"`
	state                   ActivityState `json:"s"`
	element                 *bpmn20.BaseElement
	parallel                bool
	inboundFlowIdsCompleted []string
}

func (ga *gatewayActivity) Key() int64 {
	return ga.key
}

func (ga *gatewayActivity) State() ActivityState {
	return ga.state
}

func (ga *gatewayActivity) Element() *bpmn20.BaseElement {
	return ga.element
}

func (ga *gatewayActivity) AreInboundFlowsCompleted() bool {
	for _, association := range (*ga.element).GetIncomingAssociation() {
		if !contains(ga.inboundFlowIdsCompleted, association) {
			return false
		}
	}
	return true
}

func (ga *gatewayActivity) SetInboundFlowCompleted(flowId string) {
	ga.inboundFlowIdsCompleted = append(ga.inboundFlowIdsCompleted, flowId)
}

func (ga *gatewayActivity) SetState(state ActivityState) {
	ga.state = state
}

func (ga gatewayActivity) MarshalJSON() ([]byte, error) {
	type Alias gatewayActivity // Create an alias to avoid infinite recursion
	return json.Marshal(&struct {
		Key                     int64         `json:"key"`
		State                   ActivityState `json:"state"`
		ElementID               string        `json:"elementId"`
		Parallel                bool          `json:"parallel"`
		InboundFlowIdsCompleted []string      `json:"inboundFlowIdsCompleted"`
	}{
		Key:                     ga.key,
		State:                   ga.state,
		ElementID:               (*ga.element).GetId(), // Get the ID from the element
		Parallel:                ga.parallel,
		InboundFlowIdsCompleted: ga.inboundFlowIdsCompleted,
	})
}

// -------------------------------------------------------------------------

type eventBasedGatewayActivity struct {
	key                       int64
	state                     ActivityState
	element                   *bpmn20.BaseElement
	OutboundActivityCompleted string
}

func (ebg *eventBasedGatewayActivity) Key() int64 {
	return ebg.key
}

func (ebg *eventBasedGatewayActivity) State() ActivityState {
	return ebg.state
}

func (ebg *eventBasedGatewayActivity) Element() *bpmn20.BaseElement {
	return ebg.element
}

func (ebg *eventBasedGatewayActivity) SetOutboundCompleted(id string) {
	ebg.OutboundActivityCompleted = id
}

func (ebg *eventBasedGatewayActivity) OutboundCompleted() bool {
	return len(ebg.OutboundActivityCompleted) > 0
}
