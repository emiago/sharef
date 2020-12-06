package streamer

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestFrameMarshal(t *testing.T) {
	var compare = func(f Framer, ftype int) {
		marshaled, err := MarshalFramer(f, ftype)
		require.Nil(t, err)
		unmarshaled, err := UnmarshalFramer(marshaled)
		require.Nil(t, err)

		assert.DeepEqual(t, f, unmarshaled)
	}

	compare(&Frame{}, FRAME_OK)
	compare(&FrameError{Err: "Some error"}, FRAME_ERROR)
	compare(&FrameNewStream{Info: StreamFile{Name: "test.txt"}}, FRAME_NEWSTREAM)
	compare(&FrameData{Data: []byte("aaabbbcc")}, FRAME_DATA)
}
