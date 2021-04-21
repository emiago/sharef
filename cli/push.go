package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/emiraganov/sharef/cli/sdp"
	"github.com/emiraganov/sharef/deamon"
	"github.com/emiraganov/sharef/errx"
	"github.com/emiraganov/sharef/streamer"
	"github.com/emiraganov/sharef/watcher"

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
	// cmd.Wait()
}

func sendFiles(args []string, keepsync bool) error {
	//Check do file exists
	for _, file := range args {
		if !osx.CheckFileExists(file) {
			return fmt.Errorf("File %s does not exist", file)
		}
	}

	//Sender
	reader, writer := sdp.SenderPipe() //This will send prompts and offer/answer from stdin,stdout
	sess := streamer.NewSession(reader, writer)
	s := streamer.NewSender(sess)

	if err := s.Dial(); err != nil {
		return errx.Wrapf(err, "Dial failed")
	}
	defer s.Close()

	//Stream files
	fmt.Println("")
	fmt.Println("Sending files:")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, file := range args {
		fi, err := os.Stat(file)
		if err != nil {
			return err
		}

		streamer := s.NewFileStreamer(file)

		if keepsync {
			w := watcher.New(file, fi)
			go w.ListenChangeFile(ctx, func(fin os.FileInfo, path string) error {
				return streamer.SubStream(fin, path)
			})
		}

		if err := streamer.Stream(ctx, fi); err != nil {
			return errx.Wrapf(err, "Streaming %s file failed", file)
		}
	}
	return nil
}

func sendStreamerWait(s *streamer.SendStreamer, wg *sync.WaitGroup) {
	<-s.Done
	wg.Done()
}
