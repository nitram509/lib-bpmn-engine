package BPMN20

type ProcessInstanceState int8

const (
	ProcessInstanceReady     ProcessInstanceState = 0
	ProcessInstanceActive    ProcessInstanceState = 1
	ProcessInstanceCompleted ProcessInstanceState = 2
)
