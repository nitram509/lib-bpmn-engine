package bpmn_engine

import "testing"

func Test_ActivityCommand_interfaces_implemented(t *testing.T) {
	var _ activityCommand = &tActivityCommand{}
}

func Test_ContinueActivityCommand_interfaces_implemented(t *testing.T) {
	var _ continueActivityCommand = &tContinueActivityCommand{}
	var _ activityCommand = &tContinueActivityCommand{}
}

func Test_FlowTransitionCommand_interfaces_implemented(t *testing.T) {
	var _ flowTransitionCommand = &tFlowTransitionCommand{}
}
