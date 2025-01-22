package bpmn_engine

import (
	"time"

	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/var_holder"

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
}

func (j job) Key() int64 {
	return j.JobKey
}

func (j job) State() ActivityState {
	return j.JobState
}

func (j job) Element() *BPMN20.BaseElement {
	return j.baseElement
}

func (state *BpmnEngineState) handleServiceTask(process *ProcessInfo, instance *processInstanceInfo, element *BPMN20.TaskElement) (bool, *job) {
	job := findOrCreateJob(state, element, instance, state.generateKey)

	// handler := state.findTaskHandler(element)
	// if handler != nil {
	variableHolder := var_holder.New(&instance.VariableHolder, nil)
	if job.JobState != Completing {
		job.JobState = Active
		// activatedJob := &activatedJob{
		// 	processInstanceInfo:      instance,
		// 	failHandler:              func(reason string) { job.JobState = Failed },
		// 	completeHandler:          func() { job.JobState = Completed },
		// 	key:                      state.generateKey(),
		// 	processInstanceKey:       instance.InstanceKey,
		// 	bpmnProcessId:            process.BpmnProcessId,
		// 	processDefinitionVersion: process.Version,
		// 	processDefinitionKey:     process.ProcessKey,
		// 	elementId:                job.ElementId,
		// 	createdAt:                job.CreatedAt,
		// 	variableHolder:           variableHolder,
		// }
		if err := evaluateLocalVariables(&variableHolder, (*element).GetInputMapping()); err != nil {
			job.JobState = Failed
			instance.State = Failed
			state.persistence.PersistJob(job)
			return false, job
		}
		// handler(activatedJob)
	}

	if job.JobState == Completing {
		if err := propagateProcessInstanceVariables(&variableHolder, (*element).GetOutputMapping()); err != nil {
			job.JobState = Failed
			instance.State = Failed
		}
		job.JobState = Completed
	}
	state.persistence.PersistJob(job)
	// }

	// FIXME: remove this one
	autoComplete(state, job)

	return job.JobState == Completed, job
}

func autoComplete(state *BpmnEngineState, j *job) {
	j.JobState = Completed
	state.persistence.PersistJob(j)
}

func (state *BpmnEngineState) JobCompleteById(jobId int64) {
	jobs := state.persistence.FindJobs("", nil, jobId)

	if len(jobs) == 0 {
		return
	}
	jobs[0].JobState = Completing
	state.persistence.PersistJob(jobs[0])

	state.RunOrContinueInstance(jobs[0].ProcessInstanceKey)

}

func findOrCreateJob(state *BpmnEngineState, element *BPMN20.TaskElement, instance *processInstanceInfo, generateKey func() int64) *job {
	be := (*element).(BPMN20.BaseElement)
	jobs := state.persistence.FindJobs(be.GetId(), instance, -1)
	if len(jobs) > 0 {
		jobs[0].baseElement = &be
		return jobs[0]
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

	state.persistence.PersistJob(&job)

	return &job
}
