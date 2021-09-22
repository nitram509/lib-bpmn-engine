package engine

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
