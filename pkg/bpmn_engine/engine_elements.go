package bpmn_engine

type elementContext struct {
	activationCounter map[string]int64
}

func (ec *elementContext) getActivationCounter(elementId string) int64 {
	val, ok := ec.activationCounter[elementId]
	if !ok {
		return 0
	}
	return val
}

func (ec *elementContext) setActivationCounter(elementId string, val int64) {
	ec.activationCounter[elementId] = val
}
