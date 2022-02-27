package main

import (
	_ "embed"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"net/http"
)

var bpmnEngine bpmn_engine.BpmnEngineState
var process *bpmn_engine.ProcessInfo

// main does start a trivial microservice, listening on port 8080
// open your web browser with http://localhost:8080/
func main() {
	initBpmnEngine()
	initHttpRoutes()
	http.ListenAndServe(":8080", nil)
}
