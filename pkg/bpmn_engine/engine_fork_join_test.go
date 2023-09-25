package bpmn_engine

import (
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

func TestForkUncontrolledJoin(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/fork-uncontrolled-join.bpmn")
	bpmnEngine.NewTaskHandler().Id("id-a-1").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("id-a-2").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("id-b-1").Handler(cp.TaskHandler)

	// when
	_, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("id-a-1,id-a-2,id-b-1,id-b-1"))
}

func TestForkControlledParallelJoin(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/fork-controlled-parallel-join.bpmn")
	bpmnEngine.NewTaskHandler().Id("id-a-1").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("id-a-2").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("id-b-1").Handler(cp.TaskHandler)

	// when
	_, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("id-a-1,id-a-2,id-b-1"))
}

func TestForkControlledExclusiveJoin(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/fork-controlled-exclusive-join.bpmn")
	bpmnEngine.NewTaskHandler().Id("id-a-1").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("id-a-2").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("id-b-1").Handler(cp.TaskHandler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("id-a-1,id-a-2,id-b-1,id-b-1"))
}
