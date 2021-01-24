package streamer

import (
	webrtc "github.com/pion/webrtc/v3"
)

func DataChannelInitFileStream() *webrtc.DataChannelInit {
	ordered := true
	maxPacketLifeTime := uint16(10000)
	return &webrtc.DataChannelInit{
		Ordered:           &ordered,
		MaxPacketLifeTime: &maxPacketLifeTime,
	}
}
