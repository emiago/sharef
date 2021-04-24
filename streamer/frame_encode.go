package streamer

import (
	"fmt"

	"github.com/golang/protobuf/proto"
)

// type FrameEncoder interface {
// 	MarshalFramer(f Framer, t int) ([]byte, error)
// 	UnmarshalFramer(data []byte) (Framer, error)
// }

// type JSONFramer struct {
// }

func MarshalFramer(f Framer, t int) ([]byte, error) {
	f.SetT(int32(t))
	// data, err := json.Marshal(f)

	data, err := proto.Marshal(f)
	return data, err
}

func UnmarshalFramer(data []byte) (Framer, error) {
	var frame Framer = &Frame{}
	// if err := json.Unmarshal(data, frame); err != nil {
	// 	return nil, err
	// }

	if err := proto.Unmarshal(data, frame); err != nil {
		return nil, err
	}

	switch frame.GetT() {
	case FRAME_DATA:
		frame = &FrameData{}

	case FRAME_NEWSTREAM:
		frame = &FrameNewStream{}
	case FRAME_OK:
		frame = &Frame{}
	case FRAME_ERROR:
		frame = &FrameError{}
	default:
		return nil, fmt.Errorf("Unknown type")
	}

	// err := json.Unmarshal(data, frame)
	err := proto.Unmarshal(data, frame)
	return frame, err
}
