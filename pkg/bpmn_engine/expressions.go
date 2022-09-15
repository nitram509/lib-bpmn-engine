package bpmn_engine

import (
	"strings"

	"github.com/antonmedv/expr"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/variable_scope"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/extensions"
)

func evaluateExpression(expression string, variableContext map[string]interface{}) (interface{}, error) {
	expression = strings.TrimSpace(expression)
	expression = strings.TrimPrefix(expression, "=")
	return expr.Eval(expression, variableContext)
}

func evaluateVariableMapping(instance *ProcessInstanceInfo, mappings []extensions.TIoMapping, dstScope variable_scope.VarScope) error {
	for _, mapping := range mappings {
		evalResult, err := evaluateExpression(mapping.Source, instance.scope.GetContext())
		if err != nil {
			return err
		}
		dstScope.SetVariable(mapping.Target, evalResult)
	}
	return nil
}