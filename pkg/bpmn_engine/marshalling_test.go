package bpmn_engine

import (
	"encoding/json"
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

	data, err := bpmnEngine.Marshal()
	then.AssertThat(t, err, is.Empty())

	fmt.Println(string(data))

	bpmnEngine, _ = Unmarshal(data)
	vars := bpmnEngine.ProcessInstances()[0].VariableHolder
	then.AssertThat(t, vars.GetVariable("hello"), is.EqualTo("world"))
	then.AssertThat(t, vars.GetVariable("john"), is.EqualTo("doe"))
	then.AssertThat(t, vars.GetVariable("valueFromHandler"), is.EqualTo(true))
}

func Test_MarshallEngineWithWrapping(t *testing.T) {
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

	type Wrapper struct {
		Type  string
		Value any
	}

	// Simple wrapper function encapsulate a variable into a string
	wrapVariableFunc := func(_ string, variable any) (any, error) {
		w := Wrapper{
			Type:  fmt.Sprintf("%T", variable),
			Value: variable,
		}

		v, err := json.Marshal(w)
		if err != nil {
			return nil, err
		}

		return string(v), nil
	}

	// Simple unwrapper function unwrap a variable from a string
	unwrapVariableFunc := func(_ string, variable any) (any, error) {
		w := &Wrapper{}
		err := json.Unmarshal([]byte(variable.(string)), w)
		if err != nil {
			return nil, err
		}
		return w.Value, nil
	}

	data, err := bpmnEngine.Marshal(WithMarshalVariableFunc(wrapVariableFunc))
	then.AssertThat(t, err, is.Empty())

	fmt.Println(string(data))

	bpmnEngine, _ = Unmarshal(data, WithUnmarshalVariableFunc(unwrapVariableFunc))
	vars := bpmnEngine.ProcessInstances()[0].VariableHolder
	then.AssertThat(t, vars.GetVariable("hello"), is.EqualTo("world"))
	then.AssertThat(t, vars.GetVariable("john"), is.EqualTo("doe"))
	then.AssertThat(t, vars.GetVariable("valueFromHandler"), is.EqualTo(true))
}
