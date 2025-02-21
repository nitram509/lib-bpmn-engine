package store

import (
	"context"

	"github.com/rqlite/rqlite/v8/command/proto"
)

type PersistentStorage interface {
	StorageWriter
	StorageReader
	Query(ctx context.Context, req *proto.QueryRequest) ([]*proto.QueryRows, error)
	Execute(ctx context.Context, req *proto.ExecuteRequest) ([]*proto.ExecuteQueryResponse, error)
	IsLeader(ctx context.Context) bool
}

type StorageWriter interface {
	// Examples of how these interfaces should look like based on BpmnEnginePersistence. Structs used in this interface should be part of this package
	// WriteProcess(ctx context.Context, process ProcessDefinition) error
	// WriteProcessInstance(ctx context.Context, processInstance ProcessInstance) error
	// WriteMessageSubscription(ctx context.Context, subscription MessageSubscription) error
	// WriteTimer(ctx context.Context, timer Timer) error
	// WriteJob(ctx context.Context, job Job) error
	// WriteActivity(ctx context.Context, activity Activity) error
}

type StorageReader interface {
	// Examples of how these interfaces should look like based on BpmnEnginePersistence. Structs used in this interface should be part of this package
	// Should IDs be directly in snowflake.ID form or are going to convert each in implementation of this interface?
	// ReadProcessDefinitions(ctx context.Context, processIds ...string) ([]ProcessDefinition, error)
	// ReadProcessInstances(ctx context.Context, processInstanceKeys ...int64) ([]ProcessInstance, error)
	// ReadMessageSubscription(ctx context.Context, originActivityKey int64, processInstanceKey int64, elementId string, state []string) ([]MessageSubscription, error)
	// ReadTimers(ctx context.Context, state TimeState) ([]Timer, error)
	// ReadJobs(ctx context.Context, state JobState) ([]Job, error)
	// ReadActivitiesByProcessInstanceKey(ctx context.Context, processInstanceKey int64) ([]Activity, error)
}
