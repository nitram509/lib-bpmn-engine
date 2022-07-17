package zeebe

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"testing"
)

func TestPublishDeploymentEvent(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New("name")

	zeebeExporter := createExporterWithHazelcastMock()
	bpmnEngine.AddEventExporter(&zeebeExporter)

	bpmnEngine.LoadFromFile("../../../../test-cases/simple_task.bpmn")

	//then.AssertThat(t, wasCalled, is.True())
}

func createExporterWithHazelcastMock() exporter {
	zeebeExporter := exporter{
		hazelcast: Hazelcast{
			sendToRingbufferFunc: func(data []byte) {
				// no action
			},
		},
	}
	return zeebeExporter
}
