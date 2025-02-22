package bpmn_engine

import (
	"fmt"
	"github.com/corbym/gocrest/is"
	"testing"

	"github.com/corbym/gocrest/then"
)

func Test_MarshallEngine(t *testing.T) {
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task-with_output_mapping.bpmn")
	bpmnEngine.NewTaskHandler().Id("id").Handler(func(job ActivatedJob) {
		job.SetVariable("valueFromHandler", true)
		job.SetVariable("otherVariable", "value")
		job.Complete()
	})
	variableContext := make(map[string]interface{})
	variableContext["hello"] = "world"
	variableContext["john"] = "doe"

	_, _ = bpmnEngine.CreateAndRunInstance(process.ProcessKey, variableContext)

	data := bpmnEngine.Marshal()

	fmt.Println(string(data))

	bpmnEngine, _ = Unmarshal(data)
	vars := bpmnEngine.ProcessInstances()[0].VariableHolder
	then.AssertThat(t, vars.GetVariable("hello"), is.EqualTo("world"))
	then.AssertThat(t, vars.GetVariable("john"), is.EqualTo("doe"))
	then.AssertThat(t, vars.GetVariable("valueFromHandler"), is.EqualTo(true))
}
