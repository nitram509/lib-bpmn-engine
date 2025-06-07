package BPMN20

import "github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/extensions"

type ElementType string
type GatewayDirection string
type BoundaryType string

const (
	Process                ElementType = "PROCESS"
	SubProcess             ElementType = "SUB_PROCESS"
	StartEvent             ElementType = "START_EVENT"
	EndEvent               ElementType = "END_EVENT"
	ServiceTask            ElementType = "SERVICE_TASK"
	UserTask               ElementType = "USER_TASK"
	ParallelGateway        ElementType = "PARALLEL_GATEWAY"
	ExclusiveGateway       ElementType = "EXCLUSIVE_GATEWAY"
	IntermediateCatchEvent ElementType = "INTERMEDIATE_CATCH_EVENT"
	IntermediateThrowEvent ElementType = "INTERMEDIATE_THROW_EVENT"
	EventBasedGateway      ElementType = "EVENT_BASED_GATEWAY"
	InclusiveGateway       ElementType = "INCLUSIVE_GATEWAY"
	SequenceFlow           ElementType = "SEQUENCE_FLOW"
	BoundaryEvent          ElementType = "BOUNDARY_EVENT"

	Unspecified GatewayDirection = "Unspecified"
	Converging  GatewayDirection = "Converging"
	Diverging   GatewayDirection = "Diverging"
	Mixed       GatewayDirection = "Mixed"

	// Type of boundary event, error, message, timer, etc
	ErrorBoundary   BoundaryType = "Error"
	UnknownBoundary BoundaryType = "Unknown"
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
	IsInclusive() bool
}

