package dmn

import (
	"fmt"
	"testing"
)

func TestRegisterHandlerByTaskIdGetsCalled(t *testing.T) {
	// setup
	dmnEngine := New()
	dmnDefinition, _ := dmnEngine.LoadFromFile("./test/test-cases/sport.dmn")

	variableContext := make(map[string]interface{})
	variableContext["equipment"] = "racket"
	variableContext["location"] = "outdoor"

	result, _ := dmnEngine.EvaluateDecision(dmnDefinition, "sport", variableContext)

	fmt.Println("Result:", result)
}
