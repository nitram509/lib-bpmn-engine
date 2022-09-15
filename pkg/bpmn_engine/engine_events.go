package bpmn_engine

import (
	"fmt"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/activity"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
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

type catchEvent struct {
	name       string
	caughtAt   time.Time
	isConsumed bool
	variables  map[string]interface{}
}

// PublishEventForInstance publishes a message with a given name and also adds variables to the process instance, which fetches this event
func (state *BpmnEngineState) PublishEventForInstance(processInstanceKey int64, messageName string, variables map[string]interface{}) error {
	processInstance := state.findProcessInstance(processInstanceKey)
	if processInstance != nil {
		event := catchEvent{
			caughtAt:   time.Now(),
			name:       messageName,
			variables:  variables,
			isConsumed: false,
		}
		processInstance.caughtEvents = append(processInstance.caughtEvents, event)
	} else {
		return fmt.Errorf("no process instance with key=%d found", processInstanceKey)
	}
	return nil
}

// GetMessageSubscriptions the list of message subscriptions
// hint: each intermediate message catch event, will create such an active subscription,
// when a processes instance reaches such an element.
func (state *BpmnEngineState) GetMessageSubscriptions() []MessageSubscription {
	subscriptions := make([]MessageSubscription, len(state.messageSubscriptions))
	for i, ms := range state.messageSubscriptions {
		subscriptions[i] = *ms
	}
	return subscriptions
}

// GetTimersScheduled the list of all scheduled timers in the engine
// A Timer is created, when a process instance reaches a Timer Intermediate Catch Event element
// and expresses a timestamp in the future
func (state *BpmnEngineState) GetTimersScheduled() []Timer {
	timers := make([]Timer, len(state.timers))
	for i, t := range state.timers {
		timers[i] = *t
	}
	return timers
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
		caughtEvent.isConsumed = true
		for k, v := range caughtEvent.variables {
			instance.SetVariable(k, v)
		}
		if err := evaluateVariableMapping(instance, ice.Output, instance.scope); err != nil {
			messageSubscription.State = activity.Failed
			instance.state = process_instance.FAILED
			return false
		}
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

// find first matching catchEvent
func findMatchingCaughtEvent(messages *[]BPMN20.TMessage, instance *ProcessInstanceInfo, ice BPMN20.TIntermediateCatchEvent) *catchEvent {
	msgName := findMessageNameById(messages, ice.MessageEventDefinition.MessageRef)
	for i := 0; i < len(instance.caughtEvents); i++ {
		var caughtEvent = &instance.caughtEvents[i]
		if !caughtEvent.isConsumed && msgName == caughtEvent.name {
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

func (state *BpmnEngineState) findProcessInstance(processInstanceKey int64) *ProcessInstanceInfo {
	for _, pi := range state.processInstances {
		if pi.GetInstanceKey() == processInstanceKey {
			return pi
		}
	}
	return nil
}
