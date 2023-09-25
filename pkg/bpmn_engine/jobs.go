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
	job := findOrCreateJob(&state.jobs, element, instance, state.generateKey)

	handler := state.findTaskHandler(element)
	if handler != nil {
		job.JobState = Active
		variableHolder := var_holder.New(&instance.VariableHolder, nil)
		activatedJob := &activatedJob{
			processInstanceInfo:      instance,
			failHandler:              func(reason string) { job.JobState = Failed },
			completeHandler:          func() { job.JobState = Completed },
			key:                      state.generateKey(),
			processInstanceKey:       instance.InstanceKey,
			bpmnProcessId:            process.BpmnProcessId,
			processDefinitionVersion: process.Version,
			processDefinitionKey:     process.ProcessKey,
			elementId:                job.ElementId,
			createdAt:                job.CreatedAt,
			variableHolder:           variableHolder,
		}
		if err := evaluateLocalVariables(&variableHolder, (*element).GetInputMapping()); err != nil {
			job.JobState = Failed
			instance.State = Failed
			return false, job
		}
		handler(activatedJob)
		if job.JobState == Completed {
			if err := propagateProcessInstanceVariables(&variableHolder, (*element).GetOutputMapping()); err != nil {
				job.JobState = Failed
				instance.State = Failed
			}
		}
	}

	return job.JobState == Completed, job
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
