package packets

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEncodingVarInts(test *testing.T) {
	for i := 0; i < 200; i++ {
		arr := EncodeVarLengthInt(i)
		x, _, _ := DecodeVarLengthInt(arr[:])

		if x != i {
			test.Error("Encoding and decoding are not symmetrical")
		}
	}
}

func TestEncodingFixedHeader(t *testing.T) {
	for _, header := range []ControlHeader{
		{Type: 4, RemainingLength: 40, Flags: 10},
		{Type: 1, RemainingLength: 25, Flags: 11},
		{Type: 12, RemainingLength: 400, Flags: 2},
		{Type: 2, RemainingLength: 5, Flags: 0},
	} {

		arr := EncodeFixedHeader(header)
		result, _, _ := DecodeFixedHeader(arr)

		if !cmp.Equal(*result, header) {
			t.Error("Encode and Decode Fixed Header are not symmetrical")
		}

	}
}
