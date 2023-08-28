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
// The TimerState is one of [ TimerCreated, TimerTriggered, TimerCancelled ]
type Timer struct {
	ElementId          string        `json:"id"`
	ElementInstanceKey int64         `json:"ik"`
	ProcessKey         int64         `json:"pk"`
	ProcessInstanceKey int64         `json:"pik"`
	TimerState         TimerState    `json:"s"`
	CreatedAt          time.Time     `json:"c"`
	DueAt              time.Time     `json:"da"`
	Duration           time.Duration `json:"du"`
	originActivity     Activity
}

type TimerState string

const TimerCreated TimerState = "CREATED"
const TimerTriggered TimerState = "TRIGGERED"
const TimerCancelled TimerState = "CANCELLED"

func (state *BpmnEngineState) handleIntermediateTimerCatchEvent(process *ProcessInfo, instance *processInstanceInfo, ice BPMN20.TIntermediateCatchEvent) (continueFlow bool, timer *Timer, err error) {
	timer = findExistingTimerNotYetTriggered(state, ice.Id, instance)

	if timer != nil && timer.originActivity != nil {
		originActivity := instance.findActivity(timer.originActivity.Key())
		if originActivity != nil && (*originActivity.Element()).GetType() == BPMN20.EventBasedGateway {
			ebgActivity := originActivity.(EventBasedGatewayActivity)
			if ebgActivity.OutboundCompleted() {
				timer.TimerState = TimerCancelled
				return false, timer, err
			}
		}
	}

	if timer == nil {
		newTimer, err := createNewTimer(process, instance, ice, state.generateKey)
		if err != nil {
			evalErr := &ExpressionEvaluationError{
				Msg: fmt.Sprintf("Error evaluating expression in intermediate timer cacht event element id='%s' name='%s'", ice.Id, ice.Name),
				Err: err,
			}
			return false, timer, evalErr
		}
		timer = newTimer
		state.timers = append(state.timers, timer)
	}
	if time.Now().After(timer.DueAt) {
		timer.TimerState = TimerTriggered
		if timer.originActivity != nil {
			originActivity := instance.findActivity(timer.originActivity.Key())
			if originActivity != nil && (*originActivity.Element()).GetType() == BPMN20.EventBasedGateway {
				ebgActivity := originActivity.(EventBasedGatewayActivity)
				ebgActivity.SetOutboundCompleted(ice.Id)
			}
		}
		return true, timer, err
	}
	return false, timer, err
}

func createNewTimer(process *ProcessInfo, instance *processInstanceInfo, ice BPMN20.TIntermediateCatchEvent,
	generateKey func() int64) (*Timer, error) {
	durationVal, err := findDurationValue(ice)
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
		TimerState:         TimerCreated,
		CreatedAt:          now,
		DueAt:              durationVal.Shift(now),
		Duration:           time.Duration(durationVal.TS) * time.Second,
	}, nil
}

func findExistingTimerNotYetTriggered(state *BpmnEngineState, id string, instance *processInstanceInfo) *Timer {
	var t *Timer
	for _, timer := range state.timers {
		if timer.ElementId == id && timer.ProcessInstanceKey == instance.GetInstanceKey() && timer.TimerState == TimerCreated {
			t = timer
			break
		}
	}
	return t
}

func findDurationValue(ice BPMN20.TIntermediateCatchEvent) (duration.Duration, error) {
	durationStr := &ice.TimerEventDefinition.TimeDuration.XMLText
	if durationStr == nil {
		return duration.Duration{}, errors.New(fmt.Sprintf("Can't find 'timeDuration' value for INTERMEDIATE_CATCH_EVENT with id=%s", ice.Id))
	}
	d, err := duration.ParseISO8601(*durationStr)
	return d, err
}
