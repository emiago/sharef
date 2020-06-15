package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func startSender() (sen *Sender, term *bufio.Reader, sdpInput *SafeBuffer, sdpOutput *SafeBuffer, connected chan error) {
	// reader, writer := io.Pipe()
	// term = bufio.NewReader(reader)

	sdpInput = &SafeBuffer{
		Buffer: bytes.NewBuffer([]byte{}),
	}

	sdpOutput = &SafeBuffer{
		Buffer: bytes.NewBuffer([]byte{}),
	}

	conn := Session{
		sdpInput:    sdpInput,
		sdpOutput:   sdpOutput,
		Done:        make(chan struct{}),
		stunServers: []string{"stun:stun.l.google.com:19302"},
		writer:      os.Stdout,
	}

	sen = NewSender(conn)

	connected = make(chan error)
	go func() {
		err := sen.Dial()
		connected <- err
		close(connected)
	}()

	return sen, term, sdpInput, sdpOutput, connected
}

func startReceiver(sdpInput *SafeBuffer, sdpOutput *SafeBuffer) (rec *Receiver, term *bufio.Reader, databuf *SafeBuffer, connected chan error) {
	// reader, writer := io.Pipe()
	// term = bufio.NewReader(reader)

	conn := Session{
		sdpInput:    sdpInput,
		sdpOutput:   sdpOutput,
		Done:        make(chan struct{}),
		stunServers: []string{"stun:stun.l.google.com:19302"},
		writer:      os.Stdout,
	}

	fmt.Println("Starting receiver")

	databuf = &SafeBuffer{
		Buffer: bytes.NewBuffer([]byte{}),
	}

	rec = NewReceiver(conn, databuf)

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

func TestReadStream(t *testing.T) {
	// data := []byte("Hello")

	// b := bytes.NewReader(data)
	// buf := make([]byte, 0)
	// n, err := b.Read(buf)

	// t.Log("Got stream", string(buf), n, err)

	reader := strings.NewReader("Clear is better than clever")
	p := make([]byte, 4)
	for {
		n, err := reader.Read(p)
		if err == io.EOF {
			break
		}
		fmt.Println(string(p[:n]))
	}

}

func TestReceiveSend(t *testing.T) {
	senddata := []byte("Hello my friend")
	// reader, writer, _ := os.Pipe()
	// os.Stdout = writer
	// term := bufio.NewReader(reader)

	sen, _, sin, sout, sconnected := startSender()
	// if err := readUntil(t, term, "Send this SDP:"); err != nil {
	// 	t.Fatal(err)
	// }
	// time.Sleep(1 * time.Second)
	rec, _, databuf, rconnected := startReceiver(sout, sin)
	rec.noFS = true //Do not create file system
	// if err := readUntil(t, term, "Send this SDP:"); err != nil {
	// 	t.Fatal(err)
	// }

	if err := <-sconnected; err != nil {
		t.Fatal(err)
	}

	if err := <-rconnected; err != nil {
		t.Fatal(err)
	}

	t.Log("Starting sender data:", string(senddata))
	b := bytes.NewReader(senddata)
	sen.SendFile("testfile.txt", b)

	expected := make([]byte, len(senddata))
	for {
		n, err := io.ReadFull(databuf, expected)

		if err == io.EOF || n == 0 {
			continue
		}

		if err != nil {
			t.Fatal(err)
		}

		break
	}

	if string(expected) != string(senddata) {
		t.Error("Data send is not received")
	}

}

func TestReadingFileDir(t *testing.T) {
	// fileinfos, err := ioutil.ReadDir("main.go")
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// for _, file := range fileinfos {
	// 	t.Log(file)
	// }

	file, err := os.Open(".")
	require.Nil(t, err)
	s, err := file.Stat()
	require.Nil(t, err)
	assert.Equal(t, s.IsDir(), true)
}
