package BPMN20

import "github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/extensions"

type TDefinitions struct {
	Id                 string     `xml:"id,attr"`
	Name               string     `xml:"name,attr"`
	TargetNamespace    string     `xml:"targetNamespace,attr"`
	ExpressionLanguage string     `xml:"expressionLanguage,attr"`
	TypeLanguage       string     `xml:"typeLanguage,attr"`
	Exporter           string     `xml:"exporter,attr"`
	ExporterVersion    string     `xml:"exporterVersion,attr"`
	Process            TProcess   `xml:"process"`
	Messages           []TMessage `xml:"message"`
	Errors             []TError   `xml:"error"`
}

type TError struct {
	Id        string `xml:"id,attr"`
	Name      string `xml:"name,attr"`
	ErrorCode string `xml:"errorCode,attr"`
}

type TErrorEventDefinition struct {
	Id       string `xml:"id,attr"`
	ErrorRef string `xml:"errorRef,attr"`
}

type TCallableElement struct {
	TRootElement
	Name string `xml:"name,attr"`
}

type TProcess struct {
	TCallableElement
	ProcessType                  string                    `xml:"processType,attr"`
	IsClosed                     bool                      `xml:"isClosed,attr"`
	IsExecutable                 bool                      `xml:"isExecutable,attr"`
	DefinitionalCollaborationRef string                    `xml:"definitionalCollaborationRef,attr"`
	StartEvents                  []TStartEvent             `xml:"startEvent"`
	EndEvents                    []TEndEvent               `xml:"endEvent"`
	SequenceFlows                []TSequenceFlow           `xml:"sequenceFlow"`
	ServiceTasks                 []TServiceTask            `xml:"serviceTask"`
	UserTasks                    []TUserTask               `xml:"userTask"`
	SubProcesses                 []TSubProcess             `xml:"subProcess"`
	ParallelGateway              []TParallelGateway        `xml:"parallelGateway"`
	ExclusiveGateway             []TExclusiveGateway       `xml:"exclusiveGateway"`
	IntermediateCatchEvent       []TIntermediateCatchEvent `xml:"intermediateCatchEvent"`
	IntermediateTrowEvent        []TIntermediateThrowEvent `xml:"intermediateThrowEvent"`
	EventBasedGateway            []TEventBasedGateway      `xml:"eventBasedGateway"`
	InclusiveGateway             []TInclusiveGateway       `xml:"inclusiveGateway"`
	BoundaryEvent                []TBoundaryEvent          `xml:"boundaryEvent"`
}

type TBoundaryEvent struct {
	TBaseElement
	Name                 string                  `xml:"name,attr"`
	AttachedToRef        string                  `xml:"attachedToRef,attr"`
	OutgoingAssociation  []string                `xml:"outgoing"`
	ErrorEventDefinition *TErrorEventDefinition  `xml:"errorEventDefinition,omitempty"`
	Output               []extensions.TIoMapping `xml:"extensionElements>ioMapping>output"`
}

type TSubProcess struct {
	TActivity
	TriggeredByEvent       bool                      `xml:"triggeredByEvent,attr"`
	StartEvents            []TStartEvent             `xml:"startEvent"`
	EndEvents              []TEndEvent               `xml:"endEvent"`
	SequenceFlows          []TSequenceFlow           `xml:"sequenceFlow"`
	ServiceTasks           []TServiceTask            `xml:"serviceTask"`
	UserTasks              []TUserTask               `xml:"userTask"`
	SubProcesses           []TSubProcess             `xml:"subProcess"`
	ParallelGateway        []TParallelGateway        `xml:"parallelGateway"`
	ExclusiveGateway       []TExclusiveGateway       `xml:"exclusiveGateway"`
	IntermediateCatchEvent []TIntermediateCatchEvent `xml:"intermediateCatchEvent"`
	IntermediateTrowEvent  []TIntermediateThrowEvent `xml:"intermediateThrowEvent"`
	EventBasedGateway      []TEventBasedGateway      `xml:"eventBasedGateway"`
	InclusiveGateway       []TInclusiveGateway       `xml:"inclusiveGateway"`
	BoundaryEvent          []TBoundaryEvent          `xml:"boundaryEvent"`
}

// TBaseElement is an "abstract" struct
type TBaseElement struct {
	Id            string `xml:"id,attr"`
	Documentation string `xml:"documentation"`
}

// TRootElement is an "abstract" struct
type TRootElement struct {
	TBaseElement
}

