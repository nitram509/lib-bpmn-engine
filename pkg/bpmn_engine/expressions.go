package bpmn_engine

import (
	"github.com/pbinitiative/feel"
	"strings"

	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/extensions"
)

func evaluateExpression(expression string, variableContext map[string]interface{}) (interface{}, error) {
	expression = strings.TrimSpace(expression)
	expression = strings.TrimPrefix(expression, "=") // FIXME: this is just for convenience, but should be removed
	res, err := feel.EvalStringWithScope(expression, variableContext)
	if err == nil {
		if num, ok := res.(*feel.Number); ok {
			// TODO: tbc: what about smart conversion to int, in case of integer value?
			return num.Float64(), nil
		}
		if b, ok := res.(bool); ok {
			return b, nil
		}
	}
	return res, err
}

func evaluateLocalVariables(varHolder *VariableHolder, mappings []extensions.TIoMapping) error {
	return mapVariables(varHolder, mappings, func(key string, value interface{}) {
		varHolder.SetVariable(key, value)
	})
}

func propagateProcessInstanceVariables(varHolder *VariableHolder, mappings []extensions.TIoMapping) error {
	if len(mappings) == 0 {
		for k, v := range varHolder.Variables() {
			varHolder.PropagateVariable(k, v)
		}
	}
	return mapVariables(varHolder, mappings, func(key string, value interface{}) {
		varHolder.PropagateVariable(key, value)
	})
}

func mapVariables(varHolder *VariableHolder, mappings []extensions.TIoMapping, setVarFunc func(key string, value interface{})) error {
	for _, mapping := range mappings {
		evalResult, err := evaluateExpression(mapping.Source, varHolder.Variables())
		if err != nil {
			return err
		}
		setVarFunc(mapping.Target, evalResult)
	}
	return nil
}
