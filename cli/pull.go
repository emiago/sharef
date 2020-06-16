package cli

import (
	"flag"
	"fmt"
	"os"
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
	sess := streamer.NewSession(os.Stdin, os.Stdout)
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
