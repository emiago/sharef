package streamer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sharef/errx"
	"sharef/rpc"
	"strings"
	"sync"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
)

const (
	SEND_BUFFER_SIZE                = 16384  // 16 * 1024
	SEND_BUFFER_AMOUNT_LOW_TRESHOLD = 524288 //512 * 1024
)

type SendStreamer struct {
	ReadFileStreamer

	channel    *webrtc.DataChannel
	stream     *os.File
	streamInfo os.FileInfo
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
}

type SendStreamerOption func(s *SendStreamer)

func openFileReader(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	return file, err
}

func NewSendStreamer(channel *webrtc.DataChannel, streamInfo os.FileInfo, path string, options ...SendStreamerOption) *SendStreamer {
	r := &SendStreamer{
		channel:     channel,
		streamInfo:  streamInfo,
		streamPath:  filepath.Clean(path),
		frameCh:     make(chan Framer),
		log:         logrus.WithField("prefix", "sendstream"),
		output:      os.Stdout,
		Done:        make(chan struct{}),
		DoneSending: make(chan struct{}),
		wg:          sync.WaitGroup{},
	}

	r.channel.SetBufferedAmountLowThreshold(SEND_BUFFER_AMOUNT_LOW_TRESHOLD)
	// r.sendFrameCb = r.sendFrame         //Neded for mocking
	r.ReadFileStreamer = &ReadFileStreamerWebrtc{channel}
	// r.openFileReaderCb = openFileReader //Neded for mocking

	for _, opt := range options {
		opt(r)
	}
	return r
}

func WithStreamChanges() SendStreamerOption {
	return func(s *SendStreamer) {
		s.streamChanges = true
	}
}

func (s *SendStreamer) SetOutput(w io.Writer) {
	s.output = w
}

func (s *SendStreamer) Stream() error {
	s.channel.OnOpen(s.OnOpen) //On open we streaming will start
	s.channel.OnMessage(s.OnMessage)
	s.channel.OnClose(s.OnClose)
	return nil
}

func (s *SendStreamer) OnClose() {
	s.log.Infof("Send Streamer %s closed", s.channel.Label())
}

func (s *SendStreamer) OnOpen() {
	s.log.Infof("Send receive streamer open")

	if err := s.processFile(s.streamInfo, s.streamPath); err != nil {
		s.log.WithError(err).Error("Failed to process file ", s.streamPath)
	}

	//Need to slove this better, but for now we should not close our self until buffer is empty
	for {
		if s.channel.BufferedAmount() == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	close(s.DoneSending)

	if s.streamChanges {
		//This will block
		s.listenChangeFile()
		return
	}
	close(s.Done)
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

	if _, err := s.postFrame(FRAME_NEWSTREAM, &FrameNewStream{Info: info}); err != nil {
		return errx.Wrapf(err, "Fail to post frame for file %s", path)
	}

	if fi.IsDir() {
		//No need to stream dir
		return nil
	}

	if err := s.streamFile(file, fi); err != nil {
		return errx.Wrapf(err, "Fail to stream file %s", path)
	}

	return nil
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

func (s *SendStreamer) streamFile(file io.Reader, fi os.FileInfo) error {
	s.log.Infof("Starting stream name=%s", fi.Name())

	data := make([]byte, SEND_BUFFER_SIZE)
	b := rpc.NewBandwithCalc(uint64(fi.Size()))
	// fmt.Fprintln(s.output, "")

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
		b.FprintOnSecond(s.output, fi.Name())
	}

	fmt.Fprintln(s.output, b.Sprint(fi.Name())) //Do last print
	s.log.Infof("File %s is successfully sent bytes %f", fi.Name(), b.Total(1))
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

	s.frameCh <- f
}

func (s *SendStreamer) postFrame(t int, f Framer) (Framer, error) {
	if _, err := s.SendFrame(t, f); err != nil {
		return nil, err
	}

	res := <-s.frameCh

	if res.GetT() == FRAME_ERROR {
		frame := res.(*FrameError)
		return nil, fmt.Errorf(frame.Err)
	}

	return res, nil
}

func (s *SendStreamer) listenChangeFile() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		s.log.Println("ERROR", err)
		return
	}
	defer watcher.Close()

	if err := watcher.Add(s.streamPath); err != nil {
		s.log.Info("ERROR ", err)
	}

	for {
		select {
		// watch for events
		case event := <-watcher.Events:
			s.log.Infof("EVENT! %s\n", event.String())
			s.checkFileChanges(event, watcher)

			// watch for errors
		case err := <-watcher.Errors:
			s.log.Info("ERROR ", err)
		}
	}
}

func (s *SendStreamer) checkFileChanges(event fsnotify.Event, watcher *fsnotify.Watcher) {
	// if s.bytesWritten < s.streamInfo.Size {
	// 	return
	// }

	switch {
	case event.Op&fsnotify.Write == fsnotify.Write:
	case event.Op&fsnotify.Create == fsnotify.Create:
		if !s.streamInfo.IsDir() { //Only streaming dir we follow create changes
			return
		}
	default:
		return
	}

	path := event.Name

	fi, err := os.Stat(path)
	if err != nil {
		s.log.Error(err)
		return
	}

	s.log.WithField("path", path).Info("Sending file changes")
	if err := s.processFile(fi, path); err != nil {
		s.log.Error(err)
		return
	}

	if fi.IsDir() {
		//Add tracking changes for this dir
		if err := watcher.Add(path); err != nil {
			s.log.Error(err)
		}
	}
}
