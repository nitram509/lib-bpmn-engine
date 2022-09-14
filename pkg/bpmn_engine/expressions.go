package bpmn_engine

import (
	"github.com/antonmedv/expr"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"strings"
)

func evaluateExpression(expression string, variableContext map[string]interface{}) (interface{}, error) {
	expression = strings.TrimSpace(expression)
	expression = strings.TrimPrefix(expression, "=")
	return expr.Eval(expression, variableContext)
}

func evaluateVariableMapping(instance *ProcessInstanceInfo, mappings []BPMN20.TIoMapping) error {
	for _, mapping := range mappings {
		evalResult, err := evaluateExpression(mapping.Source, instance.scope.GetContext())
		if err != nil {
			return err
		}
		instance.SetVariable(mapping.Target, evalResult)
	}
	return nil
}