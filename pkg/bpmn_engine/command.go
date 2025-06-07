package bpmn_engine

import "github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"

type commandType string

const (
	flowTransitionType   commandType = "flowTransition"
	activityType         commandType = "activity"
	continueActivityType commandType = "continueActivity"
	// A command that there is a technical error and the engine should fail the process instance
	errorType                     commandType = "error"
	eventSubProcessCompletedType  commandType = "eventSubProcessCompletedType"
	checkExclusiveGatewayDoneType commandType = "checkExclusiveGatewayDone"
)

type command interface {
	Type() commandType
}

// ---------------------------------------------------------------------

type flowTransitionCommand struct {
	sourceId       string
	sourceActivity activity
	sequenceFlowId string
}

func (f flowTransitionCommand) Type() commandType {
	return flowTransitionType
}

// ---------------------------------------------------------------------

type activityCommand struct {
	sourceId       string
	element        *BPMN20.BaseElement
	originActivity activity
}

func (a activityCommand) Type() commandType {
	return activityType
}

// ---------------------------------------------------------------------

type continueActivityCommand struct {
	activity       activity
	originActivity activity
}

func (ga continueActivityCommand) Type() commandType {
	return continueActivityType
}

// ---------------------------------------------------------------------

type errorCommand struct {
	err         error
	elementId   string
	elementName string
}

func (e errorCommand) Type() commandType {
	return errorType
}

// ---------------------------------------------------------------------

type checkExclusiveGatewayDoneCommand struct {
	gatewayActivity eventBasedGatewayActivity
}

func (t checkExclusiveGatewayDoneCommand) Type() commandType {
	return checkExclusiveGatewayDoneType
}

type eventSubProcessCompletedCommand struct {
	// activity reference to the event sub-process, which has completed
	activity activity
}

func (t eventSubProcessCompletedCommand) Type() commandType {
	return eventSubProcessCompletedType
}
