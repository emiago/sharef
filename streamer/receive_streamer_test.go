package streamer

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	webrtc "github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MyWriteCloser struct {
	buf *bytes.Buffer
	// buf []byte
}

func (mwc *MyWriteCloser) Write(data []byte) (int, error) {
	return mwc.buf.Write(data)
	// log.Printf("Write %p", mwc.buf)
	// n := copy(mwc.buf, data)
	// return n, nil
}

func (mwc *MyWriteCloser) Close() error {
	return nil
}

type MockWriteFileStreamer struct {
	sentFrames []Framer
	writedata  *bytes.Buffer
	readN      uint64
	openfiles  map[string][]os.FileInfo
}

func (s *MockWriteFileStreamer) SendFrame(t int, f Framer) (n uint64, err error) {
	s.sentFrames = append(s.sentFrames, f)
	return s.readN, nil
}

func (s *MockWriteFileStreamer) OpenFile(path string, mode os.FileMode) (io.WriteCloser, error) {
	fi := &FileStat{
		name:    path,
		size:    0,
		mode:    mode,
		modtime: time.Now(),
	}
	s.openfiles[path] = append(s.openfiles[path], fi)

	// s.writedata = make([]byte, 0, 1024)
	// buf := bytes.NewBuffer(s.writedata)
	// buf.Write()
	file := &MyWriteCloser{s.writedata}
	return file, nil
}

func (s *MockWriteFileStreamer) Mkdir(path string, mode os.FileMode) error {
	fi := &FileStat{
		name:    path,
		size:    0,
		mode:    mode,
		modtime: time.Now(),
	}
	s.openfiles[path] = append(s.openfiles[path], fi)
	return nil
}

func NewMockReceiveStreamer(name string) (*ReceiveStreamer, error) {
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, err
	}

	dataChannel, err := conn.CreateDataChannel(name, DataChannelInitFileStream())
	sender := NewReceiveStreamer(dataChannel, "")
	return sender, nil
}

func checkNewStreamFrameSingle(t *testing.T, sf StreamFile, receiver *ReceiveStreamer, mocker *MockWriteFileStreamer) {
	err := receiver.handleNewStreamFrame(sf)
	require.Nil(t, err)

	opened, ok := mocker.openfiles[receiver.streamInfo.FullPath]
	require.True(t, ok)
	require.Equal(t, 1, len(opened))

	file := opened[0]
	assert.Equal(t, file.Name(), sf.Name)
}

func TestReceiveStreamerNewStreamFrame(t *testing.T) {
	receiver, err := NewMockReceiveStreamer("test")
	require.Nil(t, err)

	mocker := &MockWriteFileStreamer{
		openfiles: make(map[string][]os.FileInfo),
	}
	receiver.WriteFileStreamer = mocker

	t.Run("File", func(t *testing.T) {
		checkNewStreamFrameSingle(t, StreamFile{Name: "file.txt", Size: int64(512), Mode: 0644}, receiver, mocker)
	})

	t.Run("FileUnderDir", func(t *testing.T) {
		checkNewStreamFrameSingle(t, StreamFile{Name: "subdir/file.txt", Size: int64(512), Mode: 0644}, receiver, mocker)
	})
}

func TestReceiveStreamerStreamData(t *testing.T) {
	receiver, err := NewMockReceiveStreamer("test")
	require.Nil(t, err)

	mocker := &MockWriteFileStreamer{
		openfiles: make(map[string][]os.FileInfo),
		writedata: bytes.NewBuffer([]byte{}),
	}
	receiver.WriteFileStreamer = mocker

	content := []byte("Here some content of file\n Giving some breaks \nNew lines")
	sf := StreamFile{Name: "file.txt", Size: int64(len(content)), Mode: 0644}

	err = receiver.handleNewStreamFrame(sf)
	require.Nil(t, err)
	require.NotNil(t, receiver.stream)

	receiver.streamFrameData(content)

	assert.Equal(t, mocker.writedata.Len(), len(content))
	assert.Equal(t, mocker.writedata.Bytes(), content)
}
