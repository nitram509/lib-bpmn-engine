package bpmn_engine

import (
	"time"

	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/activity"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
)

type job struct {
	ElementId          string
	ElementInstanceKey int64
	ProcessInstanceKey int64
	JobKey             int64
	State              activity.LifecycleState
	CreatedAt          time.Time
}

// ActivatedJob is a struct to provide information for registered task handler
type activatedJob struct {
	processInstanceInfo *ProcessInstanceInfo
	completeHandler     func()
	failHandler         func(reason string)

	// the key, a unique identifier for the job
	Key int64
	// the job's process instance key
	ProcessInstanceKey int64
	// the bpmn process ID of the job process definition
	BpmnProcessId string
	// the version of the job process definition
	ProcessDefinitionVersion int32
	// the key of the job process definition
	ProcessDefinitionKey int64
	// the associated task element ID
	ElementId string
	// when the job was created
	CreatedAt time.Time
}

// ActivatedJob represents an abstraction for the activated job
// don't forget to call Fail or Complete when your task worker job is complete or not.
type ActivatedJob interface {
	ProcessInstance

	// Get job unique key
	GetKey() int64

	GetProcessInstanceKey() int64

	// Retrieve id of the job process definition
	GetBpmnProcessId() string

	// Retrieve version of the job process definition
	GetProcessDefinitionVersion() int32

	// Retrieve key of the job process definition
	GetProcessDefinitionKey() int64

	// Get element id of the job
	GetElementId() string

	// Fail does set the state the worker missed completing the job
	// Fail and Complete mutual exclude each other
	Fail(reason string)

	// Complete does set the state the worker successfully completing the job
	// Fail and Complete mutual exclude each other
	Complete()
}

var _ ActivatedJob = &activatedJob{}

func (state *BpmnEngineState) handleServiceTask(process *ProcessInfo, instance *ProcessInstanceInfo, element *BPMN20.BaseElement) bool {
	id := (*element).GetId()
	job := findOrCreateJob(state.jobs, id, instance, state.generateKey)

	if nil != state.handlers && nil != state.handlers[id] {
		job.State = activity.Active
		activatedJob := &activatedJob{
			processInstanceInfo:      instance,
			failHandler:              func(reason string) { job.State = activity.Failed },
			completeHandler:          func() { job.State = activity.Completed },
			Key:                      state.generateKey(),
			ProcessInstanceKey:       instance.instanceKey,
			BpmnProcessId:            process.BpmnProcessId,
			ProcessDefinitionVersion: process.Version,
			ProcessDefinitionKey:     process.ProcessKey,
			ElementId:                job.ElementId,
			CreatedAt:                job.CreatedAt,
		}

		// TODO retries ...
		state.handlers[id](activatedJob)
	}

	return job.State == activity.Completed
}

func findOrCreateJob(jobs []*job, id string, instance *ProcessInstanceInfo, generateKey func() int64) *job {
	for _, job := range jobs {
		if job.ElementId == id {
			return job
		}
	}

	elementInstanceKey := generateKey()
	job := job{
		ElementId:          id,
		ElementInstanceKey: elementInstanceKey,
		ProcessInstanceKey: instance.GetInstanceKey(),
		JobKey:             elementInstanceKey + 1,
		State:              activity.Active,
		CreatedAt:          time.Now(),
	}

	jobs = append(jobs, &job)

	return &job
}

// GetCreatedAt implements ActivatedJob
func (aj *activatedJob) GetCreatedAt() time.Time {
	return aj.CreatedAt
}

// GetInstanceKey implements ActivatedJob
func (aj *activatedJob) GetInstanceKey() int64 {
	return aj.processInstanceInfo.GetInstanceKey()
}

// GetProcessInfo implements ActivatedJob
func (aj *activatedJob) GetProcessInfo() *ProcessInfo {
	return aj.processInstanceInfo.GetProcessInfo()
}

// GetState implements ActivatedJob
func (aj *activatedJob) GetState() process_instance.State {
	return aj.processInstanceInfo.GetState()
}

// GetElementId implements ActivatedJob
func (aj *activatedJob) GetElementId() string {
	return aj.ElementId
}

// GetKey implements ActivatedJob
func (aj *activatedJob) GetKey() int64 {
	return aj.Key
}

// GetBpmnProcessId implements ActivatedJob
func (aj *activatedJob) GetBpmnProcessId() string {
	return aj.BpmnProcessId
}

// GetProcessDefinitionKey implements ActivatedJob
func (aj *activatedJob) GetProcessDefinitionKey() int64 {
	return aj.ProcessDefinitionKey
}

// GetProcessDefinitionVersion implements ActivatedJob
func (aj *activatedJob) GetProcessDefinitionVersion() int32 {
	return aj.ProcessDefinitionVersion
}

// GetProcessInstanceKey implements ActivatedJob
func (aj *activatedJob) GetProcessInstanceKey() int64 {
	return aj.ProcessInstanceKey
}

// GetVariable implements ActivatedJob
func (aj *activatedJob) GetVariable(key string) interface{} {
	return aj.processInstanceInfo.GetVariable(key)
}

// SetVariable implements ActivatedJob
func (aj *activatedJob) SetVariable(key string, value interface{}) {
	aj.processInstanceInfo.SetVariable(key, value)
}

// Fail implements ActivatedJob
func (aj *activatedJob) Fail(reason string) {
	aj.failHandler(reason)
}

// Complete implements ActivatedJob
func (aj *activatedJob) Complete() {
	aj.completeHandler()
}
