package cli

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sharef/deamon"
	"sharef/errx"
	"sharef/streamer"
	"sync"
	"time"

	"github.com/emiraganov/goextra/osx"
	log "github.com/sirupsen/logrus"
)

func Push(args []string) {
	flagset := flag.NewFlagSet("push", flag.ExitOnError)
	var daemonize = flagset.Bool("d", false, "- Daemonize Sender, you must kill it")
	var keepsync = flag.Bool("f", false, "- Stream/Sync files")

	flagset.Parse(args)
	args = flagset.Args()

	//Check do we deamonize
	if *daemonize {
		bootstrapSenderDeamon()
		return
	}

	//Check do we have deamon running
	cdaemon := deamon.InitSenderDeamonClient()
	if cdaemon != nil {
		cdaemon.ProcessArgs(args)
		return
	}

	//Proceed with normal streaming
	if err := sendFiles(args, *keepsync); err != nil {
		fmt.Println(err.Error())
	}
}

func bootstrapSenderDeamon() {
	name := os.Args[0]
	if name == "" {
		name = "sharef"
	}

	cmd := exec.Command(name, "deamon")

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	fmt.Println("Starting deamon, please wait...")
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	//Now client needs to fulfil SDP connections
	var running bool
	for i := 0; i < 300; i++ {
		time.Sleep(1 * time.Second) //Give some timeout for boot

		s := deamon.InitSenderDeamonClient()
		if s != nil {
			running = true
			fmt.Println("Deamon is up and running")
			s.Close() //Close connection
			break
		}
	}

	if !running {
		fmt.Println("Timeout")
		cmd.Process.Kill()
		return
	}

	if err := cmd.Process.Release(); err != nil {
		log.Fatal(err)
	}
}

func sendFiles(args []string, keepsync bool) error {
	//Check do file exists
	for _, file := range args {
		if !osx.CheckFileExists(file) {
			return fmt.Errorf("File %s does not exist", file)
		}
	}

	//Sender
	sess := streamer.NewSession(os.Stdin, os.Stdout)
	s := streamer.NewSender(sess)
	if err := s.Dial(); err != nil {
		return errx.Wrapf(err, "Dial failed")
	}
	// defer s.Close()

	//Stream files
	fmt.Println("")
	fmt.Println("Sending files:")
	wg := sync.WaitGroup{}

	var streamopt []streamer.SendStreamerOption
	if keepsync {
		streamopt = append(streamopt, streamer.WithStreamChanges())
	}

	for _, file := range args {
		wg.Add(1)
		// s.SendFile(file, streamopt...)
		streamer, err := s.InitFileStreamer(file, streamopt...)
		if err != nil {
			return errx.Wrapf(err, "Sending %s file failed", file)
		}

		if err := streamer.Stream(); err != nil {
			return errx.Wrapf(err, "Streaming %s file failed", file)
		}

		<-streamer.DoneSending
		go sendStreamerWait(streamer, &wg)
	}
	wg.Wait()
	return nil
}

func sendStreamerWait(s *streamer.SendStreamer, wg *sync.WaitGroup) {
	<-s.Done
	wg.Done()
}
