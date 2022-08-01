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
		position: calculateStartPosition(),
		hazelcast: Hazelcast{
			sendToRingbufferFunc: func(data []byte) error {
				return sendHazelcast(ringbuffer, data)
			},
		},
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

func sendHazelcast(rb *hazelcast.Ringbuffer, data []byte) error {
	_, err := rb.Add(context.Background(), data, hazelcast.OverflowPolicyOverwrite)
	return err
}

func (e *exporter) NewProcessEvent(event *bpmnEngineExporter.ProcessEvent) {

	e.updatePosition()

	rcd := ProcessRecord{
		Metadata: &RecordMetadata{
			PartitionId:          1,
			Position:             e.position,
			Key:                  event.ProcessKey,
			Timestamp:            time.Now().UnixMilli(),
			RecordType:           RecordMetadata_EVENT,
			Intent:               string(bpmnEngineExporter.Created),
			ValueType:            RecordMetadata_PROCESS,
			SourceRecordPosition: e.position,
			RejectionReason:      "NULL_VAL",
		},
		BpmnProcessId:        event.ProcessId,
		Version:              event.Version,
		ProcessDefinitionKey: event.ProcessKey,
		ResourceName:         event.ResourceName,
		Checksum:             []byte(event.Checksum),
		Resource:             event.XmlData,
	}

	e.sendAsRecord(&rcd)
}

func (e *exporter) EndProcessEvent(event *bpmnEngineExporter.ProcessInstanceEvent) {
	e.updatePosition()

	processInstanceRecord := ProcessInstanceRecord{
		Metadata: &RecordMetadata{
			PartitionId:          1,
			Position:             e.position,
			Key:                  event.ProcessInstanceKey,
			Timestamp:            time.Now().UnixMilli(),
			RecordType:           RecordMetadata_EVENT,
			Intent:               string(bpmnEngineExporter.ElementCompleted),
			ValueType:            RecordMetadata_PROCESS_INSTANCE,
			SourceRecordPosition: e.position,
			RejectionReason:      "NULL_VAL",
		},
		BpmnProcessId:            event.ProcessId,
		Version:                  event.Version,
		ProcessDefinitionKey:     event.ProcessKey,
		ProcessInstanceKey:       event.ProcessInstanceKey,
		ElementId:                event.ProcessId,
		FlowScopeKey:             noInstanceKey,
		BpmnElementType:          "PROCESS",
		ParentProcessInstanceKey: noInstanceKey,
		ParentElementInstanceKey: noInstanceKey,
	}

	e.sendAsRecord(&processInstanceRecord)
}

func (e *exporter) NewProcessInstanceEvent(event *bpmnEngineExporter.ProcessInstanceEvent) {
	e.updatePosition()

	processInstanceRecord := ProcessInstanceRecord{
		Metadata: &RecordMetadata{
			PartitionId:          1,
			Position:             e.position,
			Key:                  event.ProcessInstanceKey,
			Timestamp:            time.Now().UnixMilli(),
			RecordType:           RecordMetadata_EVENT,
			Intent:               string(bpmnEngineExporter.ElementActivated),
			ValueType:            RecordMetadata_PROCESS_INSTANCE,
			SourceRecordPosition: e.position,
			RejectionReason:      "NULL_VAL",
		},
		BpmnProcessId:            event.ProcessId,
		Version:                  event.Version,
		ProcessDefinitionKey:     event.ProcessKey,
		ProcessInstanceKey:       event.ProcessInstanceKey,
		ElementId:                event.ProcessId,
		FlowScopeKey:             noInstanceKey,
		BpmnElementType:          "PROCESS",
		ParentProcessInstanceKey: noInstanceKey,
		ParentElementInstanceKey: noInstanceKey,
	}

	e.sendAsRecord(&processInstanceRecord)
}

func (e *exporter) NewElementEvent(event *bpmnEngineExporter.ProcessInstanceEvent, elementInfo *bpmnEngineExporter.ElementInfo) {
	e.updatePosition()

	processInstanceRecord := ProcessInstanceRecord{
		Metadata: &RecordMetadata{
			PartitionId:          1,
			Position:             e.position,
			Key:                  event.ProcessInstanceKey,
			Timestamp:            time.Now().UnixMilli(),
			RecordType:           RecordMetadata_EVENT,
			Intent:               elementInfo.Intent,
			ValueType:            RecordMetadata_PROCESS_INSTANCE,
			SourceRecordPosition: e.position,
			RejectionReason:      "NULL_VAL",
		},
		BpmnProcessId:            event.ProcessId,
		Version:                  event.Version,
		ProcessDefinitionKey:     event.ProcessKey,
		ProcessInstanceKey:       event.ProcessInstanceKey,
		ElementId:                elementInfo.ElementId,
		FlowScopeKey:             event.ProcessInstanceKey,
		BpmnElementType:          elementInfo.BpmnElementType,
		ParentProcessInstanceKey: noInstanceKey,
		ParentElementInstanceKey: noInstanceKey,
	}

	e.sendAsRecord(&processInstanceRecord)
}

func (e *exporter) sendAsRecord(msg proto.Message) error {
	serializedMessage, err := anypb.New(msg)
	if err != nil {
		panic(fmt.Errorf("cannot marshal 'msg' proto message to binary: %w", err))
	}

	record := Record{
		Record: serializedMessage,
	}

	serializedRecord, err := proto.Marshal(&record)
	if err != nil {
		panic(fmt.Errorf("cannot marshal 'record' proto message to binary: %w", err))
	}

	return e.hazelcast.SendToRingbuffer(serializedRecord)
}

// convenient updates of position, so we can track if we lost a message.
func (e *exporter) updatePosition() {
	e.position = e.position + 1
}

// we need to have a start position, because Zeebe Simple Monitor will filter duplicate events,
// by identical record IDs. A record ID is composed of 'partitionId' and 'position'.
// By using a timestamp in millis, we have a useful base figure = for debugging purpose.
// By shifting 8 bits, we could potentially fire 255 events, within a millisecond.
// The goal is to reduce the chance of collisions, when one will use the same Hazelcast ringbuffer
// and Zeebe Simple Monitor instance and does restart the application using this Zeebe exporter.
func calculateStartPosition() int64 {
	return time.Now().UnixMilli() << 8
}
