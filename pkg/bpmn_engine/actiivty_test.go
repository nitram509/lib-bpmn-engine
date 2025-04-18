package bpmn_engine

import (
	"fmt"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"reflect"
	"testing"
)

func Test_Activity_interfaces_implemented(t *testing.T) {
	var _ activity = &elementActivity{}
}

func Test_GatewayActivity_interfaces_implemented(t *testing.T) {
	var _ activity = &gatewayActivity{}
}

func Test_EventBaseGatewayActivity_interfaces_implemented(t *testing.T) {
	var _ activity = &gatewayActivity{}
}

func Test_Timer_implements_Activity(t *testing.T) {
	var _ activity = &Timer{}
}

func Test_Job_implements_Activity(t *testing.T) {
	var _ activity = &job{}
}

func Test_MessageSubscription_implements_Activity(t *testing.T) {
	var _ activity = &MessageSubscription{}
}

func Test_SetState_is_working(t *testing.T) {
	// to avoid errors, when anyone forgets the pointer on the receiver type
	tests := []struct {
		a activity
	}{
		{&MessageSubscription{}},
		{&Timer{}},
		{&elementActivity{}},
		{&eventBasedGatewayActivity{}},
		{&gatewayActivity{}},
		{&job{}},
		{&processInstanceInfo{}},
		{&subProcessInfo{}},
	}
	for _, test := range tests {

		t.Run(fmt.Sprintf("%s", reflect.TypeOf(test.a)), func(t *testing.T) {
			test.a.SetState(Completed)
			then.AssertThat(t, test.a.State(), is.EqualTo(Completed))
		})
	}
}
