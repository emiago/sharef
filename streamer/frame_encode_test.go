package streamer

import (
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestFrameMarshal(t *testing.T) {
	encoder := &ProtobufFrameEncoder{}

	var compare = func(f Framer, ftype int) {
		marshaled, err := encoder.MarshalFramer(f, ftype)
		require.Nil(t, err)
		unmarshaled, err := encoder.UnmarshalFramer(marshaled)
		require.Nil(t, err)

		assert.DeepEqual(t, f, unmarshaled, cmpopts.IgnoreUnexported(StreamFile{}))
	}

	compare(&Frame{}, FRAME_OK)
	compare(&FrameError{Err: "Some error"}, FRAME_ERROR)
	compare(&FrameNewStream{Info: &StreamFile{Name: "test.txt"}}, FRAME_NEWSTREAM)
	compare(&FrameData{Data: []byte("aaabbbcc")}, FRAME_DATA)
}
