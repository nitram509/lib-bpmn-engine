package BPMN20

import (
	"github.com/corbym/gocrest/has"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20/extensions"
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

// tests to get quick compiler warnings, when interface is not correctly implemented

func Test_all_interfaces_implemented(t *testing.T) {
	var _ TaskElement = &TServiceTask{}
	var _ TaskElement = &TUserTask{}

	var _ BaseElement = &TStartEvent{}
	var _ BaseElement = &TEndEvent{}
	var _ BaseElement = &TServiceTask{}
	var _ BaseElement = &TUserTask{}
	var _ BaseElement = &TParallelGateway{}
	var _ BaseElement = &TExclusiveGateway{}
	var _ BaseElement = &TIntermediateCatchEvent{}
	var _ BaseElement = &TIntermediateThrowEvent{}
	var _ BaseElement = &TEventBasedGateway{}
	var _ BaseElement = &TInclusiveGateway{}
	var _ BaseElement = &TBoundaryEvent{}
}

func Test_ErrorBoundaryEvent(t *testing.T) {

	event := TBoundaryEvent{
		TBaseElement:        TBaseElement{Id: "event_1", Documentation: "Documentation"},
		Name:                "Boundary Event",
		AttachedToRef:       "task_1",
		OutgoingAssociation: []string{"flow_1"},
		ErrorEventDefinition: &TErrorEventDefinition{
			Id:       "errorDef_1",
			ErrorRef: "error_1",
		},
		Output: []extensions.TIoMapping{
			{
				Source: "ioSource",
				Target: "ioTarget",
			},
		},
	}
	then.AssertThat(t, event.GetId(), is.EqualTo("event_1"))
	then.AssertThat(t, event.GetName(), is.EqualTo("Boundary Event"))
	then.AssertThat(t, event.GetIncomingAssociation(), has.Length(0))
	then.AssertThat(t, event.GetOutgoingAssociation(), has.Length(1))
	then.AssertThat(t, event.GetType(), is.EqualTo(BoundaryEvent))
	then.AssertThat(t, event.GetBoundaryType(), is.EqualTo(ErrorBoundary))
	then.AssertThat(t, event.GetOutputMapping(), has.Length(1))
}

func Test_UnknownBoundaryEvent(t *testing.T) {

	event := TBoundaryEvent{
		TBaseElement:        TBaseElement{Id: "event_1", Documentation: "Documentation"},
		Name:                "Boundary Event",
		AttachedToRef:       "task_1",
		OutgoingAssociation: []string{"flow_1"},
		Output: []extensions.TIoMapping{
			{
				Source: "ioSource",
				Target: "ioTarget",
			},
		},
	}
	then.AssertThat(t, event.GetId(), is.EqualTo("event_1"))
	then.AssertThat(t, event.GetName(), is.EqualTo("Boundary Event"))
	then.AssertThat(t, event.GetIncomingAssociation(), has.Length(0))
	then.AssertThat(t, event.GetOutgoingAssociation(), has.Length(1))
	then.AssertThat(t, event.GetType(), is.EqualTo(BoundaryEvent))
	then.AssertThat(t, event.GetBoundaryType(), is.EqualTo(UnknownBoundary))
	then.AssertThat(t, event.GetOutputMapping(), has.Length(1))
}
