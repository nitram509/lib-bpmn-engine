package zeebe

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"testing"
)

var numberOfHazelcastSendToRingbufferCalls = 0

func TestPublishNewProcessEvent(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New("name")
	zeebeExporter := createExporterWithHazelcastMock()
	bpmnEngine.AddEventExporter(&zeebeExporter)

	// when
	bpmnEngine.LoadFromFile("../../../../test-cases/simple_task.bpmn")

	then.AssertThat(t, numberOfHazelcastSendToRingbufferCalls, is.EqualTo(1))
}

func TestPublishNewProcessInstanceEvent(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New("name")
	zeebeExporter := createExporterWithHazelcastMock()
	bpmnEngine.AddEventExporter(&zeebeExporter)
	process, _ := bpmnEngine.LoadFromFile("../../../../test-cases/simple_task.bpmn")
	numberOfHazelcastSendToRingbufferCalls = 0 // reset

	// when
	bpmnEngine.CreateInstance(process.ProcessKey, nil)

	then.AssertThat(t, numberOfHazelcastSendToRingbufferCalls, is.EqualTo(1))
}

func TestPublishNewElementEvent(t *testing.T) {
	// setup
	bpmnEngine := bpmn_engine.New("name")
	zeebeExporter := createExporterWithHazelcastMock()
	bpmnEngine.AddEventExporter(&zeebeExporter)
	process, _ := bpmnEngine.LoadFromFile("../../../../test-cases/simple_task.bpmn")
	numberOfHazelcastSendToRingbufferCalls = 0 // reset

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	then.AssertThat(t, numberOfHazelcastSendToRingbufferCalls, is.GreaterThan(1))
}

func createExporterWithHazelcastMock() exporter {
	numberOfHazelcastSendToRingbufferCalls = 0
	zeebeExporter := exporter{
		hazelcast: Hazelcast{
			sendToRingbufferFunc: func(data []byte) error {
				numberOfHazelcastSendToRingbufferCalls = numberOfHazelcastSendToRingbufferCalls + 1
				return nil
			},
		},
	}
	return zeebeExporter
}
