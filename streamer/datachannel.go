package streamer

import "github.com/pion/webrtc/v2"

func DataChannelInitFileStream() *webrtc.DataChannelInit {
	ordered := true
	maxPacketLifeTime := uint16(10000)
	return &webrtc.DataChannelInit{
		Ordered:           &ordered,
		MaxPacketLifeTime: &maxPacketLifeTime,
	}
}
