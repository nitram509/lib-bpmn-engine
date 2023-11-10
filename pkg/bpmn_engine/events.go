package bpmn_engine

import (
	"fmt"
	"time"

	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
)

type MessageSubscription struct {
	ElementId          string        `json:"id"`
	ElementInstanceKey int64         `json:"ik"`
	ProcessKey         int64         `json:"pk"`
	ProcessInstanceKey int64         `json:"pik"`
	Name               string        `json:"n"`
	MessageState       ActivityState `json:"s"`
	CreatedAt          time.Time     `json:"c"`
	originActivity     activity
	baseElement        *BPMN20.BaseElement
}

func (m MessageSubscription) Key() int64 {
	return m.ElementInstanceKey
}

func (m MessageSubscription) State() ActivityState {
	return m.MessageState
}

func (m MessageSubscription) Element() *BPMN20.BaseElement {
	return m.baseElement
}

type catchEvent struct {
	name       string
	caughtAt   time.Time
	isConsumed bool
	variables  map[string]interface{}
}

// PublishEventForInstance publishes a message with a given name and also adds variables to the process instance, which fetches this event
func (state *BpmnEngineState) PublishEventForInstance(processInstanceKey int64, messageName string, variables map[string]interface{}) error {
	processInstance := state.FindProcessInstance(processInstanceKey)
	if processInstance != nil {
		event := catchEvent{
			caughtAt:   time.Now(),
			name:       messageName,
			variables:  variables,
			isConsumed: false,
		}
		processInstance.CaughtEvents = append(processInstance.CaughtEvents, event)
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

func (state *BpmnEngineState) handleIntermediateMessageCatchEvent(process *ProcessInfo, instance *processInstanceInfo, ice BPMN20.TIntermediateCatchEvent, originActivity activity) (continueFlow bool, ms *MessageSubscription, err error) {
	ms = findMatchingActiveSubscriptions(state.messageSubscriptions, ice.Id)

	if originActivity != nil && (*originActivity.Element()).GetType() == BPMN20.EventBasedGateway {
		ebgActivity := originActivity.(*eventBasedGatewayActivity)
		if ebgActivity.OutboundCompleted() {
			ms.MessageState = WithDrawn // FIXME: is this correct?
			return false, ms, err
		}
	}

	if ms == nil {
		ms = state.createMessageSubscription(instance, ice)
		ms.originActivity = originActivity
	}

	messages := state.findMessagesByProcessKey(process.ProcessKey)
	caughtEvent := findMatchingCaughtEvent(messages, instance, ice)

	if caughtEvent != nil {
		caughtEvent.isConsumed = true
		for k, v := range caughtEvent.variables {
			instance.SetVariable(k, v)
		}
		if err := evaluateLocalVariables(&instance.VariableHolder, ice.Output); err != nil {
			ms.MessageState = Failed
			instance.State = Failed
			evalErr := &ExpressionEvaluationError{
				Msg: fmt.Sprintf("Error evaluating expression in intermediate message catch event element id='%s' name='%s'", ice.Id, ice.Name),
				Err: err,
			}
			return false, ms, evalErr
		}
		ms.MessageState = Completed
		if ms.originActivity != nil {
			originActivity := instance.findActivity(ms.originActivity.Key())
			if originActivity != nil && (*originActivity.Element()).GetType() == BPMN20.EventBasedGateway {
				ebgActivity := originActivity.(*eventBasedGatewayActivity)
				ebgActivity.SetOutboundCompleted(ice.Id)
			}
		}
		return true, ms, err
	}
	return false, ms, err
}

func (state *BpmnEngineState) createMessageSubscription(instance *processInstanceInfo, ice BPMN20.TIntermediateCatchEvent) *MessageSubscription {
	var be BPMN20.BaseElement = ice
	ms := &MessageSubscription{
		ElementId:          ice.Id,
		ElementInstanceKey: state.generateKey(),
		ProcessKey:         instance.ProcessInfo.ProcessKey,
		ProcessInstanceKey: instance.GetInstanceKey(),
		Name:               ice.Name,
		CreatedAt:          time.Now(),
		MessageState:       Active,
		baseElement:        &be,
	}
	state.messageSubscriptions = append(state.messageSubscriptions, ms)
	return ms
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
func findMatchingCaughtEvent(messages *[]BPMN20.TMessage, instance *processInstanceInfo, ice BPMN20.TIntermediateCatchEvent) *catchEvent {
	msgName := findMessageNameById(messages, ice.MessageEventDefinition.MessageRef)
	for i := 0; i < len(instance.CaughtEvents); i++ {
		var caughtEvent = &instance.CaughtEvents[i]
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
		if ms.MessageState == Active && ms.ElementId == id {
			existingSubscription = ms
			return existingSubscription
		}
	}
	return nil
}