// TEventDefinition is an "abstract" struct
type TEventDefinition struct {
	TRootElement
}

type TFlowElement struct {
	TBaseElement
	Name string `xml:"name,attr"`
}

type TGateway struct {
	TFlowNode
	GatewayDirection GatewayDirection `xml:"gatewayDirection,attr"`
}

type TFlowNode struct {
	TFlowElement
	IncomingAssociation []string `xml:"incoming"`
	OutgoingAssociation []string `xml:"outgoing"`
}

// TEvent is an "abstract" struct
type TEvent struct {
	TFlowNode
}

// TCatchEvent is an "abstract" struct
type TCatchEvent struct {
	TEvent
}

type TSequenceFlow struct {
	TFlowElement
	SourceRef           string        `xml:"sourceRef,attr"`
	TargetRef           string        `xml:"targetRef,attr"`
	ConditionExpression []TExpression `xml:"conditionExpression"`
}

type TExpression struct {
	Text string `xml:",innerxml"`
}

type TStartEvent struct {
	TCatchEvent
	IsInterrupting       bool                  `xml:"isInterrupting,attr"`
	ParallelMultiple     bool                  `xml:"parallelMultiple,attr"`
	ErrorEventDefinition TErrorEventDefinition `xml:"errorEventDefinition"`
}

type TEndEvent struct {
	TThrowEvent
	ErrorEventDefinition TErrorEventDefinition `xml:"errorEventDefinition"`
}

type TServiceTask struct {
	TTask
	Default            string                     `xml:"default,attr"`
	CompletionQuantity int                        `xml:"completionQuantity,attr"`
	IsForCompensation  bool                       `xml:"isForCompensation,attr"`
	OperationRef       string                     `xml:"operationRef,attr"`
	Implementation     string                     `xml:"implementation,attr"`
	Input              []extensions.TIoMapping    `xml:"extensionElements>ioMapping>input"`
	Output             []extensions.TIoMapping    `xml:"extensionElements>ioMapping>output"`
	TaskDefinition     extensions.TTaskDefinition `xml:"extensionElements>taskDefinition"`
}

// TActivity is an "abstract" struct
type TActivity struct {
	TFlowNode
	IsForCompensation  bool `xml:"isForCompensation,attr"`
	StartQuantity      int  `xml:"startQuantity,attr" default:"1"`
	CompletionQuantity int  `xml:"completionQuantity,attr"`
}

type TTask struct {
	TActivity
}

type TUserTask struct {
	TTask
	Input                []extensions.TIoMapping          `xml:"extensionElements>ioMapping>input"`
	Output               []extensions.TIoMapping          `xml:"extensionElements>ioMapping>output"`
	AssignmentDefinition extensions.TAssignmentDefinition `xml:"extensionElements>assignmentDefinition"`
}

type TParallelGateway struct {
	TGateway
}

type TExclusiveGateway struct {
	TGateway
}

type TIntermediateCatchEvent struct {
	TCatchEvent
	MessageEventDefinition TMessageEventDefinition `xml:"messageEventDefinition"`
	TimerEventDefinition   TTimerEventDefinition   `xml:"timerEventDefinition"`
	LinkEventDefinition    TLinkEventDefinition    `xml:"linkEventDefinition"`
	ParallelMultiple       bool                    `xml:"parallelMultiple"`
	Output                 []extensions.TIoMapping `xml:"extensionElements>ioMapping>output"`
}

type TThrowEvent struct {
	TEvent
}

type TIntermediateThrowEvent struct {
	TThrowEvent
	LinkEventDefinition TLinkEventDefinition    `xml:"linkEventDefinition"`
	Output              []extensions.TIoMapping `xml:"extensionElements>ioMapping>output"`
}

type TEventBasedGateway struct {
	TGateway
}

type TMessageEventDefinition struct {
	TEventDefinition
	MessageRef string `xml:"messageRef,attr"`
}

type TTimerEventDefinition struct {
	TEventDefinition
	TimeDuration TTimeDuration `xml:"timeDuration"`
}

type TTimeDuration struct {
	XMLText string `xml:",innerxml"`
}

type TLinkEventDefinition struct {
	TEventDefinition
	Name string `xml:"name,attr"`
}

type TMessage struct {
	TRootElement
	Name    string `xml:"name,attr"`
	ItemRef string `xml:"itemRef,attr"`
}

type TInclusiveGateway struct {
	TGateway
}
