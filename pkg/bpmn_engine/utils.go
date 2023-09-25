package bpmn_engine

func areEqual(a [16]byte, b [16]byte) bool {
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func contains(strings []string, s string) bool {
	for _, aString := range strings {
		if aString == s {
			return true
		}
	}
	return false
}
