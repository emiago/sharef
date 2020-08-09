package cli

import (
	"flag"
	"fmt"
	"sharef/cli/sdp"
	"sharef/streamer"
	"sync"

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
	r := streamer.NewReceiver(sess, "")

	wg := sync.WaitGroup{}
	wg.Add(1)
	//We expect only one receiver in this case
	r.OnNewReceiveStreamer = func(receiver *streamer.ReceiveStreamer) {
		fmt.Println("")
		fmt.Println("Receiving files:")
		receiver.Stream()
		<-receiver.Done
		wg.Done()
	}

	if err := r.Dial(); err != nil {
		log.Fatal(err)
	}

	wg.Wait()
}
