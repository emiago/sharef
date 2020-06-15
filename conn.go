package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pion/webrtc/v2"
	"github.com/siddontang/go-log/log"
)

type CompletionHandler func()

// Session contains common elements to perform send/receive
type Session struct {
	Done           chan struct{}
	sdpInput       io.Reader
	sdpOutput      io.Writer
	peerConnection *webrtc.PeerConnection
	onCompletion   CompletionHandler
	stunServers    []string
	writer         io.Writer
}

// New creates a new Session
func NewConn() Session {
	sess := Session{
		sdpInput:    os.Stdin,
		sdpOutput:   os.Stdout,
		Done:        make(chan struct{}),
		stunServers: []string{"stun:stun.l.google.com:19302"},
	}

	return sess
}

// CreateConnection prepares a WebRTC connection
func (s *Session) CreateConnection(onConnectionStateChange func(connectionState webrtc.ICEConnectionState)) error {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: s.stunServers,
			},
		},
	}

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return err
	}
	s.peerConnection = peerConnection
	peerConnection.OnICEConnectionStateChange(onConnectionStateChange)

	return nil
}

// ReadSDP from the SDP input stream
func (s *Session) ReadSDP() error {
	var sdp webrtc.SessionDescription

	fmt.Fprintln(s.writer, "Please, paste the remote SDP:")
	for {
		encoded, err := MustReadStream(s.sdpInput)
		if err == nil {
			if err := Decode(encoded, &sdp); err == nil {
				break
			}
			return err
		}
		fmt.Fprintln(s.writer, "Invalid SDP, try again...")
	}

	return s.peerConnection.SetRemoteDescription(sdp)
}

// OnDataChannel sets an OnDataChannel handler
func (s *Session) OnDataChannel(handler func(d *webrtc.DataChannel)) {
	s.peerConnection.OnDataChannel(handler)
}

// CreateAnswer set the local description and print the answer SDP
func (s *Session) CreateAnswer() error {
	// Create an answer
	answer, err := s.peerConnection.CreateAnswer(nil)
	if err != nil {
		return err
	}
	return s.createSessionDescription(answer)
}

// CreateOffer set the local description and print the offer SDP
func (s *Session) CreateOffer() error {
	// Create an offer
	answer, err := s.peerConnection.CreateOffer(nil)
	if err != nil {
		return err
	}
	return s.createSessionDescription(answer)
}

// createSessionDescription set the local description and print the SDP
func (s *Session) createSessionDescription(desc webrtc.SessionDescription) error {
	// Sets the LocalDescription, and starts our UDP listeners
	if err := s.peerConnection.SetLocalDescription(desc); err != nil {
		return err
	}
	desc.SDP = StripSDP(desc.SDP)

	// writer the SDP in base64 so we can paste it in browser
	resp, err := Encode(desc)
	if err != nil {
		return err
	}
	fmt.Fprintln(s.writer, "Send this SDP:")
	fmt.Fprintf(s.sdpOutput, "%s\n", resp)
	return nil
}

func (s *Session) onConnectionStateChange() func(connectionState webrtc.ICEConnectionState) {
	return func(connectionState webrtc.ICEConnectionState) {
		log.Infof("ICE Connection State has changed: %s\n", connectionState.String())
		// if connectionState == webrtc.ICEConnectionStateDisconnected {
		// 	s.stopSending <- struct{}{}
		// }
	}
}

// MustReadStream blocks until input is received from the stream
func MustReadStream(stream io.Reader) (string, error) {
	r := bufio.NewReader(stream)

	var in string
	for {
		var err error
		in, err = r.ReadString('\n')
		if err != io.EOF {
			if err != nil {
				return "", err
			}
		}
		in = strings.TrimSpace(in)
		if len(in) > 0 {
			break
		}
	}

	return in, nil
}

// StripSDP remove useless elements from an SDP
func StripSDP(originalSDP string) string {
	finalSDP := strings.Replace(originalSDP, "a=group:BUNDLE audio video data", "a=group:BUNDLE data", -1)
	tmp := strings.Split(finalSDP, "m=audio")
	beginningSdp := tmp[0]

	var endSdp string
	if len(tmp) > 1 {
		tmp = strings.Split(tmp[1], "a=end-of-candidates")
		endSdp = strings.Join(tmp[2:], "a=end-of-candidates")
	} else {
		endSdp = strings.Join(tmp[1:], "a=end-of-candidates")
	}

	finalSDP = beginningSdp + endSdp
	finalSDP = strings.Replace(finalSDP, "\r\n\r\n", "\r\n", -1)
	finalSDP = strings.Replace(finalSDP, "\n\n", "\n", -1)
	return finalSDP
}

// Encode encodes the input in base64
// It can optionally zip the input before encoding
func Encode(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	var gzbuff bytes.Buffer
	gz, err := gzip.NewWriterLevel(&gzbuff, gzip.BestCompression)
	if err != nil {
		return "", err
	}
	if _, err := gz.Write(b); err != nil {
		return "", err
	}
	if err := gz.Flush(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzbuff.Bytes()), nil
}

// Decode decodes the input from base64
// It can optionally unzip the input after decoding
func Decode(in string, obj interface{}) error {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return err
	}

	gz, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer gz.Close()
	s, err := ioutil.ReadAll(gz)
	if err != nil {
		return err
	}

	return json.Unmarshal(s, obj)
}
