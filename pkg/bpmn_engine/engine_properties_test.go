package bpmn_engine

import (
	"os"
	"strings"
	"testing"

	"github.com/corbym/gocrest/has"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

func Test_FindProcessInstance_ComfortFunction_ReturnsNilIfNoInstanceFound(t *testing.T) {
	engine := New()
	instanceInfo := engine.FindProcessInstance(1234)
	then.AssertThat(t, instanceInfo, is.Nil())
}

func Test_FindProcessesById_ComfortFunction_ReturnsNilIfNoInstanceFound(t *testing.T) {
	engine := New()
	instanceInfo := engine.FindProcessesById("unknown-id")
	then.AssertThat(t, instanceInfo, is.Nil())
}

func Test_FindProcessesById_result_is_ordered_by_version(t *testing.T) {
	engine := New()

	// setup
	dataV1, err := os.ReadFile("../../test-cases/simple_task.bpmn")
	then.AssertThat(t, err, is.Nil())
	_, err = engine.LoadFromBytes(dataV1)
	then.AssertThat(t, err, is.Nil())

	// given
	dataV2 := strings.Replace(string(dataV1), "StartEvent_1", "StartEvent_2", -1)
	then.AssertThat(t, dataV2, is.Not(is.EqualTo(string(dataV1))))
	_, err = engine.LoadFromBytes([]byte(dataV2))
	then.AssertThat(t, err, is.Nil())

	// when
	infos := engine.FindProcessesById("Simple_Task_Process")

	// then
	then.AssertThat(t, infos, has.Length(2))
	then.AssertThat(t, infos[0].Version, is.LessThanOrEqualTo(infos[1].Version))
}
