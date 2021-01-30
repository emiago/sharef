package streamer

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sharef/errx"
	"sharef/rpc"
	"strings"
	"sync"
	"time"

	webrtc "github.com/pion/webrtc/v3"
	"github.com/sirupsen/logrus"
)

const (
	SEND_BUFFER_SIZE                = 16384  // 16 * 1024
	SEND_BUFFER_AMOUNT_LOW_TRESHOLD = 524288 //512 * 1024
)

type SendStreamer struct {
	ReadFileStreamer

	channel    *webrtc.DataChannel
	streamPath string
	// sendFrameCb      func(t int, f Framer) (n uint64, err error)
	openFileReaderCb func(path string) (io.ReadCloser, error)

	bytesWritten int64
	frameCh      chan Framer
	log          logrus.FieldLogger
	Done         chan struct{}
	DoneSending  chan struct{}
	wg           sync.WaitGroup
	//Optional variables
	output        io.Writer
	streamChanges bool

	bandwithCalc rpc.StreamBandwithCalculator
}

type SendStreamerOption func(s *SendStreamer)

func openFileReader(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	return file, err
}

// func NewSendStreamer(channel *webrtc.DataChannel, streamInfo os.FileInfo, rootpath string, options ...SendStreamerOption) *SendStreamer {
func NewSendStreamer(channel *webrtc.DataChannel, rootpath string, options ...SendStreamerOption) *SendStreamer {
	r := &SendStreamer{
		channel:     channel,
		streamPath:  filepath.Clean(rootpath),
		frameCh:     make(chan Framer),
		log:         logrus.WithField("prefix", "sendstream"),
		output:      os.Stdout,
		Done:        make(chan struct{}),
		DoneSending: make(chan struct{}),
		wg:          sync.WaitGroup{},
	}

	r.bandwithCalc = rpc.NewBandwithCalc(r.output)
	r.channel.SetBufferedAmountLowThreshold(SEND_BUFFER_AMOUNT_LOW_TRESHOLD)
	// r.sendFrameCb = r.sendFrame         //Neded for mocking
	r.ReadFileStreamer = &ReadFileStreamerWebrtc{channel}
	// r.openFileReaderCb = openFileReader //Neded for mocking

	for _, opt := range options {
		opt(r)
	}
	return r
}

func (s *SendStreamer) SetOutput(w io.Writer) {
	s.output = w
}

func (s *SendStreamer) AsyncStream(streamInfo os.FileInfo) error {
	s.channel.OnMessage(s.OnMessage)
	s.channel.OnClose(s.OnClose)

	s.channel.OnOpen(func() {
		s.log.Infof("Send receive streamer open")

		if err := s.SubStream(streamInfo, s.streamPath); err != nil {
			s.log.WithError(err).Error("Failed to process file ", s.streamPath)
		}
		// close(s.DoneSending)
		close(s.Done)
	}) //On open we streaming will start

	return nil
}

func (s *SendStreamer) Stream(ctx context.Context, streamInfo os.FileInfo) error {
	if !s.openChannel(ctx) {
		return fmt.Errorf("Channel not opened, or it too early exit")
	}
	err := s.SubStream(streamInfo, s.streamPath)
	return err
}

// StreamReader is more generic function that can stream any io reader as file on other side
func (s *SendStreamer) StreamReader(ctx context.Context, reader io.Reader, info StreamFile) error {
	if !s.openChannel(ctx) {
		return fmt.Errorf("Channel not opened, or it too early exit")
	}
	err := s.processNewStream(reader, info)
	return err
}

func (s *SendStreamer) openChannel(ctx context.Context) bool {
	s.channel.OnMessage(s.OnMessage)
	s.channel.OnClose(s.OnClose)

	opened := make(chan struct{})
	s.channel.OnOpen(func() {
		close(opened)
	}) //On open we streaming will start

	select {
	case <-ctx.Done():
	case <-opened:
		return true
	}
	return false
}

func (s *SendStreamer) OnClose() {
	s.log.Infof("Send Streamer %s closed", s.channel.Label())
}

