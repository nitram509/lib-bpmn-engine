package engine

import (
	"github.com/nitram509/golib-bpmn-model/pgk/spec/BPMN/20100501/BPMN20"
)

type Node struct {
	Name string
	Id   string
}

type registeredProcess struct {
	processInfo WorkflowMetadata
}

type BpmnEngine interface {
	LoadFromDefinitions(definitions BPMN20.TDefinitions)
}

type BpmnEngineCore struct {
	registry []registeredProcess
}

func (core BpmnEngineCore) LoadFromDefinitions(definitions BPMN20.TDefinitions) (DeployWorkflowResponse, error) {
	info := WorkflowMetadata{
		bpmnProcessId: "123",
		version:       1,
		processKey:    456,
		resourceName:  "xxx",
	}

	core.registry = append(core.registry, registeredProcess{info})

	result := DeployWorkflowResponse{}
	result.key = "1234567890"
	result.processes = append(result.processes, info)
	return result, nil
}

// see https://github.com/camunda-cloud/zeebe/blob/0.13.1/gateway-protocol/src/main/proto/gateway.proto
type DeployWorkflowResponse struct {
	key       string
	processes []WorkflowMetadata
}

// see https://github.com/camunda-cloud/zeebe/blob/0.13.1/gateway-protocol/src/main/proto/gateway.proto
type WorkflowMetadata struct {
	bpmnProcessId string
	version       int32
	processKey    int64
	resourceName  string
}
