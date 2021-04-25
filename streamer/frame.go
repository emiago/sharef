package streamer

import (
	"google.golang.org/protobuf/runtime/protoiface"
)

const (
	FRAME_NEWSTREAM = iota
	FRAME_DATA
	FRAME_OK
	FRAME_ERROR
)

type Framer interface {
	// proto.Message
	protoiface.MessageV1
	//protoreflect.ProtoMessage
	SetT(t int32)
	GetT() int32
}



// type Frame struct {
// 	Type int
// }

func (f *Frame) SetT(t int32) {
	f.T = t
}

func (f *FrameError) SetT(t int32) {
	f.T = t
}

func (f *FrameNewStream) SetT(t int32) {
	f.T = t
}

func (f *FrameData) SetT(t int32) {
	f.T = t
}

// func (f *Frame) GetT() int {
// 	return f.Type
// }

// type FrameError struct {
// 	Frame
// 	Err string
// }

// type FrameNewStream struct {
// 	Frame
// 	Info StreamFile
// }

// type FrameData struct {
// 	Frame
// 	Data []byte
// }
