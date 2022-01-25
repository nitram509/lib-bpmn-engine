package bpmn_engine

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func Test_FindProcessInstanceById_ComfortFunction_ReturnsNilIfNoInstanceFound(t *testing.T) {
	engine := New("name")
	instanceInfo := engine.FindProcessInstanceById(1234)
	then.AssertThat(t, instanceInfo, is.Nil())
}
