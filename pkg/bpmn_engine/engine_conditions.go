package bpmn_engine

import (
	"github.com/antonmedv/expr"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
)

func exclusivelyFilterByConditionExpression(flows []BPMN20.TSequenceFlow, variableContext map[string]interface{}) (ret []BPMN20.TSequenceFlow) {
	for _, flow := range flows {
		if flow.HasConditionExpression() {
			expression := flow.GetConditionExpression()
			out, err := expr.Eval(expression, variableContext)
			if err != nil {
				panic(err.Error())
			}
			if out == true {
				ret = append(ret, flow)
			}
		}
	}
	return ret
}
