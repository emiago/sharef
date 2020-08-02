package cli

import (
	"flag"
	"fmt"
	"os"
	"sharef/deamon"
	"sharef/streamer"

	log "github.com/sirupsen/logrus"
)

func Deamon(args []string) {
	flagset := flag.NewFlagSet("deamon", flag.ExitOnError)
	flagset.Parse(args)

	//Deamon start
	deamonizeMe()
}

func deamonizeMe() {
	sess := streamer.NewSession(os.Stdin, os.Stdout)

	s := streamer.NewSender(sess)
	if err := s.Dial(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Running in daemon mode...")
	if err := deamon.StartSenderDaemonServer(s, 9876); err != nil {
		fmt.Printf("Fail to star deamon: %s\n", err)
	}
}
