package bpmn_engine

import (
	"fmt"
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

const XmlTestString = `<?xml version="1.0" encoding="UTF-8"?><bpmn:process id="Simple_Task_Process" name="aName" isExecutable="true"></bpmn:process></xml>`

func Test_compress_and_encode_produces_ascii_chars(t *testing.T) {
	str := compressAndEncode([]byte(XmlTestString))
	encodedBytes := []byte(str)
	for i := 0; i < len(encodedBytes); i++ {
		b := encodedBytes[i]
		t.Run(fmt.Sprintf("string, index=%d", i), func(t *testing.T) {
			then.AssertThat(t, b,
				is.AllOf(
					is.GreaterThanOrEqualTo(byte(33)),
					is.LessThanOrEqualTo(byte(117)),
				).Reason("every encoded byte shall have an ordinary value of 33 <= x <= 117"))
		})

	}
}

func Test_compress_and_decompress_roundtrip(t *testing.T) {
	encoded := compressAndEncode([]byte(XmlTestString))
	decoded := decodeAndDecompress(encoded)

	then.AssertThat(t, decoded, is.EqualTo([]byte(XmlTestString)))
}
