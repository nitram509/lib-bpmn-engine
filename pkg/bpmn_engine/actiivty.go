package bpmn_engine

import "github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"

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

type Activity interface {
	Key() int64
	State() ActivityState
	Element() *BPMN20.BaseElement
}

type tActivity struct {
	key     int64
	state   ActivityState
	element *BPMN20.BaseElement
}

func (a tActivity) Key() int64 {
	return a.key
}

func (a tActivity) State() ActivityState {
	return a.state
}

func (a tActivity) Element() *BPMN20.BaseElement {
	return a.element
}

// ------------------------

type GatewayActivity interface {
	Activity
	IsParallel() bool
	AreInboundFlowsCompleted() bool
	SetInboundFlowIdCompleted(id string)
	SetState(completed ActivityState)
}

type tGatewayActivity struct {
	key                     int64
	state                   ActivityState
	element                 *BPMN20.BaseElement
	parallel                bool
	inboundFlowIdsCompleted []string
}

func (ga *tGatewayActivity) Key() int64 {
	return ga.key
}

func (ga *tGatewayActivity) State() ActivityState {
	return ga.state
}

func (ga *tGatewayActivity) Element() *BPMN20.BaseElement {
	return ga.element
}

func (ga *tGatewayActivity) IsParallel() bool {
	return ga.parallel
}

func (ga *tGatewayActivity) AreInboundFlowsCompleted() bool {
	for _, association := range (*ga.element).GetIncomingAssociation() {
		if !contains(ga.inboundFlowIdsCompleted, association) {
			return false
		}
	}
	return true
}

func (ga *tGatewayActivity) SetInboundFlowIdCompleted(flowId string) {
	ga.inboundFlowIdsCompleted = append(ga.inboundFlowIdsCompleted, flowId)
}

func (ga *tGatewayActivity) SetState(state ActivityState) {
	ga.state = state
}

// -----------

type EventBasedGatewayActivity interface {
	Activity
	SetOutboundCompleted(id string)
	OutboundCompleted() bool
}

type tEventBasedGatewayActivity struct {
	key                       int64
	state                     ActivityState
	element                   *BPMN20.BaseElement
	outboundActivityCompleted string
}

func (ebg *tEventBasedGatewayActivity) Key() int64 {
	return ebg.key
}

func (ebg *tEventBasedGatewayActivity) State() ActivityState {
	return ebg.state
}

func (ebg *tEventBasedGatewayActivity) Element() *BPMN20.BaseElement {
	return ebg.element
}

func (ebg *tEventBasedGatewayActivity) SetOutboundCompleted(id string) {
	ebg.outboundActivityCompleted = id
}

func (ebg *tEventBasedGatewayActivity) OutboundCompleted() bool {
	return len(ebg.outboundActivityCompleted) > 0
}
