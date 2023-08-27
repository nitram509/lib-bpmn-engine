package bpmn_engine

import "testing"

func Test_Activity_interfaces_implemented(t *testing.T) {
	var _ Activity = &tActivity{}
}

func Test_GatewayActivity_interfaces_implemented(t *testing.T) {
	var _ GatewayActivity = &tGatewayActivity{}
	var _ Activity = &tGatewayActivity{}
}

func Test_EventBaseGatewayActivity_interfaces_implemented(t *testing.T) {
	var _ EventBasedGatewayActivity = &tEventBasedGatewayActivity{}
	var _ Activity = &tGatewayActivity{}
}
