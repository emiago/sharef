package main

import (
	"fmt"
	"io"
	"os"

	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
)

// Session is a receiver session
type Receiver struct {
	Session
	stream      io.Writer
	msgChannel  chan webrtc.DataChannelMessage
	initialized bool
	outputDir   string

	noFS bool
}

func NewReceiver(s Session, f io.Writer) *Receiver {
	r := &Receiver{
		Session:     s,
		stream:      f,
		msgChannel:  make(chan webrtc.DataChannelMessage, 4096*2),
		initialized: false,
		noFS:        false,
		outputDir:   ".",
	}

	return r
}

func (s *Receiver) Dial() error {
	if err := s.CreateConnection(s.onConnectionStateChange()); err != nil {
		log.Errorln(err)
		return err
	}

	s.OnDataChannel(func(d *webrtc.DataChannel) {
		log.Infof("New DataChannel %s %d\n", d.Label(), d.ID())
		// s.sess.NetworkStats.Start()

		switch d.Label() {
		case "data":
			return
		default:
		}
		streamer := &ReceiveStreamer{
			stream:  s.stream,
			channel: d,
		}

		if !s.noFS {
			// _, err := os.Stat(d.Label())
			// if err != nil {d
			// 	log.Error("Fail to stat file", err)
			// 	return
			// }
			// if err == os.ErrNotExist {
			// 	f, err := os.Create(d.Label())
			// 	if err != nil {
			// 		log.Error("Fail to create file", err)
			// 		return
			// 	}
			// 	f.Close()
			// }

			dest := fmt.Sprintf("%s/%s", s.outputDir, d.Label())
			log.Info("Creating file ", dest)

			file, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				log.Error("Fail to open file ", dest, err)
				return
			}

			streamer.stream = file
		}

		d.OnMessage(streamer.OnMessage)
		d.OnClose(streamer.OnClose)

		// s.streamersMux.Lock()
		// defer s.streamersMux.Unlock()
		// s.streamers[d.Label()] = streamer
	})

	if err := s.ReadSDP(); err != nil {
		log.Errorln(err)
		return err
	}

	if err := s.CreateAnswer(); err != nil {
		log.Errorln(err)
		return err
	}

	log.Infoln("Starting to receive data...")
	return nil
}

func (s *Receiver) SetOutputDir(dir string) {
	s.outputDir = dir
}

func (s *Receiver) onConnectionStateChange() func(connectionState webrtc.ICEConnectionState) {
	return func(connectionState webrtc.ICEConnectionState) {
		log.Infof("ICE Connection State has changed: %s\n", connectionState.String())
	}
}
