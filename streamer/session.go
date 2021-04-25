package streamer

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"

	webrtc "github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"
)

const (
	SDP_OFFER_PROMPT          = "Send this offer:"
	SDP_OFFER_WAITING_PROMPT  = "Please, paste the remote offer:"
	SDP_ANSWER_PROMPT         = "Send this answer:"
	SDP_ANSWER_WAITING_PROMPT = "Please, paste the remote answer:"
)

var (
	ICEServerList = []webrtc.ICEServer{
		{
			URLs: []string{
				"stun:stun.l.google.com:19302",
			},
		},
	}
)

type CompletionHandler func()

// Session contains common elements to perform send/receive
type Session struct {
	sdpReader      io.Reader
	sdpWriter      io.Writer
	peerConnection *webrtc.PeerConnection
	onCompletion   CompletionHandler
	encoding       FrameEncoding
}

// New creates a new Session
func NewSession(SDPReader io.Reader, SDPWriter io.Writer) Session {
	sess := Session{
		sdpReader: SDPReader,
		sdpWriter: SDPWriter,
		encoding:  FrameEncodingProto,
	}

	return sess
}

func (s *Session) Close() error {
	if s.peerConnection != nil {
		return s.peerConnection.Close()
	}

	return nil
}

func (s *Session) SetEncoding(enc FrameEncoding) {
	s.encoding = enc
}

// CreateConnection prepares a WebRTC connection
func (s *Session) CreateConnection(onConnectionStateChange func(connectionState webrtc.ICEConnectionState)) error {
	config := webrtc.Configuration{
		ICEServers:         ICEServerList,
		ICETransportPolicy: webrtc.ICETransportPolicyAll,
		BundlePolicy:       webrtc.BundlePolicyBalanced,
		// SDPSemantics: webrtc.SDPSemanticsPlanB,
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

	for {
		encoded, err := MustReadStream(s.sdpReader)
		if err == nil {
			if err := Decode(encoded, &sdp); err == nil {
				break
			}
		}
		return err
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
	// s.peerConnection.AddTransceiverFromKind(webrtc.RTPCodecType(webrtc.RTPTransceiverDirectionSendrecv))

	answer, err := s.peerConnection.CreateAnswer(nil)
	if err != nil {
		return err
	}

	if err := s.createSessionDescription(answer, SDP_ANSWER_PROMPT); err != nil {
		return err
	}

	return nil
}

// CreateOffer set the local description and print the offer SDP
func (s *Session) CreateOffer() error {
	// Create an offer
	// s.peerConnection.AddTransceiverFromKind(webrtc.RTPCodecType(webrtc.RTPTransceiverDirectionSendrecv))
	// s.peerConnection.AddTransceiverFromKind(webrtc.RTPCodecType(webrtc.))

	offer, err := s.peerConnection.CreateOffer(nil)
	if err != nil {
		return err
	}

	if err := s.createSessionDescription(offer, SDP_OFFER_PROMPT); err != nil {
		return err
	}

	return nil
}

// createSessionDescription set the local description and print the SDP
func (s *Session) createSessionDescription(desc webrtc.SessionDescription, prompt string) error {
	// Sets the LocalDescription, and starts our UDP listeners
	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(s.peerConnection)

	if err := s.peerConnection.SetLocalDescription(desc); err != nil {
		return err
	}

	<-gatherComplete

	desc = *s.peerConnection.LocalDescription()
	log.Debugf("Gather ICE completed %s\n", desc.SDP)

	// writer the SDP in base64 so we can paste it in browser
	resp, err := Encode(desc)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.sdpWriter, "%s\n", resp)
	return nil
}

func (s *Session) onConnectionStateConnected(connected chan struct{}) func(connectionState webrtc.ICEConnectionState) {
	once := &sync.Once{}
	return func(sig webrtc.ICEConnectionState) {
		log.Debug("ICE STATE: ", sig.String())
		if sig == webrtc.ICEConnectionStateConnected {
			once.Do(func() {
				close(connected)
			})
		}
	}
}

func (s *Session) createFileChannel() (*webrtc.DataChannel, error) {
	label := DatachannelLabelProtobufstream
	switch s.encoding {
	case FrameEncodingJSON:
		label = DatachannelLabelJsonstream
	}

	dataChannel, err := s.peerConnection.CreateDataChannel(label, DataChannelInitFileStream())
	return dataChannel, err
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
