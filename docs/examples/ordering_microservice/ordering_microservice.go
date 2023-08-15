package main

import (
	_ "embed"
	"encoding/json"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine"
	"net/http"
)

var bpmnEngine bpmn_engine.BpmnEngineState
var process *bpmn_engine.ProcessInfo

type random struct {
	Glossary struct {
		Title    string `json:"title"`
		GlossDiv struct {
			Title     string `json:"title"`
			GlossList struct {
				GlossEntry struct {
					ID        string `json:"ID"`
					SortAs    string `json:"SortAs"`
					GlossTerm string `json:"GlossTerm"`
					Acronym   string `json:"Acronym"`
					Abbrev    string `json:"Abbrev"`
					GlossDef  struct {
						Para         string   `json:"para"`
						GlossSeeAlso []string `json:"GlossSeeAlso"`
					} `json:"GlossDef"`
					GlossSee string `json:"GlossSee"`
				} `json:"GlossEntry"`
			} `json:"GlossList"`
		} `json:"GlossDiv"`
	} `json:"glossary"`
}

// main does start a trivial microservice, listening on port 8080
// open your web browser with http://localhost:8080/
func main() {
	r := random{}
	json.Unmarshal([]byte(""), &r)

	initHttpRoutes()
	http.ListenAndServe(":8080", nil)
}
