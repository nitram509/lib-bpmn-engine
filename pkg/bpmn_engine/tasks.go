package bpmn_engine

import "github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"

func (state *BpmnEngineState) handleServiceTask(process *ProcessInfo, instance *processInstanceInfo, element *BPMN20.TaskElement) (bool, *job) {
	job := findOrCreateJob(&state.jobs, element, instance, state.generateKey)

	handler := state.findTaskHandler(element)
	if handler != nil {
		job.JobState = Active
		variableHolder := NewVarHolder(&instance.VariableHolder, nil)
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

func (state *BpmnEngineState) handleUserTask(process *ProcessInfo, instance *processInstanceInfo, element *BPMN20.TaskElement) *job {
	// TODO consider different handlers, since Service Tasks are different in their definition than user tasks
	_, j := state.handleServiceTask(process, instance, element)
	return j
}
