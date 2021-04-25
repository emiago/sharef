package streamer

import (
	webrtc "github.com/pion/webrtc/v3"
)

const (
	DatachannelLabelJsonstream     = "jsonstream"
	DatachannelLabelProtobufstream = "protostream"
)

func DataChannelInitFileStream() *webrtc.DataChannelInit {
	ordered := true
	maxPacketLifeTime := uint16(10000)
	return &webrtc.DataChannelInit{
		Ordered:           &ordered,
		MaxPacketLifeTime: &maxPacketLifeTime,
	}
}

type ReadWriteFramer interface {
	SendFrame(t int, f Framer) (n uint64, err error)
	ReadFrame(msg []byte) (f Framer, err error)
}

type DataChannelFramer struct {
	channel *webrtc.DataChannel
	encoder FrameEncoder
}

func NewDataChannelFramer(channel *webrtc.DataChannel) *DataChannelFramer {
	switch channel.Label() {
	case DatachannelLabelJsonstream:
		return &DataChannelFramer{
			channel,
			&JSONFrameEncoder{},
		}
	default:
		return &DataChannelFramer{
			channel,
			&ProtobufFrameEncoder{},
		}
	}
}

func (s *DataChannelFramer) SendFrame(t int, f Framer) (n uint64, err error) {
	data, err := s.encoder.MarshalFramer(f, t)
	if err != nil {
		return 0, err
	}

	n = uint64(len(data))
	err = s.channel.Send(data)
	return
}

func (s *DataChannelFramer) ReadFrame(msg []byte) (f Framer, err error) {
	return s.encoder.UnmarshalFramer(msg)
}
