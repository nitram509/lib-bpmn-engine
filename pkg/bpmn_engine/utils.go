package bpmn_engine

func areEqual(a [16]byte, b [16]byte) bool {
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