type ProcessElement interface {
	BaseElement
	GetStartEvents() []TStartEvent
	GetEndEvents() []TEndEvent
	GetSequenceFlows() []TSequenceFlow
	GetServiceTasks() []TServiceTask
	GetUserTasks() []TUserTask
	GetParallelGateway() []TParallelGateway
	GetExclusiveGateway() []TExclusiveGateway
	GetIntermediateCatchEvent() []TIntermediateCatchEvent
	GetIntermediateTrowEvent() []TIntermediateThrowEvent
	GetEventBasedGateway() []TEventBasedGateway
	GetSubProcess() []TSubProcess
	GetInclusiveGateway() []TInclusiveGateway
	GetBoundaryEvent() []TBoundaryEvent
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

func (parallelGateway TParallelGateway) IsInclusive() bool {
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

func (exclusiveGateway TExclusiveGateway) IsInclusive() bool {
	return false
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

func (eventBasedGateway TEventBasedGateway) IsInclusive() bool {
	return false
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

func (inclusiveGateway TInclusiveGateway) GetId() string {
	return inclusiveGateway.Id
}

func (inclusiveGateway TInclusiveGateway) GetName() string {
	return inclusiveGateway.Name
}

func (inclusiveGateway TInclusiveGateway) GetIncomingAssociation() []string {
	return inclusiveGateway.IncomingAssociation
}

func (inclusiveGateway TInclusiveGateway) GetOutgoingAssociation() []string {
	return inclusiveGateway.OutgoingAssociation
}

func (inclusiveGateway TInclusiveGateway) GetType() ElementType {
	return InclusiveGateway
}

func (inclusiveGateway TInclusiveGateway) IsParallel() bool {
	return false
}

func (inclusiveGateway TInclusiveGateway) IsExclusive() bool {
	return false
}

func (inclusiveGateway TInclusiveGateway) IsInclusive() bool {
	return true
}

// -------------------------------------------------------------------------

func (process TProcess) GetId() string {
	return process.Id
}

func (process TProcess) GetName() string {
	return process.Name
}

func (process TProcess) GetIncomingAssociation() []string {
	return nil
}

func (process TProcess) GetOutgoingAssociation() []string {
	return nil
}

func (process TProcess) GetType() ElementType {
	return Process
}

func (process TProcess) GetStartEvents() []TStartEvent {
	return process.StartEvents
}

func (process TProcess) GetEndEvents() []TEndEvent {
	return process.EndEvents
}

func (process TProcess) GetSequenceFlows() []TSequenceFlow {
	return process.SequenceFlows
}

func (process TProcess) GetServiceTasks() []TServiceTask {
	return process.ServiceTasks
}

func (process TProcess) GetUserTasks() []TUserTask {
	return process.UserTasks
}

func (process TProcess) GetParallelGateway() []TParallelGateway {
	return process.ParallelGateway
}

func (process TProcess) GetExclusiveGateway() []TExclusiveGateway {
	return process.ExclusiveGateway
}

func (process TProcess) GetIntermediateCatchEvent() []TIntermediateCatchEvent {
	return process.IntermediateCatchEvent
}

func (process TProcess) GetIntermediateTrowEvent() []TIntermediateThrowEvent {
	return process.IntermediateTrowEvent
}

func (process TProcess) GetEventBasedGateway() []TEventBasedGateway {
	return process.EventBasedGateway
}

func (process TProcess) GetSubProcess() []TSubProcess {
	return process.SubProcesses
}

func (process TProcess) GetInclusiveGateway() []TInclusiveGateway {
	return process.InclusiveGateway
}

func (process TProcess) GetBoundaryEvent() []TBoundaryEvent {
	return process.BoundaryEvent
}

func (subProcess TSubProcess) GetId() string {
	return subProcess.Id
}

func (subProcess TSubProcess) GetName() string {
	return subProcess.Name
}

func (subProcess TSubProcess) GetIncomingAssociation() []string {
	return subProcess.IncomingAssociation
}

func (subProcess TSubProcess) GetOutgoingAssociation() []string {
	return subProcess.OutgoingAssociation
}

func (subProcess TSubProcess) GetType() ElementType {
	return SubProcess
}

func (subProcess TSubProcess) GetStartEvents() []TStartEvent {
	return subProcess.StartEvents
}

func (subProcess TSubProcess) GetEndEvents() []TEndEvent {
	return subProcess.EndEvents
}

func (subProcess TSubProcess) GetSequenceFlows() []TSequenceFlow {
	return subProcess.SequenceFlows
}

func (subProcess TSubProcess) GetServiceTasks() []TServiceTask {
	return subProcess.ServiceTasks
}

func (subProcess TSubProcess) GetUserTasks() []TUserTask {
	return subProcess.UserTasks
}

func (subProcess TSubProcess) GetParallelGateway() []TParallelGateway {
	return subProcess.ParallelGateway
}

func (subProcess TSubProcess) GetExclusiveGateway() []TExclusiveGateway {
	return subProcess.ExclusiveGateway
}

func (subProcess TSubProcess) GetIntermediateCatchEvent() []TIntermediateCatchEvent {
	return subProcess.IntermediateCatchEvent
}

func (subProcess TSubProcess) GetIntermediateTrowEvent() []TIntermediateThrowEvent {
	return subProcess.IntermediateTrowEvent
}

func (subProcess TSubProcess) GetEventBasedGateway() []TEventBasedGateway {
	return subProcess.EventBasedGateway
}

func (subProcess TSubProcess) GetSubProcess() []TSubProcess {
	return subProcess.SubProcesses
}

func (subProcess TSubProcess) GetInclusiveGateway() []TInclusiveGateway {
	return subProcess.InclusiveGateway
}

func (process TSubProcess) GetBoundaryEvent() []TBoundaryEvent {
	return process.BoundaryEvent
}

func (be TBoundaryEvent) GetId() string {
	return be.Id
}

func (be TBoundaryEvent) GetName() string {
	return be.Name
}

func (be TBoundaryEvent) GetIncomingAssociation() []string {
	return []string{}
}

func (be TBoundaryEvent) GetOutgoingAssociation() []string {
	return be.OutgoingAssociation
}

func (be TBoundaryEvent) GetType() ElementType {
	return BoundaryEvent
}

func (be TBoundaryEvent) GetBoundaryType() BoundaryType {
	if be.ErrorEventDefinition != nil {
		return ErrorBoundary
	}
	return UnknownBoundary
}

func (be TBoundaryEvent) GetOutputMapping() []extensions.TIoMapping {
	return be.Output
}
