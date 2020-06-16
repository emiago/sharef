package streamer

import (
	"encoding/json"
	"fmt"
)

const (
	FRAME_NEWSTREAM = iota
	FRAME_DATA
	FRAME_OK
	FRAME_ERROR
)

type Framer interface {
	T(t int)
	GetT() int
}

type Frame struct {
	Type int
}

func (f *Frame) T(t int) {
	f.Type = t
}

func (f *Frame) GetT() int {
	return f.Type
}

type FrameError struct {
	Frame
	Err error
}

type FrameNewStream struct {
	Frame
	Info StreamFile
}

type FrameData struct {
	Frame
	Data []byte
}

func ParseFrameData(data []byte) (Framer, error) {
	var frame Framer
	frame = &Frame{}
	if err := json.Unmarshal(data, &frame); err != nil {
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

	err := json.Unmarshal(data, &frame)
	return frame, err
}
