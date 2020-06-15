package main

import (
	"fmt"
	"log"
	"os"
	"path"
)

func main() {
	fmt.Println(os.Args)
	args := os.Args[1:]
	fmt.Println(args, len(args))
	ParseArgs(args)
}

func ParseArgs(args []string) {
	// files := args
	// for _, f := range files {
	// 	fmt.Println("Sending file", f)
	// }

	sess := Session{
		sdpInput:    os.Stdin,
		sdpOutput:   os.Stdout,
		Done:        make(chan struct{}),
		stunServers: []string{"stun:stun.l.google.com:19302"},
		writer:      os.Stdout,
	}

	if len(args) == 0 {
		//Receiver
		fmt.Println("Starting receiver")
		// b := bytes.NewBuffer([]byte{})
		s := NewReceiver(sess, os.Stdout)
		if err := s.Dial(); err != nil {
			log.Fatal(err)
		}

		<-s.Done
		// fmt.Println("GOT DATA", b.String())
		return
	}

	fmt.Println("Starting sender")

	s := NewSender(sess)
	if err := s.Dial(); err != nil {
		log.Fatal(err)
	}

	for _, file := range args {
		stream, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		s.SendFile(path.Base(stream.Name()), stream)
	}
	<-s.Done
}
