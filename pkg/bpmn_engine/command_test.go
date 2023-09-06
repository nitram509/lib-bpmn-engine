package bpmn_engine

import "testing"

func Test_ActivityCommand_interfaces_implemented(t *testing.T) {
	var _ command = &activityCommand{}
}

func Test_FlowTransitionCommand_interfaces_implemented(t *testing.T) {
	var _ command = &flowTransitionCommand{}
}

func Test_ContinueActivityCommand_interfaces_implemented(t *testing.T) {
	var _ command = &continueActivityCommand{}
}

func Test_ErrorCommand_interfaces_implemented(t *testing.T) {
	var _ command = &errorCommand{}
}
func Test_CheckExclusiveGatewayDoneCommand_interfaces_implemented(t *testing.T) {
	var _ command = &checkExclusiveGatewayDoneCommand{}
}
