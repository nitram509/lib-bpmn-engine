package zeebe

import (
	"context"
	"fmt"
	"time"

	bpmnEngineExporter "github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/exporter"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

type redisExporter struct {
	position int64
	redis    *redis.Client
}

func NewRedisExporter(opt redis.Options) (redisExporter, error) {
	ctx := context.Background()
	// Start the client with defaults.
	client := redis.NewClient(&opt)
	if err := client.Ping(ctx).Err(); err != nil {
		return redisExporter{}, err
	}
	return redisExporter{
		position: calculateStartPosition(),
		redis:    client,
	}, nil
}

func (e *redisExporter) NewProcessEvent(event *bpmnEngineExporter.ProcessEvent) {

	e.updatePosition()

	processInstanceRecord := ProcessRecord{
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

	e.sendAsRecord(&processInstanceRecord, int32(RecordMetadata_PROCESS))
}

func (e *redisExporter) EndProcessEvent(event *bpmnEngineExporter.ProcessInstanceEvent) {
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

	e.sendAsRecord(&processInstanceRecord, int32(RecordMetadata_PROCESS_INSTANCE))
}

func (e *redisExporter) NewProcessInstanceEvent(event *bpmnEngineExporter.ProcessInstanceEvent) {
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

	e.sendAsRecord(&processInstanceRecord, int32(RecordMetadata_PROCESS_INSTANCE))
}
func (e *redisExporter) NewElementEvent(event *bpmnEngineExporter.ProcessInstanceEvent, elementInfo *bpmnEngineExporter.ElementInfo) {
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

	e.sendAsRecord(&processInstanceRecord, int32(RecordMetadata_PROCESS_INSTANCE))
}

// convenient updates of position, so we can track if we lost a message.
func (e *redisExporter) updatePosition() {
	e.position++
}

func (e *redisExporter) sendAsRecord(msg proto.Message, valueType int32) error {
	serializedMessage, err := anypb.New(msg)
	if err != nil {
		return err
	}

	record := Record{
		Record: serializedMessage,
	}

	serializedRecord, err := proto.Marshal(&record)
	if err != nil {
		return err
	}

	return e.redis.XAdd(context.TODO(), &redis.XAddArgs{
		Stream: fmt.Sprintf("zeebe:%s", RecordMetadata_ValueType_name[valueType]),
		Values: map[string]interface{}{"record": serializedRecord},
	}).Err()
}

func (e *redisExporter) UpdateVaribleRecord(event *bpmnEngineExporter.ProcessInstanceEvent, key string, value interface{}) {
	e.updatePosition()

	variableRecord := VariableRecord{
		Metadata: &RecordMetadata{
			PartitionId:          1,
			Position:             e.position,
			Key:                  event.ProcessInstanceKey,
			Timestamp:            time.Now().UnixMilli(),
			RecordType:           RecordMetadata_EVENT,
			Intent:               string(bpmnEngineExporter.ElementActivated),
			ValueType:            RecordMetadata_VARIABLE,
			SourceRecordPosition: e.position,
			RejectionReason:      "NULL_VAL",
		},
		ProcessInstanceKey: event.ProcessInstanceKey,
		ScopeKey:           event.ProcessInstanceKey,
		Name:               key,
		Value:              fmt.Sprint(value),
	}

	e.sendAsRecord(&variableRecord, int32(RecordMetadata_VARIABLE))
}
