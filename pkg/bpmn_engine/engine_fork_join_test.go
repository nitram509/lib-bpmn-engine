package bpmn_engine

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func TestForkUncontrolledJoin(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/fork-uncontrolled-join.bpmn")
	bpmnEngine.AddTaskHandler("id-a-1", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("id-a-2", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("id-b-1", cp.CallPathHandler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("id-a-1,id-a-2,id-b-1,id-b-1"))
}

func TestForkControlledParallelJoin(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/fork-controlled-parallel-join.bpmn")
	bpmnEngine.AddTaskHandler("id-a-1", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("id-a-2", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("id-b-1", cp.CallPathHandler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("id-a-1,id-a-2,id-b-1"))
}

func TestForkControlledExclusiveJoin(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/fork-controlled-exclusive-join.bpmn")
	bpmnEngine.AddTaskHandler("id-a-1", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("id-a-2", cp.CallPathHandler)
	bpmnEngine.AddTaskHandler("id-b-1", cp.CallPathHandler)

	// when
	bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("id-a-1,id-a-2,id-b-1,id-b-1"))
}
