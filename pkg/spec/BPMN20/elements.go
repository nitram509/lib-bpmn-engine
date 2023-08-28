package BPMN20

import "github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/extensions"

type ElementType string
type GatewayDirection string

const (
	StartEvent             ElementType = "START_EVENT"
	EndEvent               ElementType = "END_EVENT"
	ServiceTask            ElementType = "SERVICE_TASK"
	UserTask               ElementType = "USER_TASK"
	ParallelGateway        ElementType = "PARALLEL_GATEWAY"
	ExclusiveGateway       ElementType = "EXCLUSIVE_GATEWAY"
	IntermediateCatchEvent ElementType = "INTERMEDIATE_CATCH_EVENT"
	IntermediateThrowEvent ElementType = "INTERMEDIATE_THROW_EVENT"
	EventBasedGateway      ElementType = "EVENT_BASED_GATEWAY"

	SequenceFlow ElementType = "SEQUENCE_FLOW"

	Unspecified GatewayDirection = "Unspecified"
	Converging  GatewayDirection = "Converging"
	Diverging   GatewayDirection = "Diverging"
	Mixed       GatewayDirection = "Mixed"
)

type BaseElement interface {
	GetId() string
	GetName() string
	GetIncomingAssociation() []string
	GetOutgoingAssociation() []string
	GetType() ElementType
}

type TaskElement interface {
	BaseElement
	GetInputMapping() []extensions.TIoMapping
	GetOutputMapping() []extensions.TIoMapping
	GetTaskDefinitionType() string
	GetAssignmentAssignee() string
	GetAssignmentCandidateGroups() []string
}

type GatewayElement interface {
	BaseElement
	IsParallel() bool
	IsExclusive() bool
}

func (startEvent TStartEvent) GetId() string {
	return startEvent.Id
}

func (startEvent TStartEvent) GetName() string {
	return startEvent.Name
}

func (startEvent TStartEvent) GetIncomingAssociation() []string {
	return startEvent.IncomingAssociation
}

func (startEvent TStartEvent) GetOutgoingAssociation() []string {
	return startEvent.OutgoingAssociation
}

func (startEvent TStartEvent) GetType() ElementType {
	return StartEvent
}

func (endEvent TEndEvent) GetId() string {
	return endEvent.Id
}

func (endEvent TEndEvent) GetName() string {
	return endEvent.Name
}

func (endEvent TEndEvent) GetIncomingAssociation() []string {
	return endEvent.IncomingAssociation
}

func (endEvent TEndEvent) GetOutgoingAssociation() []string {
	return endEvent.OutgoingAssociation
}

func (endEvent TEndEvent) GetType() ElementType {
	return EndEvent
}

func (serviceTask TServiceTask) GetId() string {
	return serviceTask.Id
}

func (serviceTask TServiceTask) GetName() string {
	return serviceTask.Name
}

func (serviceTask TServiceTask) GetIncomingAssociation() []string {
	return serviceTask.IncomingAssociation
}

func (serviceTask TServiceTask) GetOutgoingAssociation() []string {
	return serviceTask.OutgoingAssociation
}

func (serviceTask TServiceTask) GetType() ElementType {
	return ServiceTask
}

func (serviceTask TServiceTask) GetInputMapping() []extensions.TIoMapping {
	return serviceTask.Input
}

func (serviceTask TServiceTask) GetOutputMapping() []extensions.TIoMapping {
	return serviceTask.Output
}

func (serviceTask TServiceTask) GetTaskDefinitionType() string {
	return serviceTask.TaskDefinition.TypeName
}

func (serviceTask TServiceTask) GetAssignmentAssignee() string {
	return ""
}

func (serviceTask TServiceTask) GetAssignmentCandidateGroups() []string {
	return []string{}
}

func (userTask TUserTask) GetId() string {
	return userTask.Id
}

func (userTask TUserTask) GetName() string {
	return userTask.Name
}

func (userTask TUserTask) GetIncomingAssociation() []string {
	return userTask.IncomingAssociation
}

func (userTask TUserTask) GetOutgoingAssociation() []string {
	return userTask.OutgoingAssociation
}

func (userTask TUserTask) GetType() ElementType {
	return UserTask
}

func (userTask TUserTask) GetInputMapping() []extensions.TIoMapping {
	return userTask.Input
}

