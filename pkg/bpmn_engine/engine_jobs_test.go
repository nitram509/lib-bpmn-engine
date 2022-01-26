package bpmn_engine

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/process_instance"
	"testing"
)

func Test_a_job_can_fail_and_keeps_the_instance_in_active_state(t *testing.T) {
	// setup
	bpmnEngine := New("name")
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")
	bpmnEngine.AddTaskHandler("id", handler)

	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	then.AssertThat(t, instance.state, is.EqualTo(process_instance.ACTIVE))
}

func handler(job ActivatedJob) {
	job.Fail("just because I can")
}
