package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/emiraganov/goextra/osx"
	"github.com/siddontang/go-log/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenCreateFile(t *testing.T) {
	filename := "./internal/received/testfile.txt"

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Error("Fail to open file ", filename, err)
		return
	}

	file.Write([]byte("Hello my friend"))
}

func TestReceiveSendFile(t *testing.T) {
	sen, _, sin, sout, sconnected := startSender()
	rec, _, _, rconnected := startReceiver(sout, sin)

	if err := <-sconnected; err != nil {
		t.Fatal(err)
	}

	if err := <-rconnected; err != nil {
		t.Fatal(err)
	}

	if err := osx.RemoveContents("./internal/received"); err != nil {
		t.Fatal(err)
	}

	recdir, err := os.OpenFile("./internal/received", os.O_RDONLY, os.ModeDir)
	if err != nil {
		t.Fatal(err)
	}
	defer recdir.Close()

	rec.SetOutputDir(recdir.Name())

	senfile, err := os.Open("./internal/send/testfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer senfile.Close()

	t.Log("Starting sending file", senfile.Name())

	sen.SendFile(path.Base(senfile.Name()), senfile)

	senddata, err := ioutil.ReadFile("./internal/send/testfile.txt")
	require.Nil(t, err)

	assert.Eventually(t, func() bool {
		file, err := os.Open("./internal/received/testfile.txt")
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
	}, 10*time.Second, 1*time.Second, "File is not received")

}
