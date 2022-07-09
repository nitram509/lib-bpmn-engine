package zeebe

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-go-client"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"time"
)

// hint:
// protoc --go_opt=paths=source_relative --go_out=. --go_opt=Mschema.proto=exporter/  schema.proto

type Exporter struct {
	position int64
}

func (exporter *Exporter) NewProcess(eventId int64, processId string, processKey int64, version int32, xmlData []byte, resourceName string, checksum string) {

	exporter.updatePosition()

	rcd := ProcessRecord{
		Metadata: &RecordMetadata{
			PartitionId:          1,
			Position:             exporter.position,
			Key:                  eventId,
			Timestamp:            time.Now().UnixMilli(),
			RecordType:           RecordMetadata_EVENT,
			Intent:               "CREATED",
			ValueType:            RecordMetadata_PROCESS,
			SourceRecordPosition: exporter.position,
			RejectionReason:      "NULL_VAL",
		},
		BpmnProcessId:        processId,    //string
		Version:              version,      //int32
		ProcessDefinitionKey: processKey,   //int64
		ResourceName:         resourceName, //string
		Checksum:             checksum,     //string
		Resource:             xmlData,      //[]byte
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

	sendHazelcast(data)
}

func (exporter *Exporter) updatePosition() {
	exporter.position = exporter.position + 1
}

func (exporter *Exporter) NewProcessInstance(eventId int64, processId string, processKey int64, version int32) {
	exporter.updatePosition()

	deploymentRecord := ProcessInstanceRecord{
		Metadata: &RecordMetadata{
			PartitionId:          1,
			Position:             exporter.position,
			Key:                  eventId,
			Timestamp:            time.Now().UnixMilli(),
			RecordType:           RecordMetadata_EVENT,
			Intent:               "ELEMENT_ACTIVATED",
			ValueType:            RecordMetadata_PROCESS_INSTANCE,
			SourceRecordPosition: exporter.position,
		},
		BpmnProcessId:            processId,
		Version:                  1,  //int32
		ProcessDefinitionKey:     1,  //int64
		ProcessInstanceKey:       1,  //int64
		ElementId:                "", //string
		FlowScopeKey:             1,  //int64
		BpmnElementType:          "", //string
		ParentProcessInstanceKey: 0,  //not supported for now
		ParentElementInstanceKey: 0,  //not supported for now

	}

	data, err := proto.Marshal(&deploymentRecord)
	if err != nil {
		panic(fmt.Errorf("cannot marshal proto message to binary: %w", err))
	}
	println(data)

	dRecord, err := anypb.New(&deploymentRecord)
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

	sendHazelcast(data)

}

func sendHazelcast(data []byte) {
	ctx := context.Background()
	// Start the client with defaults.
	client, err := hazelcast.StartNewClient(ctx)
	if err != nil {
		panic(err)
	}
	// Get a reference to the queue.
	rb, err := client.GetRingbuffer(ctx, "zeebe")
	if err != nil {
		panic(err)
	}

	// Add an item to the queue if space is available (non-blocking).
	_, err = rb.Add(ctx, data, hazelcast.OverflowPolicyOverwrite)
	if err != nil {
		panic(err)
	}
}
