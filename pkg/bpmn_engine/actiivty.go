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

// FIXME: do we need a dedicated implementation? Or, shall we enrich e.g. processInstanceInfo
//type activity struct {
//	key     int64
//	state   ActivityState
//	element *BPMN20.BaseElement
//}
//
//func (a activity) Key() int64 {
//	return a.key
//}
//
//func (a activity) State() ActivityState {
//	return a.state
//}
//
//func (a activity) Element() *BPMN20.BaseElement {
//	return a.element
//}
