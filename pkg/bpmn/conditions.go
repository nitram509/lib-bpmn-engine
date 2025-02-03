package bpmn

import (
	"fmt"
	"strings"

	"github.com/pbinitiative/zenbpm/pkg/bpmn/model/bpmn20"
)

// exclusivelyFilterByConditionExpression
// [From BPMN 2.0 Specification, chapter 10.5.3 Inclusive Gateway]
// A diverging Exclusive Gateway (Decision) is used to create alternative paths within a Process flow. This is basically
// the “diversion point in the road” for a Process. For a given instance of the Process, only one of the paths can be taken.
// A Decision can be thought of as a question that is asked at a particular point in the Process. The question has a defined
// set of alternative answers. Each answer is associated with a condition Expression that is associated with a Gateway’s
// outgoing Sequence Flows.
// A default path can optionally be identified, to be taken in the event that none of the conditional Expressions evaluate
// to true. If a default path is not specified and the Process is executed such that none of the conditional Expressions
// evaluates to true, a runtime exception occurs.
// A converging Exclusive Gateway is used to merge alternative paths. Each incoming Sequence Flow token is routed
// to the outgoing Sequence Flow without synchronization.
func exclusivelyFilterByConditionExpression(flows []bpmn20.TSequenceFlow, variableContext map[string]interface{}) ([]bpmn20.TSequenceFlow, error) {
	var ret []bpmn20.TSequenceFlow
	flowIds := strings.Builder{}
	for _, flow := range flows {
		if flow.HasConditionExpression() {
			flowIds.WriteString(fmt.Sprintf("[id='%s',name='%s']", flow.Id, flow.Name))
			expression := flow.GetConditionExpression()
			out, err := evaluateExpression(expression, variableContext)
			if err != nil {
				return nil, &ExpressionEvaluationError{
					Msg: fmt.Sprintf("Error evaluating expression in flow element id='%s' name='%s'", flow.Id, flow.Name),
					Err: err,
				}
			}
			if out == true {
				ret = append(ret, flow)
				break
			}
		}
	}
	if len(ret) == 0 {
		ret = append(ret, findDefaultFlow(flows)...)
	}
	if len(ret) == 0 {
		return nil, &ExpressionEvaluationError{
			Msg: fmt.Sprintf("No default flow, nor matching expressions found, for flow elements: %s", flowIds.String()),
			Err: nil,
		}
	}
	return ret, nil
}

// inclusivelyFilterByConditionExpression
// [From BPMN 2.0 Specification, chapter 10.5.3 Inclusive Gateway]
// A diverging Inclusive Gateway (Inclusive Decision) can be used to create alternative but also parallel paths within a
// Process flow. Unlike the Exclusive Gateway, all condition Expressions are evaluated. The true evaluation of one
// condition Expression does not exclude the evaluation of other condition Expressions. All Sequence Flows with
// a true evaluation will be traversed by a token. Since each path is considered to be independent, all combinations of the
// paths MAY be taken, from zero to all.
func inclusivelyFilterByConditionExpression(flows []bpmn20.TSequenceFlow, variableContext map[string]interface{}) ([]bpmn20.TSequenceFlow, error) {
	var ret []bpmn20.TSequenceFlow
	for _, flow := range flows {
		if flow.HasConditionExpression() {
			expression := flow.GetConditionExpression()
			out, err := evaluateExpression(expression, variableContext)
			if err != nil {
				return nil, &ExpressionEvaluationError{
					Msg: fmt.Sprintf("Error evaluating expression in flow element id='%s' name='%s'", flow.Id, flow.Name),
					Err: err,
				}
			}
			if out == true {
				ret = append(ret, flow)
			}
		}
	}
	ret = append(ret, findDefaultFlow(flows)...)
	return ret, nil
}

func findDefaultFlow(flows []bpmn20.TSequenceFlow) (ret []bpmn20.TSequenceFlow) {
	for _, flow := range flows {
		if !flow.HasConditionExpression() {
			ret = append(ret, flow)
			break
		}
	}
	return ret
}
