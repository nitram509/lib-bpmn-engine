package dmn

import "github.com/pbinitiative/zenbpm/pkg/dmn/model/drd"

type DmnDefinition struct {
	definitions drd.TDefinitions // parsed file content
	rawData     string           // the raw source data, compressed and encoded via ascii85
	checksum    [16]byte         // internal checksum to identify different versions
}

type DmnInstance interface {
}

type DmnInstanceImpl struct {
	result DrdInstanceResult
}

type DrdInstanceResult struct {
	decisionResultMapping map[string]DecisionInstanceResult
}

type TInputInstance struct {
	input drd.TInput
	value interface{}
}

type TOutputInstance struct {
	input drd.TOutput
	value interface{}
}

type DecisionInstanceResult struct {
	inputs  []TInputInstance
	outputs []TOutputInstance
}
