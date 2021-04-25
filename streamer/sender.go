package streamer

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/emiraganov/sharef/fsx"

	webrtc "github.com/pion/webrtc/v3"

	"github.com/sirupsen/logrus"
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
	log         logrus.FieldLogger
	filechannel *webrtc.DataChannel
}

// New creates a new sender session
func NewSender(s Session) *Sender {
	return &Sender{
		Session: s,
		log:     logrus.WithField("prefix", "sender"),
	}
}

// Change encoding.
// proto = binary encoding with protobuf
// json = json encoding

// Start the connection and the file transfer
func (s *Sender) Dial() error {
	connected := make(chan struct{})

	if err := s.CreateConnection(s.onConnectionStateChange(connected)); err != nil {
		return err
	}

	channel, err := s.createFileChannel()
	if err != nil {
		return err
	}
	s.log.Debugf("Channel created %s", channel.Label())

	if err := s.CreateOffer(); err != nil {
		return err
	}

	if err := s.ReadSDP(); err != nil {
		return err
	}

	s.log.Debug("Waiting for connection before sending")
	select {
	case <-connected:
	case <-time.After(10 * time.Second):
		return fmt.Errorf("Fail to get connected")
	}

	s.filechannel = channel

	return nil
}

func (s *Sender) DialReverse() error {
	onfilechannel := make(chan struct{})
	if err := s.CreateConnection(nil); err != nil {
		return err
	}

	s.OnDataChannel(func(d *webrtc.DataChannel) {
		s.log.Infof("New DataChannel %s %d\n", d.Label(), d.ID())
		s.filechannel = d
		close(onfilechannel)
	})

	if err := s.ReadSDP(); err != nil {
		s.log.Errorln(err)
		return err
	}

	if err := s.CreateAnswer(); err != nil {
		s.log.Errorln(err)
		return err
	}

	s.log.Debug("Waiting for connection before sending")
	select {
	case <-onfilechannel:
	case <-time.After(10 * time.Second):
		return fmt.Errorf("Fail to get connected")
	}

	return nil
}

func (s *Sender) SendFile(dest string) (err error) {
	fi, err := os.Stat(dest)
	if err != nil {
		return err
	}

	sender := s.NewFileStreamer(dest)

	if err := sender.Stream(context.Background(), fi); err != nil {
		return err
	}

	return err
}

func (s *Sender) NewFileStreamer(dest string) (streamer *SendStreamer) {
	return NewSendStreamer(s.filechannel, dest, fsx.NewFileReader())
}

func (s *Sender) NewFileStreamerWithReader(dest string, freader ReadFileStreamer) (streamer *SendStreamer) {
	return NewSendStreamer(s.filechannel, dest, freader)
}

func (s *Sender) onConnectionStateChange(connected chan struct{}) func(connectionState webrtc.ICEConnectionState) {
	once := &sync.Once{}
	return func(sig webrtc.ICEConnectionState) {
		s.log.Debug("ICE STATE: ", sig.String())
		if sig == webrtc.ICEConnectionStateConnected {
			once.Do(func() {
				close(connected)
			})
		}
	}
}

// func (s *Sender) onConnectionStateChange(connected chan struct{}) func(connectionState webrtc.ICEConnectionState) {
// 	once := &sync.Once{}
// 	return func(sig webrtc.ICEConnectionState) {
// 		s.log.Debug("ICE STATE: ", sig.String())
// 		if sig == webrtc.ICEConnectionStateConnected {
// 			once.Do(func() {
// 				close(connected)
// 			})
// 		}
// 	}
// }
