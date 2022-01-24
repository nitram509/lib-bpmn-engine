package bpmn_engine

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/activity"
	"time"
)

type Job struct {
	ElementId          string
	ElementInstanceKey int64
	ProcessInstanceKey int64
	JobKey             int64
	State              activity.LifecycleState
	CreatedAt          time.Time
}

type ActivatedJob struct {
	processInstanceInfo *ProcessInstanceInfo

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

func (state *BpmnEngineState) handleServiceTask(id string, process *ProcessInfo, instance *ProcessInstanceInfo) {
	job := findOrCreateJob(state.jobs, id, instance)
	if nil != state.handlers && nil != state.handlers[id] {
		job.State = activity.Active
		activatedJob := ActivatedJob{
			processInstanceInfo:      instance,
			Key:                      generateKey(),
			ProcessInstanceKey:       instance.instanceKey,
			BpmnProcessId:            process.BpmnProcessId,
			ProcessDefinitionVersion: process.Version,
			ProcessDefinitionKey:     process.ProcessKey,
			ElementId:                job.ElementId,
			CreatedAt:                job.CreatedAt,
		}
		// TODO: set to failed in case of panic
		state.handlers[id](activatedJob)
		job.State = activity.Completed
	}
}

func findOrCreateJob(jobs []*Job, id string, instance *ProcessInstanceInfo) *Job {
	for _, job := range jobs {
		if job.ElementId == id {
			return job
		}
	}
	elementInstanceKey := generateKey()
	job := Job{
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

func (activatedJob ActivatedJob) GetVariable(name string) string {
	return activatedJob.processInstanceInfo.variableContext[name]
}

func (activatedJob ActivatedJob) SetVariable(name string, value string) {
	activatedJob.processInstanceInfo.variableContext[name] = value
}