func (userTask TUserTask) GetOutputMapping() []extensions.TIoMapping {
	return userTask.Output
}

func (userTask TUserTask) GetTaskDefinitionType() string {
	return ""
}

func (userTask TUserTask) GetAssignmentAssignee() string {
	return userTask.AssignmentDefinition.Assignee
}

func (userTask TUserTask) GetAssignmentCandidateGroups() []string {
	return userTask.AssignmentDefinition.GetCandidateGroups()
}

func (parallelGateway TParallelGateway) GetId() string {
	return parallelGateway.Id
}

func (parallelGateway TParallelGateway) GetName() string {
	return parallelGateway.Name
}

func (parallelGateway TParallelGateway) GetIncomingAssociation() []string {
	return parallelGateway.IncomingAssociation
}

func (parallelGateway TParallelGateway) GetOutgoingAssociation() []string {
	return parallelGateway.OutgoingAssociation
}

func (parallelGateway TParallelGateway) GetType() ElementType {
	return ParallelGateway
}

func (parallelGateway TParallelGateway) IsParallel() bool {
	return true
}
func (parallelGateway TParallelGateway) IsExclusive() bool {
	return false
}

func (exclusiveGateway TExclusiveGateway) GetId() string {
	return exclusiveGateway.Id
}

func (exclusiveGateway TExclusiveGateway) GetName() string {
	return exclusiveGateway.Name
}

func (exclusiveGateway TExclusiveGateway) GetIncomingAssociation() []string {
	return exclusiveGateway.IncomingAssociation
}

func (exclusiveGateway TExclusiveGateway) GetOutgoingAssociation() []string {
	return exclusiveGateway.OutgoingAssociation
}

func (exclusiveGateway TExclusiveGateway) GetType() ElementType {
	return ExclusiveGateway
}

func (exclusiveGateway TExclusiveGateway) IsParallel() bool {
	return false
}
func (exclusiveGateway TExclusiveGateway) IsExclusive() bool {
	return true
}

func (intermediateCatchEvent TIntermediateCatchEvent) GetId() string {
	return intermediateCatchEvent.Id
}

func (intermediateCatchEvent TIntermediateCatchEvent) GetName() string {
	return intermediateCatchEvent.Name
}

func (intermediateCatchEvent TIntermediateCatchEvent) GetIncomingAssociation() []string {
	return intermediateCatchEvent.IncomingAssociation
}

func (intermediateCatchEvent TIntermediateCatchEvent) GetOutgoingAssociation() []string {
	return intermediateCatchEvent.OutgoingAssociation
}

func (intermediateCatchEvent TIntermediateCatchEvent) GetType() ElementType {
	return IntermediateCatchEvent
}

// -------------------------------------------------------------------------

func (eventBasedGateway TEventBasedGateway) GetId() string {
	return eventBasedGateway.Id
}

func (eventBasedGateway TEventBasedGateway) GetName() string {
	return eventBasedGateway.Name
}

func (eventBasedGateway TEventBasedGateway) GetIncomingAssociation() []string {
	return eventBasedGateway.IncomingAssociation
}

func (eventBasedGateway TEventBasedGateway) GetOutgoingAssociation() []string {
	return eventBasedGateway.OutgoingAssociation
}

func (eventBasedGateway TEventBasedGateway) GetType() ElementType {
	return EventBasedGateway
}

func (eventBasedGateway TEventBasedGateway) IsParallel() bool {
	return false
}

func (eventBasedGateway TEventBasedGateway) IsExclusive() bool {
	return true
}

// -------------------------------------------------------------------------

func (intermediateThrowEvent TIntermediateThrowEvent) GetId() string {
	return intermediateThrowEvent.Id
}

func (intermediateThrowEvent TIntermediateThrowEvent) GetName() string {
	return intermediateThrowEvent.Name
}

func (intermediateThrowEvent TIntermediateThrowEvent) GetIncomingAssociation() []string {
	return intermediateThrowEvent.IncomingAssociation
}

func (intermediateThrowEvent TIntermediateThrowEvent) GetOutgoingAssociation() []string {
	// by specification, not supported
	return nil
}

func (intermediateThrowEvent TIntermediateThrowEvent) GetType() ElementType {
	return IntermediateThrowEvent
}
