package streamer

import (
	"fmt"
	"time"

	"github.com/emiraganov/sharef/fsx"

	webrtc "github.com/pion/webrtc/v3"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// Session is a receiver session
type Receiver struct {
	Session
	Done chan struct{}
	log  logrus.FieldLogger

	filechannel *webrtc.DataChannel
}

func NewReceiver(s Session) *Receiver {

	r := &Receiver{
		Session: s,
		log:     logrus.WithField("prefix", "receiver"),
	}

	return r
}

func (s *Receiver) Dial() error {
	onfilechannel := make(chan struct{})
	if err := s.CreateConnection(s.onConnectionStateChange()); err != nil {
		log.Errorln(err)
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

	s.log.Debug("Waiting for connection before receiving")
	select {
	case <-onfilechannel:
	case <-time.After(10 * time.Second):
		return fmt.Errorf("Fail to get connected")
	}

	return nil
}

func (s *Receiver) DialReverse() error {
	connected := make(chan struct{})
	if err := s.CreateConnection(s.onConnectionStateConnected(connected)); err != nil {
		return err
	}

	d, err := s.createFileChannel()
	if err != nil {
		return err
	}
	s.log.Debugf("Channel created %s", d.Label())

	if err := s.CreateOffer(); err != nil {
		s.log.Errorln(err)
		return err
	}

	if err := s.ReadSDP(); err != nil {
		s.log.Errorln(err)
		return err
	}

	s.log.Debug("Waiting for connection before receiving")
	select {
	case <-connected:
	case <-time.After(10 * time.Second):
		return fmt.Errorf("Fail to get connected")
	}

	s.filechannel = d
	return nil
}

func (s *Receiver) NewFileStreamer(rootpath string) *ReceiveStreamer {
	return NewReceiveStreamer(s.filechannel, rootpath, fsx.NewFileWriter())
}

func (s *Receiver) NewFileStreamerWithWritter(rootpath string, fwriter WriteFileStreamer) *ReceiveStreamer {
	return NewReceiveStreamer(s.filechannel, rootpath, fwriter)
}

func (s *Receiver) onConnectionStateChange() func(connectionState webrtc.ICEConnectionState) {
	return func(connectionState webrtc.ICEConnectionState) {
		s.log.Infof("ICE Connection State has changed: %s\n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateDisconnected {
			close(s.Done)
		}
	}
}
