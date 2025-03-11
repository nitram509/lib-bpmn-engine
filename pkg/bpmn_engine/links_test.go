package bpmn_engine

import (
	"testing"

	"github.com/corbym/gocrest/has"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

func Test_Link_events_are_thrown_and_caught_and_flow_continued(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-link-events.bpmn")
	bpmnEngine.NewTaskHandler().Type("task").Handler(cp.TaskHandler)
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance.State, is.EqualTo(Completed))
	then.AssertThat(t, cp.CallPath, is.EqualTo("Task-A,Task-B"))
}

func Test_missing_intermediate_link_catch_event_stops_engine_with_error(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-link-event-broken.bpmn")
	bpmnEngine.NewTaskHandler().Type("task").Handler(cp.TaskHandler)
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, err, is.Not(is.Nil()))
	then.AssertThat(t, err.Error(), has.Prefix("missing link intermediate catch event with linkName="))
	then.AssertThat(t, instance.State, is.EqualTo(Failed))
	then.AssertThat(t, cp.CallPath, is.EqualTo(""))
}

func Test_missing_intermediate_link_variables_mapped(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-link-event-output-variables.bpmn")
	bpmnEngine.NewTaskHandler().Type("task").Handler(cp.TaskHandler)
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance.State, is.EqualTo(Completed))
	then.AssertThat(t, cp.CallPath, is.EqualTo("Task"))

	// then
	then.AssertThat(t, instance.GetVariable("throw"), is.Not(is.Nil()))
	then.AssertThat(t, instance.GetVariable("throw").(string), is.EqualTo("throw"))
	// then
	then.AssertThat(t, instance.GetVariable("catch"), is.Not(is.Nil()))
	then.AssertThat(t, instance.GetVariable("catch").(string), is.EqualTo("catch"))
}
