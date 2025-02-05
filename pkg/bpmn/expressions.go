package bpmn

import (
	"strings"

	"github.com/antonmedv/expr"
	"github.com/pbinitiative/zenbpm/pkg/bpmn/model/bpmn20/extensions"
	"github.com/pbinitiative/zenbpm/pkg/bpmn/var_holder"
)

func evaluateExpression(expression string, variableContext map[string]interface{}) (interface{}, error) {
	expression = strings.TrimSpace(expression)
	expression = strings.TrimPrefix(expression, "=")
	return expr.Eval(expression, variableContext)
}

func evaluateLocalVariables(varHolder *var_holder.VariableHolder, mappings []extensions.TIoMapping) error {
	return mapVariables(varHolder, mappings, func(key string, value interface{}) {
		varHolder.SetVariable(key, value)
	})
}

func propagateProcessInstanceVariables(varHolder *var_holder.VariableHolder, mappings []extensions.TIoMapping) error {
	if len(mappings) == 0 {
		for k, v := range varHolder.Variables() {
			varHolder.PropagateVariable(k, v)
		}
	}
	return mapVariables(varHolder, mappings, func(key string, value interface{}) {
		varHolder.PropagateVariable(key, value)
	})
}

func mapVariables(varHolder *var_holder.VariableHolder, mappings []extensions.TIoMapping, setVarFunc func(key string, value interface{})) error {
	for _, mapping := range mappings {
		evalResult, err := evaluateExpression(mapping.Source, varHolder.Variables())
		if err != nil {
			return err
		}
		setVarFunc(mapping.Target, evalResult)
	}
	return nil
}
