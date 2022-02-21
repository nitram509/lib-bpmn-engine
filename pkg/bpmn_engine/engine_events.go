package bpmn_engine

import (
	"fmt"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/activity"
	"time"
)

type MessageSubscription struct {
	ElementId          string
	ElementInstanceKey int64
	ProcessInstanceKey int64
	Name               string
	State              activity.LifecycleState
	CreatedAt          time.Time
}

type CatchEvent struct {
	Name       string
	CaughtAt   time.Time
	IsConsumed bool
}

func (state *BpmnEngineState) handleIntermediateMessageCatchEvent(id string, name string, instance *ProcessInstanceInfo) bool {
	var caughtEvent *CatchEvent
	// find first matching caught event
	for i, ce := range instance.caughtEvents {
		if ce.IsConsumed && ce.Name != name {
			continue
		}
		caughtEvent = &instance.caughtEvents[i]
	}

	var existingSubscription *MessageSubscription
	for _, ms := range state.messageSubscriptions {
		if ms.ElementId != id && ms.State != activity.Ready {
			continue
		}
		existingSubscription = ms
	}

	if caughtEvent != nil && existingSubscription != nil {
		existingSubscription.State = activity.Completed
		caughtEvent.IsConsumed = true
		// TODO: that's semantically more a "are all pre-conditions met" flag. should be renamed
		return continueNextElement
	} else {
		messageSubscription := MessageSubscription{
			ElementId:          id,
			ElementInstanceKey: state.generateKey(),
			ProcessInstanceKey: instance.GetInstanceKey(),
			Name:               name,
			CreatedAt:          time.Now(),
			State:              activity.Active,
		}
		state.messageSubscriptions = append(state.messageSubscriptions, &messageSubscription)
		if caughtEvent != nil {
			messageSubscription.State = activity.Completed
			caughtEvent.IsConsumed = true
			return continueNextElement
		}
	}
	return false
}

func (state *BpmnEngineState) PublishEventForInstance(processInstanceKey int64, messageName string) error {
	processInstance := state.findProcessInstance(processInstanceKey)
	if processInstance != nil {
		event := CatchEvent{
			CaughtAt:   time.Now(),
			Name:       messageName,
			IsConsumed: false,
		}
		processInstance.caughtEvents = append(processInstance.caughtEvents, event)
	} else {
		return fmt.Errorf("no process instance with key=%d found", processInstanceKey)
	}
	return nil
}

func (state *BpmnEngineState) GetMessageSubscriptions() []MessageSubscription {
	subscriptions := make([]MessageSubscription, len(state.messageSubscriptions))
	for i, ms := range state.messageSubscriptions {
		subscriptions[i] = *ms
	}
	return subscriptions
}

func (state *BpmnEngineState) findProcessInstance(processInstanceKey int64) *ProcessInstanceInfo {
	for _, pi := range state.processInstances {
		if pi.GetInstanceKey() == processInstanceKey {
			return pi
		}
	}
	return nil
}
