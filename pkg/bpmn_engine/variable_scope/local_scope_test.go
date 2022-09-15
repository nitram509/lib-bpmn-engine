package variable_scope

import (
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

func Test_LocalVariableScope(t *testing.T) {
	// setup
	scope := NewLocalScope(nil)
	scope.SetVariable("name", "bpmn")
	scope.Propagation()

	// want
	then.AssertThat(t, scope.GetVariable("name"), is.EqualTo("bpmn"))
	then.AssertThat(t, scope.GetContext(), is.EqualTo(map[string]interface{}{
		"name": "bpmn",
	}))
}
