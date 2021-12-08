package bpmn_engine

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"testing"
)

func Test_IntermediateCatchEvent_sets_process_instance_ready(t *testing.T) {
	// setup
	bpmnEngine := New("name")

	// given
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/message-intermediate-catch-event.bpmn")

	// when
	pi, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	// then
	then.AssertThat(t, pi.GetState(), is.EqualTo(BPMN20.ProcessInstanceReady))
}
