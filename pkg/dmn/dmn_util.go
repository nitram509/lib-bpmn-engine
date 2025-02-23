package dmn

import "github.com/pbinitiative/zenbpm/pkg/dmn/model/drd"

func findDecision(dmnDefinition *DmnDefinition, decisionId string) *drd.TDecision {
	for _, decision := range dmnDefinition.definitions.Decisions {
		if decision.Id == decisionId {
			return &decision
		}
	}
	return nil
}
