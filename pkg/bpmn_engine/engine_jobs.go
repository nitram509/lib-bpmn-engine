package bpmn_engine

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/activity"
	"time"
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
// don't forget to call Fail or Complete when your task worker job is complete or not.
type ActivatedJob struct {
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

func (state *BpmnEngineState) handleServiceTask(process *ProcessInfo, instance *ProcessInstanceInfo, element *BPMN20.BaseElement) bool {
	id := (*element).GetId()
	job := findOrCreateJob(state.jobs, id, instance, state.generateKey)
	if nil != state.handlers && nil != state.handlers[id] {
		job.State = activity.Active
		activatedJob := ActivatedJob{
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

// GetVariable from the process instance's variable context
func (activatedJob ActivatedJob) GetVariable(key string) interface{} {
	return activatedJob.processInstanceInfo.GetVariable(key)
}

// SetVariable to the process instance's variable context
func (activatedJob ActivatedJob) SetVariable(key string, value interface{}) {
	activatedJob.processInstanceInfo.SetVariable(key, value)
}

// Fail does set the state the worker missed completing the job
// Fail and Complete mutual exclude each other
func (activatedJob ActivatedJob) Fail(reason string) {
	activatedJob.failHandler(reason)
}

// Complete does set the state the worker successfully completing the job
// Fail and Complete mutual exclude each other
func (activatedJob ActivatedJob) Complete() {
	activatedJob.completeHandler()
}
