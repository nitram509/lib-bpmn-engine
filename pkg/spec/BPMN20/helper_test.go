package BPMN20

import (
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

func Test_no_expression_when_only_blanks(t *testing.T) {
	// given
	flow := TSequenceFlow{
		ConditionExpression: []TExpression{
			{Text: "   "},
		},
	}
	// when
	result := flow.HasConditionExpression()
	// then
	then.AssertThat(t, result, is.False())
}

func Test_has_expression_when_some_characters_present(t *testing.T) {
	// given
	flow := TSequenceFlow{
		ConditionExpression: []TExpression{
			{Text: " x>y "},
		},
	}
	// when
	result := flow.HasConditionExpression()
	// then
	then.AssertThat(t, result, is.True())
}
