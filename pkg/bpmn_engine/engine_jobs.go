package bpmn_engine

import (
	"time"

	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/activity"
)

type job struct {
	ElementId          string
	ElementInstanceKey int64
	ProcessInstanceKey int64
	JobKey             int64
	State              activity.LifecycleState
	CreatedAt          time.Time
}

func (state *BpmnEngineState) handleServiceTask(process *ProcessInfo, instance *ProcessInstanceInfo, element *BPMN20.TaskElement) bool {
	id := (*element).GetId()
	job := findOrCreateJob(state.jobs, id, instance, state.generateKey)

	if nil != state.handlers && nil != state.handlers[id] {
		job.State = activity.Active
		activatedJob := &activatedJob{
			processInstanceInfo:      instance,
			failHandler:              func(reason string) { job.State = activity.Failed },
			completeHandler:          func() { job.State = activity.Completed },
			key:                      state.generateKey(),
			processInstanceKey:       instance.instanceKey,
			bpmnProcessId:            process.BpmnProcessId,
			processDefinitionVersion: process.Version,
			processDefinitionKey:     process.ProcessKey,
			elementId:                job.ElementId,
			createdAt:                job.CreatedAt,
		}
		if err := evaluateVariableMapping(instance, (*element).GetInputMapping()); err != nil {
			job.State = activity.Failed
			return false
		}
		state.handlers[id](activatedJob)
		if err := evaluateVariableMapping(instance, (*element).GetOutputMapping()); err != nil {
			job.State = activity.Failed
			return false
		}
	}

	return job.State == activity.Completed
}

func evaluateVariableMapping(instance *ProcessInstanceInfo, mappings []BPMN20.TIoMapping) error {
	for _, mapping := range mappings {
		evalResult, err := evaluateExpression(mapping.Source, instance.variableContext)
		if err != nil {
			return err
		}
		instance.SetVariable(mapping.Target, evalResult)
	}
	return nil
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
