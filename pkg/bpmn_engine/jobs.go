package bpmn_engine

import (
	"time"

	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
)

type job struct {
	ElementId          string        `json:"id"`
	ElementInstanceKey int64         `json:"ik"`
	ProcessInstanceKey int64         `json:"pik"`
	JobKey             int64         `json:"jk"`
	JobState           ActivityState `json:"s"`
	CreatedAt          time.Time     `json:"c"`
	baseElement        *BPMN20.BaseElement
	// Failure returned by a handler with job.Fail(string)
	Failure string `json:"f,omitempty"`
	// ErrorCode event thrown by a handler with job.ThrowError(string)
	ErrorCode string `json:"ec,omitempty"`
}

func (j job) Key() int64 {
	return j.JobKey
}

func (j job) State() ActivityState {
	return j.JobState
}

func (j *job) SetState(state ActivityState) {
	j.JobState = state
}

func (j job) Element() *BPMN20.BaseElement {
	return j.baseElement
}

func findOrCreateJob(jobs *[]*job, element *BPMN20.TaskElement, instance *processInstanceInfo, generateKey func() int64) *job {
	be := (*element).(BPMN20.BaseElement)
	for _, job := range *jobs {
		if job.ElementId == be.GetId() {
			return job
		}
	}

	elementInstanceKey := generateKey()
	job := job{
		ElementId:          be.GetId(),
		ElementInstanceKey: elementInstanceKey,
		ProcessInstanceKey: instance.GetInstanceKey(),
		JobKey:             elementInstanceKey + 1,
		JobState:           Active,
		CreatedAt:          time.Now(),
		baseElement:        &be,
	}

	*jobs = append(*jobs, &job)

	return &job
}
