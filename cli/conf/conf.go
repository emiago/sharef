package conf

import (
	"fmt"
	"os"
	"sharef/streamer"
	"strings"

	"github.com/pion/webrtc/v3"
)

func ReadEnv() {
	// SHAREF_STUNS is enviroment variable to allow changin default stun server list
	if stunservers := os.Getenv("SHAREF_STUNS"); stunservers != "" {
		list := strings.Split(stunservers, ",")
		stuns := make([]string, 0, len(list))
		for _, st := range strings.Split(stunservers, ",") {
			stuns = append(stuns, fmt.Sprintf("stun:%s", st))
		}
		streamer.ICEServerList = []webrtc.ICEServer{
			{URLs: stuns},
		}
	}

}
