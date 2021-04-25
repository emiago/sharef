package streamer

import (
	"io"
	"os"
	"path/filepath"

	webrtc "github.com/pion/webrtc/v3"
	"github.com/sirupsen/logrus"
)

type WriteFileStreamer interface {
	// SendFrame(t int, f Framer) (n uint64, err error)
	OpenFile(path string, mode os.FileMode) (io.WriteCloser, error)
	Mkdir(path string, mode os.FileMode) (err error)
}

type ReceiveStreamer struct {
	WriteFileStreamer
	ReadWriteFramer

	channel       *webrtc.DataChannel
	stream        io.WriteCloser
	streamInfo    StreamFile
	bandwidthCalc StreamBandwithCalculator

	bytesWritten int64
	log          logrus.FieldLogger
	outputDir    string
	output       io.Writer

	// Done chan struct{}

	// FilesCount is number of files received, which can be read at the end of stream. Not safe to read during streaming
	FilesCount int
}

func NewReceiveStreamer(channel *webrtc.DataChannel, outputDir string, fwriter WriteFileStreamer) *ReceiveStreamer {
	if outputDir == "" {
		outputDir = "."
	}

	s := &ReceiveStreamer{
		channel: channel,
		// stream:     stream,
		// streamInfo: streamInfo,
		outputDir: outputDir,
		log:       logrus.WithField("prefix", "receivestream"),
		output:    os.Stdout,
		// Done:      make(chan struct{}),
	}

	s.bandwidthCalc = NewBandwithCalc(s.output)
	s.WriteFileStreamer = fwriter
	s.ReadWriteFramer = NewDataChannelFramer(channel)
	return s
}

func (s *ReceiveStreamer) Stream() (done chan struct{}) {
	s.channel.OnOpen(s.OnOpen)
	s.channel.OnMessage(s.OnMessage)

	done = make(chan struct{})
	s.channel.OnClose(func() {
		s.log.Infof("Recive Streamer %s closed", s.channel.Label())
		close(done)
	})
	return done
}

func (s *ReceiveStreamer) OnOpen() {
	s.log.Infof("Receive send  streamer open")
	// fmt.Fprintln(s.output, "\nReceiving files:")
}

// func (s *ReceiveStreamer) OnClose() {
// 	s.log.Infof("Recive Streamer %s closed", s.channel.Label())
// 	close(s.Done)
// }

func (s *ReceiveStreamer) OnMessage(msg webrtc.DataChannelMessage) {
	f, err := s.ReadFrame(msg.Data)
	if err != nil {
		s.log.Error(err)
		return
	}

	// Each stream must be fully sent.
	// If sender wants to send new stream before finish sending current that will endup with error.
	switch m := f.(type) {
	case *FrameData:
		s.streamFrameData(m.Data)
		return
	case *FrameNewStream:
		s.log.WithField("name", m.Info.Name).Debug("Receiver new stream")
		if !s.isCurrentStreamSynced() {
			s.SendFrame(FRAME_ERROR, &FrameError{Err: "Current Stream not synced"})
			return
		}

		if err := s.handleNewStreamFrame(*m.Info); err != nil {
			s.SendFrame(FRAME_ERROR, &FrameError{Err: err.Error()})
			return
		}
	}

	s.SendFrame(FRAME_OK, &Frame{})
}

func (s *ReceiveStreamer) streamFrameData(data []byte) {
	n, err := s.stream.Write(data)
	s.bytesWritten += int64(n)
	b := s.bandwidthCalc

	if err != nil {
		s.log.Errorln(err)
		return
	}

	b.Add(uint64(n))

	if s.bytesWritten >= s.streamInfo.SizeLen {
		s.stream.Close()
		b.Finish()
	}
}

func (s *ReceiveStreamer) isCurrentStreamSynced() bool {
	if s.bytesWritten >= s.streamInfo.SizeLen {
		s.log.Info("File is fully send")
		return true
	}
	return false
}

func (s *ReceiveStreamer) handleNewStreamFrame(info StreamFile) error {
	// info.FullPath = fmt.Sprintf("%s/%s", s.outputDir, info.Name)
	info.FullPath = filepath.Join(s.outputDir, info.Name)
	s.log.Infof("Opening file %s %s", info.FullPath, info.Mode)

	if info.IsDir() {
		//If this is a directory, just create it
		if err := s.Mkdir(info.FullPath, info.FileMode()); err != nil {
			return err
		}
		return nil
	}

	//Here is problem. What if file exists, we have no way to return it.
	file, err := s.OpenFile(info.FullPath, info.FileMode())
	if err != nil {
		return err
	}

	s.stream = file
	s.streamInfo = info
	s.bandwidthCalc.NewStream(info.Name, uint64(info.SizeLen))
	s.bytesWritten = 0

	// Track some stats
	s.FilesCount++
	return nil
}
