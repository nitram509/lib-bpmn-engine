package BPMN20

import (
	"encoding/xml"
)

type TDefinitions struct {
	XMLName            xml.Name `xml:"definitions"`
	Id                 string   `xml:"id,attr"`
	Name               string   `xml:"name,attr"`
	TargetNamespace    string   `xml:"targetNamespace,attr"`
	ExpressionLanguage string   `xml:"expressionLanguage,attr"`
	TypeLanguage       string   `xml:"typeLanguage,attr"`
	Exporter           string   `xml:"exporter,attr"`
	ExporterVersion    string   `xml:"exporterVersion,attr"`
	Process            TProcess `xml:"process"`
}

type TProcess struct {
	XMLName                      xml.Name        `xml:"process"`
	Id                           string          `xml:"id,attr"`
	Name                         string          `xml:"name,attr"`
	ProcessType                  string          `xml:"processType,attr"`
	IsClosed                     bool            `xml:"isClosed,attr"`
	IsExecutable                 bool            `xml:"isExecutable,attr"`
	DefinitionalCollaborationRef string          `xml:"definitionalCollaborationRef,attr"`
	StartEvents                  []TStartEvent   `xml:"startEvent"`
	EndEvents                    []TEndEvent     `xml:"endEvent"`
	SequenceFlows                []TSequenceFlow `xml:"sequenceFlow"`
	ServiceTasks                 []TServiceTask  `xml:"serviceTask"`
}

//sequenceFlow id="Flow_0xt1d7q" sourceRef="StartEvent_1" targetRef="Activity_1yyow37"

type TSequenceFlow struct {
	XMLName          xml.Name `xml:"sequenceFlow"`
	Id               string   `xml:"id,attr"`
	Name             string   `xml:"name,attr"`
	IsInterrupting   string   `xml:"sourceRef,attr"`
	ParallelMultiple string   `xml:"targetRef,attr"`
}

type TStartEvent struct {
	XMLName          xml.Name `xml:"startEvent"`
	Id               string   `xml:"id,attr"`
	Name             string   `xml:"name,attr"`
	IsInterrupting   bool     `xml:"isInterrupting,attr"`
	ParallelMultiple bool     `xml:"parallelMultiple,attr"`
}

type TEndEvent struct {
	XMLName xml.Name `xml:"endEvent"`
	Id      string   `xml:"id,attr"`
	Name    string   `xml:"name,attr"`
}

type TServiceTask struct {
	XMLName            xml.Name `xml:"serviceTask"`
	Id                 string   `xml:"id,attr"`
	Name               string   `xml:"name,attr"`
	Default            string   `xml:"default,attr"`
	CompletionQuantity int      `xml:"completionQuantity,attr"`
	IsForCompensation  bool     `xml:"isForCompensation,attr"`
	OperationRef       string   `xml:"operationRef,attr"`
	Implementation     string   `xml:"implementation,attr"`
}
