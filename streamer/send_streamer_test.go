package streamer

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

type MockReadFileStreamer struct {
	sentFrames []Framer
	fakedata   []byte
	readN      uint64
	fakeDir    map[string][]os.FileInfo
}

func (s *MockReadFileStreamer) SendFrame(t int, f Framer) (n uint64, err error) {
	s.sentFrames = append(s.sentFrames, f)
	return s.readN, nil
}

func (s *MockReadFileStreamer) OpenFile(path string) (io.ReadCloser, error) {
	buf := bytes.NewReader(s.fakedata)
	file := ioutil.NopCloser(buf)
	return file, nil
}

func (s *MockReadFileStreamer) ReadDir(path string) ([]os.FileInfo, error) {
	return s.fakeDir[path], nil
}

func NewMockSendStreamer(name string, rootpath string) (*SendStreamer, error) {
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, err
	}

	dataChannel, err := conn.CreateDataChannel(name, DataChannelInitFileStream())
	sender := NewSendStreamer(dataChannel, rootpath)
	return sender, nil
}

func TestNewMockStream(t *testing.T) {
	_, err := NewMockSendStreamer("test", "")
	require.Nil(t, err)
}

func TestSendStreamerPrepareNewStream(t *testing.T) {
	// streamInfo := &FileStat{name: "mydir"}
	sender, err := NewMockSendStreamer("test", "/opt/my/some/mydir")
	require.Nil(t, err)

	fi := &FileStat{
		name: "file.txt",
	}

	t.Run("Simple", func(t *testing.T) {
		info := sender.prepareNewStream(fi, "/opt/my/some/mydir/subdir/sub/file.txt")
		assert.Equal(t, info.Name, "mydir/subdir/sub/file.txt")
	})

	t.Run("BadFormat", func(t *testing.T) {
		//Some badly formated
		info := sender.prepareNewStream(fi, "./opt/my/some//mydir/subdir/sub//file.txt")
		assert.Equal(t, info.Name, "mydir/subdir/sub/file.txt")
	})
}

func TestSendStreamerStreamfile(t *testing.T) {
	fi := &FileStat{
		name: "file.txt",
	}

	sender, err := NewMockSendStreamer("test", "/opt/file.txt")
	require.Nil(t, err)

	mocker := &MockReadFileStreamer{
		fakedata: []byte("0123456789"),
		readN:    10,
	}

	sender.ReadFileStreamer = mocker

	file, _ := sender.OpenFile("file.txt")
	err = sender.streamReader(file, fi.Size(), fi.Name())
	require.Nil(t, err)

	sentFrames := mocker.sentFrames
	require.Equal(t, int(1), len(sentFrames))

	//FirstFrame should be new frame data
	frame, ok := sentFrames[0].(*FrameData)
	require.True(t, ok)
	assert.DeepEqual(t, frame.Data, mocker.fakedata)
}

func TestSendStreamerProcessFile(t *testing.T) {
	rootname := "mydir"
	rootpath := "/opt/mydir"

	fi := &FileStat{
		name: rootname,
		mode: os.ModeDir,
	}

	sender, err := NewMockSendStreamer("test", rootpath)
	require.Nil(t, err)

	mocker := &MockReadFileStreamer{
		fakedata: []byte("0123456789"), //This is for file.txt
		readN:    10,
		fakeDir: map[string][]os.FileInfo{
			rootpath: {
				&FileStat{name: "subdir1", size: 0, mode: os.ModeDir, modtime: time.Now()},
				&FileStat{name: "subdir2", size: 0, mode: os.ModeDir, modtime: time.Now()},
			},
			rootpath + "/subdir1": {
				&FileStat{name: "file.txt", size: 10, mode: 0, modtime: time.Now()},
			},
		},
	}

	sender.ReadFileStreamer = mocker
	go func() {
		//Fake response from receiver, and check is every processed
		for {
			sender.frameCh <- &Frame{Type: FRAME_OK}
		}
	}()

	err = sender.processFile(fi, sender.streamPath)
	require.Nil(t, err)

	sentFrames := mocker.sentFrames

	var checkFrame func(subfi os.FileInfo, rname string, rpath string)
	checkFrame = func(subfi os.FileInfo, rname string, rpath string) {
		if len(sentFrames) == 0 {
			t.Fatalf("Missing frames")
		}
		frame := sentFrames[0]
		newframe, ok := frame.(*FrameNewStream)
		if !ok {
			t.Fatalf("Not New stream frame")
		}

		//If this is not file we should expect data flow
		if !subfi.IsDir() {
			sentFrames = sentFrames[1:]
			frame, ok := sentFrames[0].(*FrameData)
			require.True(t, ok)
			assert.DeepEqual(t, frame.Data, mocker.fakedata)
		}

		//Each Send stream must begin from root of stream, not from our relative path
		fname := filepath.Join(rname, subfi.Name())

		assert.Equal(t, fname, newframe.Info.Name)
		sentFrames = sentFrames[1:]

		fpath := filepath.Join(rpath, subfi.Name())
		for _, subfi := range mocker.fakeDir[fpath] {
			checkFrame(subfi, fname, fname)
		}
	}

	checkFrame(fi, "", "")
}
