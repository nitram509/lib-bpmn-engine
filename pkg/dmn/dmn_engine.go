package dmn

import (
	"crypto/md5"
	"encoding/xml"
	"github.com/antonmedv/expr"
	"github.com/pbinitiative/zenbpm/pkg/dmn/model/drd"
	"os"
)

type DmnEngine interface {
	LoadFromFile(filename string) (*DmnDefinition, error)
	EvaluateDecision(dmnDefinition *DmnDefinition, decisionId string, variableContext map[string]interface{}) (*DrdInstanceResult, error)
}

type DmnEngineImpl struct {
}

// New creates a new instance of the BPMN Engine;
func New() DmnEngine {
	return &DmnEngineImpl{}
}

func (engine *DmnEngineImpl) LoadFromFile(filename string) (*DmnDefinition, error) {
	xmlData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return engine.load(xmlData, filename)
}

func (state *DmnEngineImpl) load(xmlData []byte, resourceName string) (*DmnDefinition, error) {
	md5sum := md5.Sum(xmlData)
	var definitions drd.TDefinitions
	err := xml.Unmarshal(xmlData, &definitions)
	if err != nil {
		return nil, err
	}

	dmnInfo := DmnDefinition{
		definitions: definitions,
		checksum:    md5sum,
	}

	return &dmnInfo, nil
}

func (engine *DmnEngineImpl) EvaluateDecision(dmnDefinition *DmnDefinition, decisionId string, variableContext map[string]interface{}) (*DrdInstanceResult, error) {
	foundDecision := findDecision(dmnDefinition, decisionId)
	if foundDecision == nil {
		return nil, nil
	}

	decisionTable := foundDecision.DecisionTable
	tInputInstances := make([]TInputInstance, len(decisionTable.Inputs))

	for i, input := range decisionTable.Inputs {
		value, _ := expr.Eval(input.InputExpression.Text, variableContext)
		tInputInstances[i] = TInputInstance{
			input: input,
			value: value,
		}
	}

	result := DrdInstanceResult{
		decisionResultMapping: make(map[string]DecisionInstanceResult),
	}

	for _, rule := range decisionTable.Rules {
		allEntriesMatch := true
		for i, inputEntry := range rule.InputEntry {
			inputInstance := tInputInstances[i]
			value, _ := expr.Eval(inputEntry.Text, variableContext)
			if value != inputInstance.value {
				allEntriesMatch = false
				break
			}
		}

		if allEntriesMatch {
			outputs := make([]TOutputInstance, len(decisionTable.Outputs))
			for i, output := range decisionTable.Outputs {
				outputs[i] = TOutputInstance{
					input: output,
					value: rule.OutputEntry[i].Text,
				}
			}
			result.decisionResultMapping[decisionId] = DecisionInstanceResult{
				tInputInstances,
				outputs,
			}
		}
	}

	return &result, nil
}
