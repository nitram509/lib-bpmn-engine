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

func (state *BpmnEngineState) handleServiceTask(id string, process *ProcessInfo, instance *ProcessInstanceInfo) {
	job := findOrCreateJob(state.jobs, id, instance)

	if nil != state.handlers && nil != state.handlers[id] {
		job.State = activity.Active
		data := ProcessInstanceContextData{
			taskId:       id,
			processInfo:  process,
			instanceInfo: instance,
		}
		// TODO: set to failed in case of panic
		state.handlers[id](&data)
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
