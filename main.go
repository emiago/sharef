package main

import (
	"flag"
	"fmt"
	"os"
	"sharef/cli"
	cliconf "sharef/cli/conf"

	log_prefixed "github.com/chappjc/logrus-prefix"
	"github.com/sirupsen/logrus"
)

var verbose = flag.Int("v", 3, "verbosity")

func init() {
	logrus.SetFormatter(&log_prefixed.TextFormatter{
		FullTimestamp: true,
	})
}

func main() {
	// receive := flag.NewFlagSet("receive", flag.ExitOnError)
	flag.Usage = func() {
		s := `Usage:
  push		- Start sending/streaming files. More options are available.
  pull 		- Start receiving files.
  `

		fmt.Fprintln(os.Stderr, s)
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()
	logrus.SetLevel(logrus.Level(*verbose))
	// Read any enviroment vars for
	cliconf.ReadEnv()

	args := flag.Args()

	// We expect always some action argument:
	if len(args) == 0 {
		flag.Usage()
		return
	}
	ParseArgs(args)
}

func ParseArgs(args []string) {
	action := args[0]
	args = args[1:]
	switch action {
	case "push":
		cli.Push(args)
	case "pull":
		cli.Pull(args)
	case "deamon":
		cli.Deamon(args)
	default:
		fmt.Println("Unknown action")
	}
}
