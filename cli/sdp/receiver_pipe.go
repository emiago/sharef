package sdp

import (
	"fmt"
	"io"
	"os"
)

const (
	SDP_ANSWER_PROMPT         = "Send this answer:"
	SDP_ANSWER_WAITING_PROMPT = "Please, paste the remote offer:"
)

type STDReceiver struct {
	readPrompt bool
}

func ReceiverPipe() (io.Reader, io.Writer) {
	s := &STDReceiver{}
	return s, s
}

func (s *STDReceiver) Read(p []byte) (n int, err error) {
	if !s.readPrompt { //Read could happen multiple times, making sure we send prompt once
		fmt.Printf("%s\n\n", SDP_ANSWER_WAITING_PROMPT)
		s.readPrompt = true
	}

	n, err = os.Stdin.Read(p)
	if err != nil {
		return
	}

	return
}

func (s *STDReceiver) Write(p []byte) (n int, err error) {
	fmt.Printf("\n%s\n\n", SDP_ANSWER_PROMPT)
	n, err = os.Stdout.Write(p)
	if err != nil {
		return
	}

	return
}
