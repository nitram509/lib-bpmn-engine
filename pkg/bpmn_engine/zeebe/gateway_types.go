package zeebe

// DON'T USE THIS FOR NOW !
// =======================
//
// VISION:
// I do plan to make the engine emit Zeebe compatible events.
// This could enable larger re-use of tools like Zeebe Simple Monitor web application.

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
}
