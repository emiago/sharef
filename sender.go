package main

import (
	"io"
	"os"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/siddontang/go-log/log"
)

const (
	// Must be <= 16384
	senderBuffSize = 16384
)

type outputMsg struct {
	n    int
	buff []byte
}

// Session is a sender session
type Sender struct {
	Session
	stream io.Reader

	dataChannel *webrtc.DataChannel
	readers     chan io.Reader
}

// New creates a new sender session
func NewSender(s Session) *Sender {
	return &Sender{
		Session: s,
		readers: make(chan io.Reader),
	}
}

// Start the connection and the file transfer
func (s *Sender) Dial() error {
	if err := s.CreateConnection(s.onConnectionStateChange()); err != nil {
		return err
	}

	if err := s.createDataChannel(); err != nil {
		return err
	}

	if err := s.CreateOffer(); err != nil {
		return err
	}

	if err := s.ReadSDP(); err != nil {
		return err
	}

	return nil
}

type StreamFile struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	Data    []byte
	// sys     syscall.Stat_t
}

func (s *Sender) SendFile(name string, f io.Reader) error {
	err := s.createFileChannel(name, f)
	return err
}

func (s *Sender) createDataChannel() error {
	ordered := true
	maxPacketLifeTime := uint16(10000)
	dataChannel, err := s.peerConnection.CreateDataChannel("data", &webrtc.DataChannelInit{
		Ordered:           &ordered,
		MaxPacketLifeTime: &maxPacketLifeTime,
	})

	if err != nil {
		return err
	}

	s.dataChannel = dataChannel
	// s.dataChannel.OnBufferedAmountLow(s.onBufferedAmountLow())
	// s.dataChannel.SetBufferedAmountLowThreshold(bufferThreshold)
	s.dataChannel.OnOpen(s.onOpenHandler())
	s.dataChannel.OnClose(s.onCloseHandler())

	return nil
}

func (s *Sender) createFileChannel(name string, f io.Reader) error {
	ordered := true
	maxPacketLifeTime := uint16(10000)
	dataChannel, err := s.peerConnection.CreateDataChannel(name, &webrtc.DataChannelInit{
		Ordered:           &ordered,
		MaxPacketLifeTime: &maxPacketLifeTime,
	})

	if err != nil {
		return err
	}

	streamer := &SendStreamer{
		stream:  f,
		channel: dataChannel,
	}

	dataChannel.OnOpen(streamer.OnOpen)
	dataChannel.OnClose(streamer.OnClose)

	return nil
}

func (s *Sender) onOpenHandler() func() {
	return func() {
		// s.sess.NetworkStats.Start()

		log.Infof("Starting to send data...")
		defer log.Infof("Stopped sending data...")

		// s.writeToNetwork()
		select {
		case stream := <-s.readers:
			if err := s.streamFile(stream); err != nil {
				log.Errorf("Error, fail to stream file: %v\n", err)
			}
		}
	}
}

func (s *Sender) streamFile(stream io.Reader) error {
	data := make([]byte, 4096)
	for {
		n, err := stream.Read(data)

		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		// Stre

		// log.Info("Sending stream: ", string(data[:n]))
		if err := s.dataChannel.Send(data[:n]); err != nil {

			return err
		}
	}

	return nil
}

func (s *Sender) onCloseHandler() func() {
	return func() {
		s.close(true)
	}
}

func (s *Sender) close(calledFromCloseHandler bool) {
	if !calledFromCloseHandler {
		s.dataChannel.Close()
	}

	// Sometime, onCloseHandler is not invoked, so it's a work-around
	s.dumpStats()
	close(s.Done)
}

func (s *Sender) dumpStats() {
	// 	fmt.Printf(`
	// Disk   : %s
	// Network: %s
	// `, s.readingStats.String(), s.sess.NetworkStats.String())
}
