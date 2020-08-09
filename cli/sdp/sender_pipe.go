package sdp

import (
	"fmt"
	"io"
	"os"
)

const (
	SDP_OFFER_PROMPT         = "Send this offer:"
	SDP_OFFER_WAITING_PROMPT = "Please, paste the remote offer:"
)

type STDSender struct {
	reader io.Reader
	writer io.Writer
}

func SenderPipe() (io.Reader, io.Writer) {
	s := &STDSender{
		reader: os.Stdin,
		writer: os.Stdout,
	}
	return os.Stdin, s
}

// func (s *STDSender) Read(p []byte) (n int, err error) {
// 	n, err = s.reader.Read(p)
// 	if err != nil {
// 		return
// 	}
// 	return
// }

func (s *STDSender) Write(p []byte) (n int, err error) {
	fmt.Printf("%s\n\n", SDP_OFFER_PROMPT)
	n, err = s.writer.Write(p)
	if err != nil {
		return
	}

	fmt.Printf("\n%s\n\n", SDP_OFFER_WAITING_PROMPT)
	return
}
