package bpmn

import (
	"fmt"
	"time"

	"github.com/pbinitiative/zenbpm/pkg/bpmn/model/bpmn20"
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
	baseElement        *bpmn20.BaseElement
}

func (m MessageSubscription) Key() int64 {
	return m.ElementInstanceKey
}

func (m MessageSubscription) State() ActivityState {
	return m.MessageState
}

func (m MessageSubscription) Element() *bpmn20.BaseElement {
	return m.baseElement
}

type catchEvent struct {
	Name       string                 `json:"n"`
	CaughtAt   time.Time              `json:"c"`
	IsConsumed bool                   `json:"i"`
	Variables  map[string]interface{} `json:"v"`
}

// PublishEventForInstance publishes a message with a given name and also adds variables to the process instance, which fetches this event
func (state *BpmnEngineState) PublishEventForInstance(processInstanceKey int64, messageName string, variables map[string]interface{}) error {
	processInstance := state.FindProcessInstance(processInstanceKey)
	if processInstance != nil {
		event := catchEvent{
			CaughtAt:   time.Now(),
			Name:       messageName,
			Variables:  variables,
			IsConsumed: false,
		}
		processInstance.CaughtEvents = append(processInstance.CaughtEvents, event)
		state.persistence.PersistProcessInstance(processInstance)
	} else {
		return fmt.Errorf("no process instance with key=%d found", processInstanceKey)
	}
	return nil
}

// GetMessageSubscriptions the list of message subscriptions
// hint: each intermediate message catch event, will create such an active subscription,
// when a processes instance reaches such an element.
func (state *BpmnEngineState) GetMessageSubscriptions() []MessageSubscription {
	messageSubscriptions := state.persistence.FindMessageSubscription(-1, nil, "")
	subscriptions := make([]MessageSubscription, len(messageSubscriptions))
	for i, ms := range messageSubscriptions {
		subscriptions[i] = *ms
	}
	return subscriptions
}

// GetTimersScheduled the list of all scheduled timers in the engine
// A Timer is created, when a process instance reaches a Timer Intermediate Catch Event element
// and expresses a timestamp in the future
func (state *BpmnEngineState) GetTimersScheduled() []Timer {
	timersPersisted := state.persistence.FindTimers(-1, -1)
	timers := make([]Timer, len(timersPersisted))
	for i, t := range timersPersisted {
		timers[i] = *t
	}
	return timers
}

func (state *BpmnEngineState) handleIntermediateMessageCatchEvent(process *ProcessInfo, instance *processInstanceInfo, ice bpmn20.TIntermediateCatchEvent, originActivity activity) (continueFlow bool, ms *MessageSubscription, err error) {
	ms = findMatchingActiveSubscriptions(state, instance, ice.Id)

	if originActivity != nil && (*originActivity.Element()).GetType() == bpmn20.EventBasedGateway {
		ebgActivity := originActivity.(*eventBasedGatewayActivity)
		if ebgActivity.OutboundCompleted() {
			ms.MessageState = WithDrawn // FIXME: is this correct?
			return false, ms, err
		}
	}

	if ms == nil {
		ms = state.createMessageSubscription(instance, ice)
		ms.originActivity = originActivity
		state.persistence.PersistNewMessageSubscription(ms)
	}

	messages := state.findMessagesByProcessKey(process.ProcessKey)
	caughtEvent := findMatchingCaughtEvent(messages, instance, ice)

	if caughtEvent != nil {
		caughtEvent.IsConsumed = true
		for k, v := range caughtEvent.Variables {
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
			if originActivity != nil && (*originActivity.Element()).GetType() == bpmn20.EventBasedGateway {
				ebgActivity := originActivity.(*eventBasedGatewayActivity)
				ebgActivity.SetOutboundCompleted(ice.Id)
			}
		}
		return true, ms, err
	}
	return false, ms, err
}

func (state *BpmnEngineState) createMessageSubscription(instance *processInstanceInfo, ice bpmn20.TIntermediateCatchEvent) *MessageSubscription {
	var be bpmn20.BaseElement = ice
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
	return ms
}

func (state *BpmnEngineState) findMessagesByProcessKey(processKey int64) *[]bpmn20.TMessage {
	p := state.persistence.FindProcessByKey(processKey)
	if p != nil {
		return &p.definitions.Messages
	}
	return nil
}

// find first matching catchEvent
func findMatchingCaughtEvent(messages *[]bpmn20.TMessage, instance *processInstanceInfo, ice bpmn20.TIntermediateCatchEvent) *catchEvent {
	msgName := findMessageNameById(messages, ice.MessageEventDefinition.MessageRef)
	for i := 0; i < len(instance.CaughtEvents); i++ {
		var caughtEvent = &instance.CaughtEvents[i]
		if !caughtEvent.IsConsumed && msgName == caughtEvent.Name {
			return caughtEvent
		}
	}
	return nil
}

func findMessageNameById(messages *[]bpmn20.TMessage, msgId string) string {
	for _, message := range *messages {
		if message.Id == msgId {
			return message.Name
		}
	}
	return ""
}

func findMatchingActiveSubscriptions(state *BpmnEngineState, processInstance *processInstanceInfo, id string) *MessageSubscription {
	messageSubscriptions := state.persistence.FindMessageSubscription(-1, processInstance, id, Active)
	if len(messageSubscriptions) > 0 {
		return messageSubscriptions[0]
	}
	return nil
}