func (s *SendStreamer) SubStream(streamInfo os.FileInfo, path string) error {
	if err := s.processFile(streamInfo, path); err != nil {
		return err
	}

	//Need to slove this better, but for now we should not close our self until buffer is empty
	for {
		if s.channel.BufferedAmount() == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (s *SendStreamer) processFile(fi os.FileInfo, path string) error {
	if fi.IsDir() {
		if err := s.processFileDir(fi, path); err != nil {
			return err
		}
		return nil
	}

	if err := s.processFileStream(fi, path); err != nil {
		return err
	}

	return nil
}

func (s *SendStreamer) processFileDir(fi os.FileInfo, root string) error {
	root = strings.TrimSuffix(root, string(os.PathSeparator))

	if err := s.processFileStream(fi, root); err != nil {
		return err
	}

	finfos, err := s.ReadDir(root)
	if err != nil {
		return err
	}

	for _, fi := range finfos {
		path := filepath.Join(root, fi.Name())
		if err := s.processFile(fi, path); err != nil {
			s.log.WithError(err).Error("Failed to process file")
		}
	}

	return nil
}

// func (s *SendStreamer) openFile(path string) (io.ReadCloser, error) {
// 	file, err := os.Open(path)
// 	return file, err
// }

func (s *SendStreamer) processFileStream(fi os.FileInfo, path string) error {
	file, err := s.OpenFile(path)
	if err != nil {
		return errx.Wrapf(err, "Fail to open file %s", path)
	}
	defer file.Close()

	info := s.prepareNewStream(fi, path)

	return s.processNewStream(file, info)
}

func (s *SendStreamer) prepareNewStream(fi os.FileInfo, path string) StreamFile {
	path = filepath.Clean(path)
	//Here we need to send file, root must be our stream name
	info := FileInfo2StreamFile(fi, "")
	//Get base of our main stream
	base := filepath.Base(s.streamPath)
	//Strip our path
	mainpath := strings.TrimPrefix(s.streamPath, ".")
	mainpath = strings.TrimPrefix(mainpath, string(os.PathSeparator))
	path = strings.TrimPrefix(path, ".")
	path = strings.TrimPrefix(path, string(os.PathSeparator))

	stripped := strings.TrimPrefix(path, mainpath)

	//Relative path for receiver must be constructed
	info.Name = filepath.Join(base, stripped)

	s.log.Infof("Sending file stream %s %s", info.Name, s.streamPath)

	return info
}

func (s *SendStreamer) processNewStream(file io.Reader, info StreamFile) error {
	if _, err := s.postFrame(FRAME_NEWSTREAM, &FrameNewStream{Info: info}); err != nil {
		return errx.Wrapf(err, "Fail to post frame for file %s", info.Path)
	}

	if info.IsDir() {
		//No need to stream dir
		return nil
	}

	if err := s.streamReader(file, info.Size, info.Name); err != nil {
		return errx.Wrapf(err, "Fail to stream file %s", info.Path)
	}
	return nil
}

func (s *SendStreamer) streamReader(file io.Reader, size int64, fname string) error {
	s.log.Infof("Starting stream name=%s", fname)

	data := make([]byte, SEND_BUFFER_SIZE)
	b := s.bandwithCalc
	b.NewStream(fname, uint64(size))

	bufflock := make(chan struct{})
	s.channel.OnBufferedAmountLow(func() {
		<-bufflock
	})

	for {
		if s.channel.BufferedAmount() >= s.channel.BufferedAmountLowThreshold() {
			bufflock <- struct{}{}
		}

		n, err := file.Read(data)

		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		// _, err = s.sendFrame(FRAME_DATA, &FrameData{Data: data[:n]})
		_, err = s.SendFrame(FRAME_DATA, &FrameData{Data: data[:n]})
		if err != nil {
			return err
		}

		b.Add(uint64(n))
	}

	b.Finish()
	s.log.Infof("File %s is successfully sent", fname)
	return nil
}

func (s *SendStreamer) OnMessage(msg webrtc.DataChannelMessage) {
	// s.log.Infof("Sender on message called")

	f, err := UnmarshalFramer(msg.Data)
	if err != nil {
		s.log.Error(err)
		return
	}
	s.log.Infof("Sender on message called %d", f.GetT())

	select {
	case s.frameCh <- f:
	default:
		s.log.Errorf("Frame missed %s", f.GetT())
	}
}

func (s *SendStreamer) postFrame(t int, f Framer) (Framer, error) {
	if _, err := s.SendFrame(t, f); err != nil {
		return nil, err
	}

	var res Framer
	select {
	case res = <-s.frameCh:
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("Timeout")
	}

	if res.GetT() == FRAME_ERROR {
		frame := res.(*FrameError)
		return nil, fmt.Errorf(frame.Err)
	}

	return res, nil
}
