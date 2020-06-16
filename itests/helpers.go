package itests

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sharef/streamer"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/emiraganov/goextra/osx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SafeBuffer struct {
	*bytes.Buffer
	mu sync.RWMutex
}

func (b *SafeBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.Write(p)
}

func (b *SafeBuffer) WriteString(s string) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.WriteString(s)
}

func (b *SafeBuffer) Read(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.Read(p)
}

// func (b *SafeBuffer) ReadString(p []byte) (n int, err error) {
// 	b.mu.Lock()
// 	defer b.mu.Unlock()
// 	return b.Buffer.Read(p)
// }

func startSender() (sen *streamer.Sender, term *bufio.Reader, sdpInput *SafeBuffer, sdpOutput *SafeBuffer, connected chan error) {
	// reader, writer := io.Pipe()
	// term = bufio.NewReader(reader)

	sdpInput = &SafeBuffer{
		Buffer: bytes.NewBuffer([]byte{}),
	}

	sdpOutput = &SafeBuffer{
		Buffer: bytes.NewBuffer([]byte{}),
	}

	conn := streamer.NewSession(sdpInput, sdpOutput)
	sen = streamer.NewSender(conn)

	connected = make(chan error)
	go func() {
		err := sen.Dial()
		connected <- err
		close(connected)
	}()

	return sen, term, sdpInput, sdpOutput, connected
}

func startReceiver(sdpInput *SafeBuffer, sdpOutput *SafeBuffer, outputDir string) (rec *streamer.Receiver, term *bufio.Reader, connected chan error) {
	// reader, writer := io.Pipe()
	// term = bufio.NewReader(reader)

	conn := streamer.NewSession(sdpInput, sdpOutput)
	rec = streamer.NewReceiver(conn, outputDir)

	connected = make(chan error)
	go func() {
		err := rec.Dial()
		connected <- err
		close(connected)
	}()

	return
}

func readUntil(t *testing.T, term *bufio.Reader, match string) error {
	for {
		s, err := term.ReadString('\n')

		if err == io.EOF {
			continue
		}

		if err != nil {
			return err
		}

		t.Log(s)
		if strings.HasPrefix(s, match) {
			break
		}
	}
	return nil
}

func IsEqualDirectory(aroot, broot string, noSize bool) bool {
	res := true

	filepath.Walk(aroot, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			res = false
			return err
		}

		if path == "" {
			//Skip root
			return nil
		}

		dest := filepath.Join(broot, strings.TrimPrefix(path, filepath.Dir(aroot)))
		fo, err := os.Stat(dest)
		if err != nil {
			res = false
			return err
		}

		res = res && fi.IsDir() == fo.IsDir()
		res = res && (fi.Size() == fo.Size() || noSize)
		res = res && fi.Mode() == fo.Mode()
		return nil
	})

	return res
}

func testFileContentAreSame(t *testing.T, senddata []byte, filename string) bool {
	file, err := os.Open(filename)
	if err == os.ErrNotExist {
		t.Log("File does not exists")
		return false
	}
	require.Nil(t, err)
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	require.Nil(t, err)

	if string(data) != string(senddata) {
		t.Log("Data in files is not same")
		return false
	}

	return true
}

func assertSendfile2Outputfile(t *testing.T, newdata []byte, sendfile string, outputFile string) {
	err := ioutil.WriteFile(sendfile, newdata, 0644)
	if err != nil {
		t.Fatal(err, err.Error())
	}

	assert.Eventually(t, func() bool {
		return testFileContentAreSame(t, newdata, outputFile)
	}, 5*time.Second, 1*time.Second, "File is not synced sendfile=%s outputfile=%s", sendfile, outputFile)
}

func assertSendDir2OutputDir(t *testing.T, sendfile string, outputFile string) {
	err := os.MkdirAll(sendfile, 0744)
	if err != nil {
		t.Fatal(err, err.Error())
	}

	assert.Eventually(t, func() bool {
		return osx.CheckFileExists(outputFile)
	}, 5*time.Second, 1*time.Second, "File is not synced sendfile=%s outputfile=%s", sendfile, outputFile)
}

func assertOutputFileContent(t *testing.T, newdata []byte, outputFile string) {
	assert.Eventually(t, func() bool {
		return testFileContentAreSame(t, newdata, outputFile)
	}, 5*time.Second, 1*time.Second, "File is not received", outputFile)
}

//This Class should just be extended
type SenderReceiverConnector struct {
	sender   *streamer.Sender
	receiver *streamer.Receiver
}

func (s *SenderReceiverConnector) SetupConnection(outputDir string) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0774)
	} else {
		if err := osx.RemoveContents(outputDir); err != nil {
			return err
		}
	}

	sen, _, sin, sout, sconnected := startSender()
	rec, _, rconnected := startReceiver(sout, sin, outputDir)

	if err := <-sconnected; err != nil {
		return err
	}

	if err := <-rconnected; err != nil {
		return err
	}

	s.sender = sen
	s.receiver = rec
	return nil
}
