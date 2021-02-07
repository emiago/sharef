package streamer

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
	Err string
}

type FrameNewStream struct {
	Frame
	Info StreamFile
}

type FrameData struct {
	Frame
	Data []byte
}
