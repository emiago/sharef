package streamer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sharef/rpc"

	"github.com/pion/webrtc/v2"
	"github.com/sirupsen/logrus"
)

type ReceiveStreamer struct {
	WriteFileStreamer

	channel        *webrtc.DataChannel
	stream         io.WriteCloser
	streamInfo     StreamFile
	streamBandwith *rpc.BandwithCalc

	bytesWritten int64
	log          logrus.FieldLogger
	outputDir    string
	output       io.Writer
	Done         chan struct{}
}

func NewReceiveStreamer(channel *webrtc.DataChannel, outputDir string) *ReceiveStreamer {
	r := &ReceiveStreamer{
		channel: channel,
		// stream:     stream,
		// streamInfo: streamInfo,
		outputDir: outputDir,
		log:       logrus.WithField("prefix", "receivestream"),
		output:    os.Stdout,
		Done:      make(chan struct{}),
	}

	r.WriteFileStreamer = &WriteFileStreamerWebrtc{channel}

	return r
}

func (s *ReceiveStreamer) Stream() {
	s.channel.OnOpen(s.OnOpen)
	s.channel.OnMessage(s.OnMessage)
	s.channel.OnClose(s.OnClose)
}

func (s *ReceiveStreamer) OnOpen() {
	s.log.Infof("Receive send  streamer open")
	// fmt.Fprintln(s.output, "\nReceiving files:")
}

func (s *ReceiveStreamer) OnClose() {
	s.log.Infof("Recive Streamer %s closed", s.channel.Label())
	close(s.Done)
}

func (s *ReceiveStreamer) OnMessage(msg webrtc.DataChannelMessage) {
	f, err := ParseFrameData(msg.Data)
	if err != nil {
		s.log.Error(err)
		return
	}
	// s.log.Infof("Receiver on message called %d", f.GetT())

	switch m := f.(type) {
	case *FrameData:
		s.streamFrameData(m.Data)
		return
	case *FrameNewStream:
		if !s.isCurrentStreamSynced() {
			s.SendFrame(FRAME_ERROR, &FrameError{Err: fmt.Errorf("Current Stream not synced")})
			return
		}

		if err := s.handleNewStreamFrame(m.Info); err != nil {
			s.SendFrame(FRAME_ERROR, &FrameError{Err: err})
			return
		}
	}

	s.SendFrame(FRAME_OK, &Frame{})
}

func (s *ReceiveStreamer) streamFrameData(data []byte) {
	n, err := s.stream.Write(data)
	s.bytesWritten += int64(n)
	b := s.streamBandwith

	if err != nil {
		s.log.Errorln(err)
		return
	}

	b.Add(uint64(n))
	b.FprintOnSecond(s.output, s.streamInfo.Name)

	if s.bytesWritten >= s.streamInfo.Size {
		fmt.Fprintln(s.output, b.Sprint(s.streamInfo.Name)) //Do last print
	}
}

func (s *ReceiveStreamer) isCurrentStreamSynced() bool {
	if s.bytesWritten >= s.streamInfo.Size {
		s.log.Info("File is fully send")
		return true
	}
	return false
}

func (s *ReceiveStreamer) handleNewStreamFrame(info StreamFile) error {
	// info.FullPath = fmt.Sprintf("%s/%s", s.outputDir, info.Name)
	info.FullPath = filepath.Join(s.outputDir, info.Name)
	s.log.Infof("Opening file %s %s", info.FullPath, info.Mode)

	if info.Mode.IsDir() {
		//If this is a directory, just create it
		if err := s.Mkdir(info.FullPath, info.Mode); err != nil {
			return err
		}
		return nil
	}

	file, err := s.OpenFile(info.FullPath, info.Mode)
	if err != nil {
		return err
	}

	s.stream = file
	s.streamInfo = info
	s.streamBandwith = rpc.NewBandwithCalc(uint64(info.Size))
	s.bytesWritten = 0

	return nil
}
