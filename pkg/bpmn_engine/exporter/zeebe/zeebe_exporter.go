package zeebe

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-go-client"
	bpmnEngineExporter "github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/exporter"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"time"
)

// hint:
// protoc --go_opt=paths=source_relative --go_out=. --go_opt=Mschema.proto=exporter/  schema.proto

type exporter struct {
	position  int64
	hazelcast Hazelcast
}

const noInstanceKey = -1

// TODO make hazelcast client configurable
func NewExporter() exporter {
	ringbuffer := createHazelcastRingbuffer()
	return exporter{
		position: 0,
		hazelcast: Hazelcast{
			sendToRingbufferFunc: func(data []byte) {
				sendHazelcast(ringbuffer, data)
			},
		},
	}
}

func (e *exporter) NewProcess(event *bpmnEngineExporter.ProcessEvent) {

	e.updatePosition()

	rcd := ProcessRecord{
		Metadata: &RecordMetadata{
			PartitionId:          1,
			Position:             e.position,
			Key:                  event.ProcessKey,
			Timestamp:            time.Now().UnixMilli(),
			RecordType:           RecordMetadata_EVENT,
			Intent:               "CREATED",
			ValueType:            RecordMetadata_PROCESS,
			SourceRecordPosition: e.position,
			RejectionReason:      "NULL_VAL",
		},
		BpmnProcessId:        event.ProcessId,
		Version:              event.Version,
		ProcessDefinitionKey: event.ProcessKey,
		ResourceName:         event.ResourceName,
		Checksum:             event.Checksum,
		Resource:             event.XmlData,
	}

	data, err := proto.Marshal(&rcd)
	if err != nil {
		panic(fmt.Errorf("cannot marshal proto message to binary: %w", err))
	}

	dRecord, err := anypb.New(&rcd)
	if err != nil {
		panic(fmt.Errorf("cannot marshal proto message to binary: %w", err))
	}

	record := Record{
		Record: dRecord,
	}

	data, err = proto.Marshal(&record)
	if err != nil {
		panic(fmt.Errorf("cannot marshal proto message to binary: %w", err))
	}

	e.hazelcast.SendToRingbuffer(data)
}

func (e *exporter) NewProcessInstance(event *bpmnEngineExporter.ProcessInstanceEvent) {
	e.updatePosition()

	processInstanceRecord := ProcessInstanceRecord{
		Metadata: &RecordMetadata{
			PartitionId:          1,
			Position:             e.position,
			Key:                  event.ProcessInstanceKey,
			Timestamp:            time.Now().UnixMilli(),
			RecordType:           RecordMetadata_EVENT,
			Intent:               "ELEMENT_ACTIVATED",
			ValueType:            RecordMetadata_PROCESS_INSTANCE,
			SourceRecordPosition: e.position,
		},
		BpmnProcessId:            event.ProcessId,
		Version:                  event.Version,
		ProcessDefinitionKey:     event.ProcessKey,
		ProcessInstanceKey:       event.ProcessInstanceKey,
		ElementId:                "",
		FlowScopeKey:             1,
		BpmnElementType:          "",
		ParentProcessInstanceKey: noInstanceKey,
		ParentElementInstanceKey: noInstanceKey,
	}

	data, err := proto.Marshal(&processInstanceRecord)
	if err != nil {
		panic(fmt.Errorf("cannot marshal proto message to binary: %w", err))
	}
	println(data)

	dRecord, err := anypb.New(&processInstanceRecord)
	if err != nil {
		panic(fmt.Errorf("cannot marshal proto message to binary: %w", err))
	}

	record := Record{
		Record: dRecord,
	}

	data, err = proto.Marshal(&record)
	if err != nil {
		panic(fmt.Errorf("cannot marshal proto message to binary: %w", err))
	}

	e.hazelcast.SendToRingbuffer(data)
}

func (e *exporter) updatePosition() {
	e.position = e.position + 1
}

func sendHazelcast(rb *hazelcast.Ringbuffer, data []byte) {
	_, err := rb.Add(context.Background(), data, hazelcast.OverflowPolicyOverwrite)
	if err != nil {
		panic(err) // TODO error handling
	}
}

func createHazelcastRingbuffer() *hazelcast.Ringbuffer {
	ctx := context.Background()
	// Start the client with defaults.
	client, err := hazelcast.StartNewClient(ctx)
	if err != nil {
		panic(err) // TODO error handling
	}
	// Get a reference to the queue.
	rb, err := client.GetRingbuffer(ctx, "zeebe")
	if err != nil {
		panic(err) // TODO error handling
	}
	return rb
}
