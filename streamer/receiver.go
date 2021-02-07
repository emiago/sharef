package streamer

import (
	"sharef/fsx"

	webrtc "github.com/pion/webrtc/v3"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// Session is a receiver session
type Receiver struct {
	Session
	outputDir            string
	Done                 chan struct{}
	log                  logrus.FieldLogger
	OnNewReceiveStreamer func(receiver *ReceiveStreamer)
}

func NewReceiver(s Session, outputDir string) *Receiver {
	if outputDir == "" {
		outputDir = "."
	}

	r := &Receiver{
		Session:   s,
		outputDir: outputDir,
		log:       logrus.WithField("prefix", "receiver"),
		Done:      make(chan struct{}),
		OnNewReceiveStreamer: func(receiver *ReceiveStreamer) {
			//By Default on new receive stream channel, receiver will stream
			receiver.Stream()
		},
	}

	return r
}

func (s *Receiver) Dial() error {
	if err := s.CreateConnection(s.onConnectionStateChange()); err != nil {
		log.Errorln(err)
		return err
	}

	s.OnDataChannel(func(d *webrtc.DataChannel) {
		s.log.Infof("New DataChannel %s %d\n", d.Label(), d.ID())

		receiver := NewReceiveStreamer(d, s.outputDir, fsx.NewFileWriter())
		go s.OnNewReceiveStreamer(receiver)
	})

	if err := s.ReadSDP(); err != nil {
		s.log.Errorln(err)
		return err
	}

	if err := s.CreateAnswer(); err != nil {
		s.log.Errorln(err)
		return err
	}

	s.log.Infoln("Starting to receive data...")
	return nil
}

func (s *Receiver) onConnectionStateChange() func(connectionState webrtc.ICEConnectionState) {
	return func(connectionState webrtc.ICEConnectionState) {
		s.log.Infof("ICE Connection State has changed: %s\n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateDisconnected {
			close(s.Done)
		}
	}
}
