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
	elementInstanceKey := generateKey()
	job := Job{
		ElementId:          id,
		ElementInstanceKey: elementInstanceKey,
		ProcessInstanceKey: instance.GetInstanceKey(),
		JobKey:             elementInstanceKey + 1,
		State:              activity.Active,
		CreatedAt:          time.Now(),
	}
	state.jobs = append(state.jobs, &job)

	// TODO: pickup the handler from the Jobs ...

	if nil != state.handlers && nil != state.handlers[id] {
		data := ProcessInstanceContextData{
			taskId:       id,
			processInfo:  process,
			instanceInfo: instance,
		}
		state.handlers[id](&data)
	}
}
