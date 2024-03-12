package bpmn_engine

import (
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

func TestMultipleTaskHanler(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/exclusive-gateway-with-condition.bpmn")
	bpmnEngine.NewTaskHandler().Id("task-a").Handler(cp.CallPathHandler)
	bpmnEngine.NewTaskHandler().Weight(10).Id("task-b").Handler(func(job ActivatedJob) {
		if len(cp.CallPath) > 0 {
			cp.CallPath += ","
		}
		cp.CallPath += job.GetElementId()
		cp.CallPath += "10"
		job.Complete()
	})
	bpmnEngine.NewTaskHandler().Id("task-b").Handler(cp.CallPathHandler)
	variables := map[string]interface{}{
		"price": -50,
	}

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, variables)

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("task-b10"))
}
