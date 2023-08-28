package BPMN20

import "testing"

// tests to get quick compiler warnings, when interface is not correctly implemented

func Test_all_interfaces_implemented(t *testing.T) {
	var _ TaskElement = &TServiceTask{}
	var _ TaskElement = &TUserTask{}

	var _ BaseElement = &TStartEvent{}
	var _ BaseElement = &TEndEvent{}
	var _ BaseElement = &TServiceTask{}
	var _ BaseElement = &TUserTask{}
	var _ BaseElement = &TParallelGateway{}
	var _ BaseElement = &TExclusiveGateway{}
	var _ BaseElement = &TIntermediateCatchEvent{}
	var _ BaseElement = &TIntermediateThrowEvent{}
	var _ BaseElement = &TEventBasedGateway{}
}
