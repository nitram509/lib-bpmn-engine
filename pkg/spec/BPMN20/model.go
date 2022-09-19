package BPMN20

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
}

type TProcess struct {
	Id                           string                    `xml:"id,attr"`
	Name                         string                    `xml:"name,attr"`
	ProcessType                  string                    `xml:"processType,attr"`
	IsClosed                     bool                      `xml:"isClosed,attr"`
	IsExecutable                 bool                      `xml:"isExecutable,attr"`
	DefinitionalCollaborationRef string                    `xml:"definitionalCollaborationRef,attr"`
	StartEvents                  []TStartEvent             `xml:"startEvent"`
	EndEvents                    []TEndEvent               `xml:"endEvent"`
	SequenceFlows                []TSequenceFlow           `xml:"sequenceFlow"`
	ServiceTasks                 []TServiceTask            `xml:"serviceTask"`
	UserTasks                    []TUserTask               `xml:"userTask"`
	ParallelGateway              []TParallelGateway        `xml:"parallelGateway"`
	ExclusiveGateway             []TExclusiveGateway       `xml:"exclusiveGateway"`
	IntermediateCatchEvent       []TIntermediateCatchEvent `xml:"intermediateCatchEvent"`
	EventBasedGateway            []TEventBasedGateway      `xml:"eventBasedGateway"`
}

type TSequenceFlow struct {
	Id                  string        `xml:"id,attr"`
	Name                string        `xml:"name,attr"`
	SourceRef           string        `xml:"sourceRef,attr"`
	TargetRef           string        `xml:"targetRef,attr"`
	ConditionExpression []TExpression `xml:"conditionExpression"`
}

type TExpression struct {
	Text string `xml:",innerxml"`
}

type TStartEvent struct {
	Id                  string   `xml:"id,attr"`
	Name                string   `xml:"name,attr"`
	IsInterrupting      bool     `xml:"isInterrupting,attr"`
	ParallelMultiple    bool     `xml:"parallelMultiple,attr"`
	IncomingAssociation []string `xml:"incoming"`
	OutgoingAssociation []string `xml:"outgoing"`
}

type TEndEvent struct {
	Id                  string   `xml:"id,attr"`
	Name                string   `xml:"name,attr"`
	IncomingAssociation []string `xml:"incoming"`
	OutgoingAssociation []string `xml:"outgoing"`
}

type TServiceTask struct {
	Id                  string          `xml:"id,attr"`
	Name                string          `xml:"name,attr"`
	Default             string          `xml:"default,attr"`
	TaskDefinition      TTaskDefinition `xml:"extensionElements>taskDefinition"`
	CompletionQuantity  int             `xml:"completionQuantity,attr"`
	IsForCompensation   bool            `xml:"isForCompensation,attr"`
	OperationRef        string          `xml:"operationRef,attr"`
	Implementation      string          `xml:"implementation,attr"`
	IncomingAssociation []string        `xml:"incoming"`
	OutgoingAssociation []string        `xml:"outgoing"`
	Input               []TIoMapping    `xml:"extensionElements>ioMapping>input"`
	Output              []TIoMapping    `xml:"extensionElements>ioMapping>output"`
}

type TUserTask struct {
	Id                  string       `xml:"id,attr"`
	Name                string       `xml:"name,attr"`
	IncomingAssociation []string     `xml:"incoming"`
	OutgoingAssociation []string     `xml:"outgoing"`
	Input               []TIoMapping `xml:"extensionElements>ioMapping>input"`
	Output              []TIoMapping `xml:"extensionElements>ioMapping>output"`
}

type TTaskDefinition struct {
	TypeName string `xml:"type,attr"`
	Retries  string `xml:"retries,attr"`
}

type TIoMapping struct {
	Source string `xml:"source,attr"`
	Target string `xml:"target,attr"`
}

type TParallelGateway struct {
	Id                  string   `xml:"id,attr"`
	Name                string   `xml:"name,attr"`
	IncomingAssociation []string `xml:"incoming"`
	OutgoingAssociation []string `xml:"outgoing"`
}

type TExclusiveGateway struct {
	Id                  string   `xml:"id,attr"`
	Name                string   `xml:"name,attr"`
	IncomingAssociation []string `xml:"incoming"`
	OutgoingAssociation []string `xml:"outgoing"`
}

type TIntermediateCatchEvent struct {
	Id                     string                  `xml:"id,attr"`
	Name                   string                  `xml:"name,attr"`
	IncomingAssociation    []string                `xml:"incoming"`
	OutgoingAssociation    []string                `xml:"outgoing"`
	MessageEventDefinition TMessageEventDefinition `xml:"messageEventDefinition"`
	TimerEventDefinition   TTimerEventDefinition   `xml:"timerEventDefinition"`
	ParallelMultiple       bool                    `xml:"parallelMultiple"`
	Output                 []TIoMapping            `xml:"extensionElements>ioMapping>output"`
}

type TEventBasedGateway struct {
	Id                  string   `xml:"id,attr"`
	Name                string   `xml:"name,attr"`
	IncomingAssociation []string `xml:"incoming"`
	OutgoingAssociation []string `xml:"outgoing"`
}

type TMessageEventDefinition struct {
	Id         string `xml:"id,attr"`
	MessageRef string `xml:"messageRef,attr"`
}

type TTimerEventDefinition struct {
	Id           string        `xml:"id,attr"`
	TimeDuration TTimeDuration `xml:"timeDuration"`
}

type TMessage struct {
	Id   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

type TTimeDuration struct {
	XMLText string `xml:",innerxml"`
}
