package streamer

import (
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/proto"
)

type FrameEncoding int

const (
	FrameEncodingJSON  FrameEncoding = 0
	FrameEncodingProto FrameEncoding = 1
)

type FrameEncoder interface {
	MarshalFramer(f Framer, t int) ([]byte, error)
	UnmarshalFramer(data []byte) (Framer, error)
}

func NewFrameByType(t int32) Framer {
	switch t {
	case FRAME_DATA:
		return &FrameData{}

	case FRAME_NEWSTREAM:
		return &FrameNewStream{}
	case FRAME_OK:
		return &Frame{}
	case FRAME_ERROR:
		return &FrameError{}
	default:
		return nil
	}
}

type JSONFrameEncoder struct {
}

func (enc *JSONFrameEncoder) MarshalFramer(f Framer, t int) ([]byte, error) {
	f.SetT(int32(t))
	data, err := json.Marshal(f)
	return data, err
}

func (enc *JSONFrameEncoder) UnmarshalFramer(data []byte) (Framer, error) {
	var frame Framer = &Frame{}
	if err := json.Unmarshal(data, frame); err != nil {
		return nil, err
	}

	frame = NewFrameByType(frame.GetT())
	if frame == nil {
		return nil, fmt.Errorf("Unknown type")
	}

	err := json.Unmarshal(data, frame)
	return frame, err
}

type ProtobufFrameEncoder struct {
}

func (enc *ProtobufFrameEncoder) MarshalFramer(f Framer, t int) ([]byte, error) {
	f.SetT(int32(t))
	data, err := proto.Marshal(f)
	return data, err
}

func (enc *ProtobufFrameEncoder) UnmarshalFramer(data []byte) (Framer, error) {
	var frame Framer = &Frame{}
	if err := proto.Unmarshal(data, frame); err != nil {
		return nil, err
	}

	frame = NewFrameByType(frame.GetT())
	if frame == nil {
		return nil, fmt.Errorf("Unknown type")
	}

	// err := json.Unmarshal(data, frame)
	err := proto.Unmarshal(data, frame)
	return frame, err
}
