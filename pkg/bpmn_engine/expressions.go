package bpmn_engine

import (
	"github.com/antonmedv/expr"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/variable_scope"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"strings"
)

func evaluateExpression(expression string, variableContext map[string]interface{}) (interface{}, error) {
	expression = strings.TrimSpace(expression)
	expression = strings.TrimPrefix(expression, "=")
	return expr.Eval(expression, variableContext)
}

func evaluateVariableMapping(src variable_scope.VarScope, mappings []BPMN20.TIoMapping, dst variable_scope.VarScope) error {
	for _, mapping := range mappings {
		evalResult, err := evaluateExpression(mapping.Source, src.GetContext())
		if err != nil {
			return err
		}
		dst.SetVariable(mapping.Target, evalResult)
	}
	return nil
}