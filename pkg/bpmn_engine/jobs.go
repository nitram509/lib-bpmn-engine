package bpmn_engine

import (
	"time"

	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
)

type job struct {
	ElementId          string               `json:"id"`
	ElementInstanceKey int64                `json:"ik"`
	ProcessInstanceKey int64                `json:"pik"`
	JobKey             int64                `json:"jk"`
	JobState           BPMN20.ActivityState `json:"s"`
	CreatedAt          time.Time            `json:"c"`
	baseElement        *BPMN20.BaseElement
}

func (j job) Key() int64 {
	return j.JobKey
}

func (j job) State() BPMN20.ActivityState {
	return j.JobState
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
		JobState:           BPMN20.Active,
		CreatedAt:          time.Now(),
		baseElement:        &be,
	}

	*jobs = append(*jobs, &job)

	return &job
}
