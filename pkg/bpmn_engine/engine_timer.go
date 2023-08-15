package bpmn_engine

import (
	"errors"
	"fmt"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"github.com/senseyeio/duration"
	"time"
)

// Timer is created, when a process instance reaches a Timer Intermediate Message Event.
// The logic is simple: CreatedAt + Duration = DueAt
// The State is one of [ TimerCreated, TimerTriggered, TimerCancelled ]
type Timer struct {
	ElementId          string
	ElementInstanceKey int64
	ProcessKey         int64
	ProcessInstanceKey int64
	State              TimerState
	CreatedAt          time.Time
	DueAt              time.Time
	Duration           time.Duration
}

type TimerState string

const TimerCreated TimerState = "CREATED"
const TimerTriggered TimerState = "TRIGGERED"
const TimerCancelled TimerState = "CANCELLED"

func (state *BpmnEngineState) handleIntermediateTimerCatchEvent(process *ProcessInfo, instance *processInstanceInfo, ice BPMN20.TIntermediateCatchEvent) bool {
	timer := findExistingTimerNotYetTriggered(state, ice.Id, instance)
	if timer == nil {
		newTimer, err := createNewTimer(process, instance, ice, state.generateKey)
		if err != nil {
			// TODO: proper error handling
			return false
		}
		timer = newTimer
		state.timers = append(state.timers, timer)
	}
	if time.Now().After(timer.DueAt) {
		timer.State = TimerTriggered
		return true
	}
	return false
}

func createNewTimer(process *ProcessInfo, instance *processInstanceInfo, ice BPMN20.TIntermediateCatchEvent,
	generateKey func() int64) (*Timer, error) {
	durationVal, err := findDurationValue(ice, process)
	if err != nil {
		return nil, &BpmnEngineError{Msg: fmt.Sprintf("Error parsing 'timeDuration' value "+
			"from element with ID=%s. Error:%s", ice.Id, err.Error())}
	}
	now := time.Now()
	return &Timer{
		ElementId:          ice.Id,
		ElementInstanceKey: generateKey(),
		ProcessKey:         process.ProcessKey,
		ProcessInstanceKey: instance.InstanceKey,
		State:              TimerCreated,
		CreatedAt:          now,
		DueAt:              durationVal.Shift(now),
		Duration:           time.Duration(durationVal.TS) * time.Second,
	}, nil
}

func findExistingTimerNotYetTriggered(state *BpmnEngineState, id string, instance *processInstanceInfo) *Timer {
	var t *Timer
	for _, timer := range state.timers {
		if timer.ElementId == id && timer.ProcessInstanceKey == instance.GetInstanceKey() && timer.State == TimerCreated {
			t = timer
			break
		}
	}
	return t
}

func findDurationValue(ice BPMN20.TIntermediateCatchEvent, process *ProcessInfo) (duration.Duration, error) {
	durationStr := &ice.TimerEventDefinition.TimeDuration.XMLText
	if durationStr == nil {
		return duration.Duration{}, errors.New(fmt.Sprintf("Can't find 'timeDuration' value for INTERMEDIATE_CATCH_EVENT with id=%s", ice.Id))
	}
	d, err := duration.ParseISO8601(*durationStr)
	return d, err
}

func checkDueTimersAndFindIntermediateCatchEvent(timers []*Timer, intermediateCatchEvents []BPMN20.TIntermediateCatchEvent, instance *processInstanceInfo) *BPMN20.BaseElement {
	for _, timer := range timers {
		if timer.ProcessInstanceKey == instance.GetInstanceKey() && timer.State == TimerCreated {
			if time.Now().After(timer.DueAt) {
				for _, ice := range intermediateCatchEvents {
					if ice.Id == timer.ElementId {
						be := BPMN20.BaseElement(ice)
						return &be
					}
				}
			}
		}
	}
	return nil
}
