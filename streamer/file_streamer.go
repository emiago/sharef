package streamer

import (
	"io"
	"io/ioutil"
	"os"

	webrtc "github.com/pion/webrtc/v3"
)

type ReadFileStreamer interface {
	SendFrame(t int, f Framer) (n uint64, err error)
	OpenFile(name string) (io.ReadCloser, error)
	ReadDir(name string) ([]os.FileInfo, error)
}

type WriteFileStreamer interface {
	SendFrame(t int, f Framer) (n uint64, err error)
	OpenFile(path string, mode os.FileMode) (io.WriteCloser, error)
	Mkdir(path string, mode os.FileMode) (err error)
}

type ReadFileStreamerWebrtc struct {
	channel *webrtc.DataChannel
}

//Implements FrameStreamer
func (s *ReadFileStreamerWebrtc) SendFrame(t int, f Framer) (n uint64, err error) {
	data, err := MarshalFramer(f, t)
	if err != nil {
		return 0, err
	}

	n = uint64(len(data))
	err = s.channel.Send(data)
	return
}

func (s *ReadFileStreamerWebrtc) OpenFile(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	return file, err
}

func (s *ReadFileStreamerWebrtc) ReadDir(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(path)
}

type WriteFileStreamerWebrtc struct {
	channel *webrtc.DataChannel
}

//Implements FrameStreamer
func (s *WriteFileStreamerWebrtc) SendFrame(t int, f Framer) (n uint64, err error) {
	data, err := MarshalFramer(f, t)
	if err != nil {
		return 0, err
	}

	n = uint64(len(data))
	err = s.channel.Send(data)
	return
}

func (s *WriteFileStreamerWebrtc) OpenFile(path string, mode os.FileMode) (io.WriteCloser, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	return file, err
}

func (s *WriteFileStreamerWebrtc) Mkdir(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}
