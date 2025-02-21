package bpmn

import (
	"os"
	"strings"
	"testing"

	"github.com/corbym/gocrest/has"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/pbinitiative/zenbpm/pkg/bpmn/tests"
)

func Test_FindProcessInstance_ComfortFunction_ReturnsNilIfNoInstanceFound(t *testing.T) {
	bpmnEngine := New(&tests.TestStorage{})
	instanceInfo := bpmnEngine.FindProcessInstance(1234)
	then.AssertThat(t, instanceInfo, is.Nil())

	// cleanup
	bpmnEngine.Stop()
}

func Test_FindProcessesById_ComfortFunction_ReturnsEmptyArrayIfNoInstanceFound(t *testing.T) {
	bpmnEngine := New(&tests.TestStorage{})
	instanceInfo := bpmnEngine.FindProcessesById("unknown-id")
	then.AssertThat(t, instanceInfo, has.Length(0))

	// cleanup
	bpmnEngine.Stop()
}

func Test_FindProcessesById_result_is_ordered_by_version(t *testing.T) {
	bpmnEngine := New(&tests.TestStorage{})

	// setup
	dataV1, err := os.ReadFile("./test-cases/simple_task.bpmn")
	then.AssertThat(t, err, is.Nil())
	_, err = bpmnEngine.LoadFromBytes(dataV1)
	then.AssertThat(t, err, is.Nil())

	// given
	dataV2 := strings.Replace(string(dataV1), "StartEvent_1", "StartEvent_2", -1)
	then.AssertThat(t, dataV2, is.Not(is.EqualTo(string(dataV1))))
	_, err = bpmnEngine.LoadFromBytes([]byte(dataV2))
	then.AssertThat(t, err, is.Nil())

	// when
	infos := bpmnEngine.FindProcessesById("Simple_Task_Process")

	// then
	for i := 0; i < len(infos)-1; i++ {
		then.AssertThat(t, infos[i].Version, is.GreaterThan(infos[i+1].Version))
	}

	// cleanup
	bpmnEngine.Stop()
}
