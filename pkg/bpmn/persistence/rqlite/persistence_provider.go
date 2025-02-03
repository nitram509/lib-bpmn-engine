package rqlite

import (
	bpmnEngineExporter "github.com/pbinitiative/zenbpm/pkg/bpmn/exporter"
	sql "github.com/pbinitiative/zenbpm/pkg/bpmn/persistence/rqlite/sql"
)

type BpmnEnginePersistence interface {
	FindProcesses(processId string, processKey int64) []*sql.ProcessDefinitionEntity
	FindProcessInstances(processInstanceKey int64, processDefinitionKey int64) []*sql.ProcessInstanceEntity
	FindMessageSubscription(originActivityKey int64, processInstanceKey int64, elementId string, state []string) []*sql.MessageSubscriptionEntity
	FindTimers(originActivityKey int64, processInstanceKey int64, state []string) []*sql.TimerEntity
	FindJobs(elementId string, processInstanceKey int64, jobKey int64, state []string) []*sql.JobEntity
	FindActivitiesByProcessInstanceKey(processInstanceKey int64) []*sql.ActivityInstanceEntity

	IsLeader() bool
	GetLeaderAddress() string

	PersistNewProcess(process *sql.ProcessDefinitionEntity) error
	PersistProcessInstance(processInstance *sql.ProcessInstanceEntity) error
	PersistNewMessageSubscription(subscription *sql.MessageSubscriptionEntity) error
	PersistNewTimer(timer *sql.TimerEntity) error
	PersistJob(job *sql.JobEntity) error
	PersistActivity(event *bpmnEngineExporter.ProcessInstanceEvent, elementInfo *bpmnEngineExporter.ElementInfo) error
}
