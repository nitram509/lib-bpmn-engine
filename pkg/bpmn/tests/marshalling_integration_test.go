package tests

import (
	_ "embed"
	"os"
	"path"
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	bpmn_engine "github.com/pbinitiative/zenbpm/pkg/bpmn"
)

func Test_unmarshalled_v1_contains_all_fields(t *testing.T) {
	tests := []struct {
		referenceFile string
	}{
		{"marshal-reference-v1/jobs.json"},
		{"marshal-reference-v1/intermediate-catch-event.json"},
		{"marshal-reference-v1/message-intermediate-timer-event.json"},
		{"marshal-reference-v1/parallel-gateway-flow.json"},
		{"marshal-reference-v1/remain.json"},
	}
	for _, test := range tests {
		testName := path.Base(test.referenceFile)
		t.Run(testName, func(t *testing.T) {
			// setup
			referenceBytes, err := os.ReadFile(test.referenceFile)
			then.AssertThat(t, err, is.Nil())

			// given
			engine, err := bpmn_engine.Unmarshal(referenceBytes)
			then.AssertThat(t, err, is.Nil())

			// when
			marshalledBytes := engine.Marshal()

			equal, err := JSONBytesEqual(referenceBytes, marshalledBytes)
			then.AssertThat(t, err, is.Nil())
			then.AssertThat(t, equal, is.True())
		})
	}
}
