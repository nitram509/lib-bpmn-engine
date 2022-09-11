package BPMN20

// tests to get quick compiler warnings, when interface is not correctly implemented

var _ TaskElement = &TServiceTask{}
var _ TaskElement = &TUserTask{}

var _ BaseElement = &TStartEvent{}
var _ BaseElement = &TEndEvent{}
var _ BaseElement = &TServiceTask{}
var _ BaseElement = &TUserTask{}
var _ BaseElement = &TParallelGateway{}
var _ BaseElement = &TExclusiveGateway{}
var _ BaseElement = &TIntermediateCatchEvent{}
var _ BaseElement = &TEventBasedGateway{}
