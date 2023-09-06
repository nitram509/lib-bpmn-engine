package bpmn_engine

import "testing"

func Test_Activity_interfaces_implemented(t *testing.T) {
	var _ activity = &elementActivity{}
}

func Test_GatewayActivity_interfaces_implemented(t *testing.T) {
	var _ activity = &gatewayActivity{}
}

func Test_EventBaseGatewayActivity_interfaces_implemented(t *testing.T) {
	var _ activity = &gatewayActivity{}
}

func Test_Timer_implements_Activity(t *testing.T) {
	var _ activity = &Timer{}
}

func Test_Job_implements_Activity(t *testing.T) {
	var _ activity = &job{}
}

func Test_MessageSubscription_implements_Activity(t *testing.T) {
	var _ activity = &MessageSubscription{}
}
