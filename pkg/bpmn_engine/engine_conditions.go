package bpmn_engine

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
)

func exclusivelyFilterByConditionExpression(flows []BPMN20.TSequenceFlow, variableContext map[string]interface{}) ([]BPMN20.TSequenceFlow, error) {
	var ret []BPMN20.TSequenceFlow
	for _, flow := range flows {
		if flow.HasConditionExpression() {
			expression := flow.GetConditionExpression()
			out, err := evaluateExpression(expression, variableContext)
			if err != nil {
				return nil, err
			}
			if out == true {
				ret = append(ret, flow)
			}
		}
	}
	if len(ret) == 0 {
		ret = append(ret, findDefaultFlow(flows)...)
	}
	return ret, nil
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
