package streamer

import (
	"context"
	"os"

	"github.com/pion/webrtc/v2"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
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

// Start the connection and the file transfer
func (s *Sender) Dial() error {
	if err := s.CreateConnection(s.onConnectionStateChange()); err != nil {
		return err
	}

	// if err := s.createRequestChannel(); err != nil {
	// 	return err
	// }

	if err := s.CreateOffer(); err != nil {
		return err
	}

	// fmt.Fprintln(s.writer) //Add one break
	if err := s.ReadSDP(); err != nil {
		return err
	}

	channel, err := s.createFileChannel("filestream")
	if err != nil {
		return err
	}
	s.filechannel = channel

	return nil
}

func (s *Sender) DialWithAnswerFirst() error {
	if err := s.CreateConnection(s.onConnectionStateChange()); err != nil {
		return err
	}

	if err := s.ReadSDP(); err != nil {
		s.log.Errorln(err)
		return err
	}

	if err := s.CreateAnswer(); err != nil {
		s.log.Errorln(err)
		return err
	}

	channel, err := s.createFileChannel("filestream")
	if err != nil {
		return err
	}
	s.filechannel = channel

	return nil
}

// func (s *Sender) StreamAudio(dest string) (err error) {
// 	codec := webrtc.NewRTPPCMACodec(webrtc.DefaultPayloadTypePCMA, 8000)
// 	track, _ := webrtc.NewTrack(webrtc.DefaultPayloadTypePCMA, 1, "audio", "test", codec)
// 	s.peerConnection.AddTrack(track)
// 	// rtp.NewPacketizer()
// 	// track.WriteRTP()
// 	// return s.SendFileWithOptions(dest, nil)
// }

func (s *Sender) SendFile(dest string, options ...SendStreamerOption) (err error) {
	fi, err := os.Stat(dest)
	if err != nil {
		return err
	}

	sender := s.NewFileStreamer(dest, options...)

	if err := sender.Stream(context.Background(), fi); err != nil {
		return err
	}

	return err
}

// func (s *Sender) InitFileStreamer(dest string, options ...SendStreamerOption) (streamer *SendStreamer, err error) {
// 	fi, err := os.Stat(dest)
// 	if err != nil {
// 		return nil, err
// 	}

// 	sender := NewSendStreamer(s.filechannel, fi, dest, options...)
// 	return sender, nil
// }

func (s *Sender) NewFileStreamer(dest string, options ...SendStreamerOption) (streamer *SendStreamer) {
	return NewSendStreamer(s.filechannel, dest, options...)
}

func (s *Sender) createFileChannel(name string) (*webrtc.DataChannel, error) {
	dataChannel, err := s.peerConnection.CreateDataChannel(name, DataChannelInitFileStream())
	return dataChannel, err
}

func (s *Sender) onConnectionStateChange() func(connectionState webrtc.ICEConnectionState) {
	return func(connectionState webrtc.ICEConnectionState) {
		log.Infof("ICE Connection State has changed: %s\n", connectionState.String())
		// if connectionState == webrtc.ICEConnectionStateDisconnected {
		// }
	}
}

// func (s *Sender) Close() error {
// 	err := s.peerConnection.Close()
// 	return err
// }
