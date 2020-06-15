package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
)

type SendStreamer struct {
	stream  io.Reader
	channel *webrtc.DataChannel
}

func (s *SendStreamer) OnOpen() {
	log.Info("Starting send streamer ", s.channel.Label())
	// go s.writeStream()
	s.streamFile()
	// s.readStream()
	// s.channel.Close()
}

func (s *SendStreamer) streamFile() error {
	data := make([]byte, 4096)
	for {
		n, err := s.stream.Read(data)

		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		if err := s.channel.Send(data[:n]); err != nil {
			return err
		}
	}

	return nil
}

func (s *SendStreamer) OnClose() {
	log.Infof("Send Streamer %s closed", s.channel.Label())
}

type ReceiveStreamer struct {
	stream  io.Writer
	channel *webrtc.DataChannel
}

func (s *ReceiveStreamer) OnMessage(msg webrtc.DataChannelMessage) {
	log.Infof("Receiver on message called")
	n, err := s.stream.Write(msg.Data)

	if err != nil {
		log.Errorln(err)
	} else {
		// currentSpeed := s.NetworkStats.Bandwidth()
		// fmt.Printf("Transferring at %.2f MB/s\r", currentSpeed)
		log.Infof("Transferring %d \n", uint64(n))
		// s.NetworkStats.AddBytes(uint64(n))
	}
}

func (s *ReceiveStreamer) OnClose() {
	log.Infof("Recive Streamer %s closed", s.channel.Label())
}

type FileChangeStreamer struct {
	stream  bufio.ReadWriter
	channel *webrtc.DataChannel
}

func (s *FileChangeStreamer) OnOpen() {
	log.Info("Starting send streamer ", s.channel.Label())
	// go s.writeStream()
	s.readStream()
	// s.readStream()
	s.channel.Close()
}

func (s *FileChangeStreamer) readStream() error {
	data := make([]byte, 4096)
	for {
		n, err := s.stream.Read(data)

		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		if err := s.channel.Send(data[:n]); err != nil {
			return err
		}
	}

	return nil
}

func (s *FileChangeStreamer) OnMessage(msg webrtc.DataChannelMessage) {
	// s.streamFile()
	n, err := s.stream.Write(msg.Data)

	if err != nil {
		log.Errorln(err)
	} else {
		// currentSpeed := s.NetworkStats.Bandwidth()
		// fmt.Printf("Transferring at %.2f MB/s\r", currentSpeed)
		fmt.Fprintf(os.Stdout, "Transferring %d \n", uint64(n))
		// s.NetworkStats.AddBytes(uint64(n))
	}

}

func (s *FileChangeStreamer) OnClose() {
	log.Infof("Recive Streamer %s closed", s.channel.Label())
}
