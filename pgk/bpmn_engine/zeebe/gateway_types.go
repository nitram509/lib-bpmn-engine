package zeebe

// see https://github.com/camunda-cloud/zeebe/blob/0.13.1/gateway-protocol/src/main/proto/gateway.proto
type DeployWorkflowResponse struct {
	key       string
	processes []WorkflowMetadata
}

// see https://github.com/camunda-cloud/zeebe/blob/0.13.1/gateway-protocol/src/main/proto/gateway.proto
type WorkflowMetadata struct {
	BpmnProcessId string
	Version       int32
	ProcessKey    int64
	ResourceName  string
	// todo this should be private and not exposed
	Md5sum [16]byte
}
