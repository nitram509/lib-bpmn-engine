package bpmn_engine

import (
	"fmt"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
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

func (state *BpmnEngineState) handleIntermediateMessageCatchEvent(process *ProcessInfo, instance *ProcessInstanceInfo, ice BPMN20.TIntermediateCatchEvent) bool {
	messageSubscription := findMatchingActiveSubscriptions(state.messageSubscriptions, ice.Id)

	if messageSubscription == nil {
		messageSubscription = &MessageSubscription{
			ElementId:          ice.Id,
			ElementInstanceKey: state.generateKey(),
			ProcessInstanceKey: instance.GetInstanceKey(),
			Name:               ice.Name,
			CreatedAt:          time.Now(),
			State:              activity.Active,
		}
		state.messageSubscriptions = append(state.messageSubscriptions, messageSubscription)
	}

	messages := state.findMessagesByProcessKey(process.ProcessKey)
	caughtEvent := findMatchingCaughtEvent(messages, instance, ice)

	if caughtEvent != nil {
		messageSubscription.State = activity.Completed
		caughtEvent.IsConsumed = true
		return continueNextElement
	}
	return !continueNextElement
}

func (state *BpmnEngineState) findMessagesByProcessKey(processKey int64) *[]BPMN20.TMessage {
	for _, p := range state.processes {
		if p.ProcessKey == processKey {
			return &p.definitions.Messages
		}
	}
	return nil
}

// find first matching CatchEvent
func findMatchingCaughtEvent(messages *[]BPMN20.TMessage, instance *ProcessInstanceInfo, ice BPMN20.TIntermediateCatchEvent) *CatchEvent {
	msgName := findMessageNameById(messages, ice.MessageEventDefinition.MessageRef)
	for i := 0; i < len(instance.caughtEvents); i++ {
		var caughtEvent = &instance.caughtEvents[i]
		if !caughtEvent.IsConsumed && msgName == caughtEvent.Name {
			return caughtEvent
		}
	}
	return nil
}

func findMessageNameById(messages *[]BPMN20.TMessage, msgId string) string {
	for _, message := range *messages {
		if message.Id == msgId {
			return message.Name
		}
	}
	return ""
}

func findMatchingActiveSubscriptions(messageSubscriptions []*MessageSubscription, id string) *MessageSubscription {
	var existingSubscription *MessageSubscription
	for _, ms := range messageSubscriptions {
		if ms.State == activity.Active && ms.ElementId == id {
			existingSubscription = ms
			return existingSubscription
		}
	}
	return nil
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
