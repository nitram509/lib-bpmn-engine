package bpmn_engine

import (
	"github.com/antonmedv/expr"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"strings"
)

func exclusivelyFilterByConditionExpression(flows []BPMN20.TSequenceFlow, variableContext map[string]interface{}) (ret []BPMN20.TSequenceFlow) {
	for _, flow := range flows {
		if flow.HasConditionExpression() {
			expression := flow.GetConditionExpression()
			out, err := evaluateExpression(expression, variableContext)
			if err != nil {
				panic(err.Error())
			}
			if out == true {
				ret = append(ret, flow)
			}
		}
	}
	if len(ret) == 0 {
		ret = append(ret, findDefaultFlow(flows)...)
	}
	return ret
}

func evaluateExpression(expression string, variableContext map[string]interface{}) (interface{}, error) {
	expression = strings.TrimSpace(expression)
	expression = strings.TrimPrefix(expression, "=")
	return expr.Eval(expression, variableContext)
}

func findDefaultFlow(flows []BPMN20.TSequenceFlow) (ret []BPMN20.TSequenceFlow) {
	for _, flow := range flows {
		if !flow.HasConditionExpression() {
			ret = append(ret, flow)
			break
		}
	}
	return ret
}
