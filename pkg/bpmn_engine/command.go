package bpmn_engine

import "github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"

type commandType string

const (
	flowTransitionType   commandType = "flow"
	activityType         commandType = "activity"
	continueActivityType commandType = "complete-activity"
	errorType            commandType = "error"
)

type command interface {
	Type() commandType
}

// -----------

type flowTransitionCommand interface {
	command
	OriginActivity() Activity
	SequenceFlowId() string
}

type tFlowTransitionCommand struct {
	sourceId       string
	sequenceFlowId string
	sourceActivity Activity
}

func (f tFlowTransitionCommand) Type() commandType {
	return flowTransitionType
}

func (f tFlowTransitionCommand) InboundFlowId() string {
	return f.sourceId
}

func (f tFlowTransitionCommand) OriginActivity() Activity {
	return f.sourceActivity
}

func (f tFlowTransitionCommand) SequenceFlowId() string {
	return f.sequenceFlowId
}

// -----------

type activityCommand interface {
	command
	Element() *BPMN20.BaseElement
	InboundFlowId() string
	OriginActivity() Activity
}

type tActivityCommand struct {
	sourceId       string
	element        *BPMN20.BaseElement
	originActivity Activity
}

func (a tActivityCommand) Type() commandType {
	return activityType
}

func (a tActivityCommand) InboundFlowId() string {
	return a.sourceId
}
func (a tActivityCommand) OriginActivity() Activity {
	return a.originActivity
}

func (a tActivityCommand) Element() *BPMN20.BaseElement {
	return a.element
}

// ------

type continueActivityCommand interface {
	activityCommand
	Activity() Activity
}

type tContinueActivityCommand struct {
	sourceId       string
	activity       *tActivity
	originActivity Activity
}

func (ga tContinueActivityCommand) OriginActivity() Activity {
	return ga.originActivity
}

func (ga tContinueActivityCommand) Type() commandType {
	return continueActivityType
}

func (ga tContinueActivityCommand) InboundFlowId() string {
	return ga.sourceId
}

func (ga tContinueActivityCommand) Element() *BPMN20.BaseElement {
	return (*ga.activity).Element()
}

func (ga tContinueActivityCommand) Activity() Activity {
	return ga.activity
}

// -------------

type ErrorCommand interface {
	command
	ElementId() string
	ElementName() string
	Error() error
}

type tErrorCommand struct {
	err         error
	elementId   string
	elementName string
}

func (e tErrorCommand) Type() commandType {
	return errorType
}

func (e tErrorCommand) ElementId() string {
	return e.elementId
}

func (e tErrorCommand) ElementName() string {
	return e.elementName
}

func (e tErrorCommand) Error() error {
	return e.err
}
