package cli

import (
	"flag"
	"fmt"

	"github.com/emiraganov/sharef/cli/sdp"
	"github.com/emiraganov/sharef/streamer"

	log "github.com/sirupsen/logrus"
)

func Pull(args []string) {
	flagset := flag.NewFlagSet("pull", flag.ExitOnError)
	flagset.Parse(args)

	//Receiver
	receiveFiles()
}

func receiveFiles() {
	reader, writer := sdp.ReceiverPipe() //This will send prompts and offer/answer from stdin,stdout
	sess := streamer.NewSession(reader, writer)
	r := streamer.NewReceiver(sess)

	if err := r.Dial(); err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	streamer := r.NewFileStreamer("")

	fmt.Println("")
	fmt.Println("Receiving files:")
	<-streamer.Stream()
}
