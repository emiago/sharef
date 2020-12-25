// +build integration

package itests

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sharef/watcher"
	"strings"
	"syscall"
	"testing"
	"time"

	log_prefixed "github.com/chappjc/logrus-prefix"
	"github.com/emiraganov/goextra/osx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)

	logrus.SetFormatter(&log_prefixed.TextFormatter{
		FullTimestamp: true,
	})
}

type SuiteSendFile struct {
	suite.Suite

	Sendfile  string
	OutputDir string

	//Do not initialize this
	SenderReceiverConnector
	outputFile string
}

func (suite *SuiteSendFile) SetupTest() {
	t := suite.T()
	outputDir := suite.OutputDir
	if err := suite.SetupConnection(outputDir); err != nil {
		t.Fatal(err)
	}
	suite.outputFile = fmt.Sprintf("%s/%s", outputDir, path.Base(suite.Sendfile))
}

func (suite *SuiteSendFile) TestReceiveFile() {
	t := suite.T()
	sen := suite.sender
	sendfile := suite.Sendfile
	outputFile := suite.outputFile

	//Make some content
	ioutil.WriteFile(sendfile, []byte("Hello My Friend"), 0644)

	//Send our file
	t.Log("Starting sending file", sendfile)
	err := sen.SendFile(sendfile)
	require.Nil(t, err)

	//Compare data received
	senddata, err := ioutil.ReadFile(sendfile)
	require.Nil(t, err)

	assert.Eventually(t, func() bool {
		return testFileContentAreSame(t, senddata, outputFile)
	}, 5*time.Second, 1*time.Second, "File is not received")
}

type SuiteStreamFile struct {
	suite.Suite

	Sendfile  string
	OutputDir string

	//Do not initialize this
	SenderReceiverConnector
	outputFile string
}

func (suite *SuiteStreamFile) SetupTest() {
	t := suite.T()
	outputDir := suite.OutputDir
	if err := suite.SetupConnection(outputDir); err != nil {
		t.Fatal(err)
	}
	suite.outputFile = fmt.Sprintf("%s/%s", outputDir, path.Base(suite.Sendfile))

	//Here we want to prepare for streaming. So file needs two be sent prior that
	sen := suite.sender
	sendfile := suite.Sendfile
	//Send our file
	t.Log("Starting sending file", sendfile)
	fi, err := os.Stat(sendfile)
	require.Nil(t, err)

	//Start listener
	w := watcher.New(sendfile, fi)
	ctx := context.Background()
	sender := sen.NewFileStreamer(sendfile, fi)
	err = sender.Stream(ctx)
	go w.ListenChangeFile(ctx, func(fi os.FileInfo, path string) error {
		return sender.SubStream(fi, path)
	})
	require.Nil(t, err)
}

func (suite *SuiteStreamFile) TestStreamFile() {
	t := suite.T()
	sendfile := suite.Sendfile
	outputfile := suite.outputFile

	//Compare data received
	senddata, err := ioutil.ReadFile(sendfile)
	require.Nil(t, err)
	assertSendfile2Outputfile(t, senddata, sendfile, outputfile)

	//Add line
	newdata := append(senddata, []byte("\nSomething on new line")...)
	assertSendfile2Outputfile(t, newdata, sendfile, outputfile)

	// newdata = append(newdata, []byte("\nSomething on new second line")...)
	// assertSendfile2Outputfile(t, newdata, sendfile, outputfile)

	newdata = append(newdata, []byte(" MOre on second line")...)
	assertSendfile2Outputfile(t, newdata, sendfile, outputfile)

	//Remove lines
	newdata = append(senddata, []byte("\nRemoved lines")...)
	assertSendfile2Outputfile(t, newdata, sendfile, outputfile)

	//Remove couple characters in middle
	newdata = append(newdata[0:len(newdata)-10], newdata[len(newdata)-5:]...)
	assertSendfile2Outputfile(t, newdata, sendfile, outputfile)
}

type SuiteSendDir struct {
	suite.Suite

	SendDir   string
	OutputDir string

	//Do not initialize this
	SenderReceiverConnector
}

func (suite *SuiteSendDir) SetupTest() {
	t := suite.T()
	outputDir := suite.OutputDir
	if err := suite.SetupConnection(outputDir); err != nil {
		t.Fatal(err)
	}
}

func (suite *SuiteSendDir) checkFileContent(senddata []byte, filename string) bool {
	t := suite.T()
	_, err := os.Stat(filename)
	if err == os.ErrNotExist {
		t.Log("File does not exists")
		return false
	}

	if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOENT {
		t.Log("File does not exists")
		return false
	}

	require.Nil(t, err)
	data, err := ioutil.ReadFile(filename)
	require.Nil(t, err)

	if string(data) != string(senddata) {
		t.Log("Data in files is not same")
		return false
	}

	return true
}

func (suite *SuiteSendDir) TestReceiveFile() {
	t := suite.T()
	sen := suite.sender
	senddir := suite.SendDir
	outputDir := suite.OutputDir
	//Send our file
	t.Log("Starting sending file", senddir)
	err := sen.SendFile(senddir)
	require.Nil(t, err)

	//Compare data received
	var sendfiles []string

	filepath.Walk(senddir, func(path string, info os.FileInfo, err error) error {
		if path != "" {
			sendfiles = append(sendfiles, path)
		}
		return nil
	})

	assert.Eventually(t, func() bool {
		for _, sendfile := range sendfiles {
			outf := filepath.Base(senddir) + strings.TrimPrefix(sendfile, senddir)
			outputFile := fmt.Sprintf("%s/%s", outputDir, outf)

			t.Logf("Comparing content %s vs %s", sendfile, outputFile)

			if !osx.CompareFilesAreSame(sendfile, outputFile) {
				return false
			}
		}
		return true
	}, 5*time.Second, 1*time.Second, "File is not received")
}

type SuiteStreamDir struct {
	suite.Suite

	SendDir   string
	OutputDir string

	//Do not initialize this
	SenderReceiverConnector
}

func (suite *SuiteStreamDir) SetupTest() {
	t := suite.T()
	outputDir := suite.OutputDir
	if err := suite.SetupConnection(outputDir); err != nil {
		t.Fatal(err)
	}

	//Clear also our stream dir
	senddir := suite.SendDir
	osx.RemoveContents(senddir)

	//Create some existing files
	sendfile := filepath.Join(senddir, "testfile.txt")
	err := ioutil.WriteFile(sendfile, []byte("Here some lines\nThis should be second line\nThirdline"), 0644)
	require.Nil(t, err)

}

func (suite *SuiteStreamDir) TestSendingDirAndChanges() {
	t := suite.T()
	sen := suite.sender
	senddir := suite.SendDir
	outputDir := suite.OutputDir
	outputsenddir := filepath.Join(outputDir, filepath.Base(senddir))
	// outputFile := suite.outputFile

	//Send our file
	t.Log("Starting sending file", senddir)
	// err := sen.SendFile(senddir)
	// require.Nil(t, err)

	fi, err := os.Stat(senddir)
	require.Nil(t, err)

	w := watcher.New(senddir, fi)
	ctx := context.Background()

	sender := sen.NewFileStreamer(senddir, fi)
	sender.Stream(context.Background())

	go w.ListenChangeFile(ctx, func(fin os.FileInfo, path string) error {
		return sender.SubStream(fin, path)
	})

	//Compare data received
	sendfile := filepath.Join(senddir, "testfile.txt")
	outputfile := filepath.Join(outputsenddir, "testfile.txt")

	err = os.Truncate(sendfile, 0)
	require.Nil(t, err)

	suite.assertSimpleFileChanges(sendfile, outputfile)

	//Add new file
	sendfile = filepath.Join(senddir, "newfile.txt")
	outputfile = filepath.Join(outputsenddir, "newfile.txt")
	assertSendfile2Outputfile(t, []byte("Some data"), sendfile, outputfile)

	//Add new dir
	sendfirstdir := filepath.Join(senddir, "firstdir")
	outputfile = filepath.Join(outputsenddir, "firstdir")
	assertSendDir2OutputDir(t, sendfirstdir, outputfile)

	//Add new file in firstdir
	sendfile = filepath.Join(sendfirstdir, "newfile.txt")
	outputfile = filepath.Join(outputsenddir, filepath.Base(sendfirstdir), "newfile.txt")
	assertSendfile2Outputfile(t, []byte("Some data\nMoreaaaaaaaa sssss"), sendfile, outputfile)

	//Add subdir in  firstdir
	sendfirstsubdir := filepath.Join(sendfirstdir, "subdir")
	outputfile = filepath.Join(outputsenddir, filepath.Base(sendfirstdir), "subdir")
	assertSendDir2OutputDir(t, sendfirstsubdir, outputfile)

	//Add new file in firstdir/subdir
	sendfile = filepath.Join(sendfirstsubdir, "newfile.txt")
	outputfile = filepath.Join(outputsenddir, filepath.Base(sendfirstdir), filepath.Base(sendfirstsubdir), "newfile.txt")
	assertSendfile2Outputfile(t, []byte("Some data\nMoreaaaaaaaa sssss"), sendfile, outputfile)
}

func (suite *SuiteStreamDir) assertSimpleFileChanges(sendfile string, outputfile string) {
	t := suite.T()
	senddata, err := ioutil.ReadFile(sendfile)
	require.Nil(t, err)
	assertSendfile2Outputfile(t, senddata, sendfile, outputfile)

	//Add line
	newdata := append(senddata, []byte("\nSomething on new line")...)
	assertSendfile2Outputfile(t, newdata, sendfile, outputfile)

	//Remove lines
	newdata = append(senddata, []byte("\nRemoved lines")...)
	assertSendfile2Outputfile(t, newdata, sendfile, outputfile)
}

func TestSendFile(t *testing.T) {
	suite.Run(t, &SuiteSendFile{
		Sendfile:  "./internal/send/testfile.txt",
		OutputDir: "./internal/received",
	})
}

func TestStreamFile(t *testing.T) {
	suite.Run(t, &SuiteStreamFile{
		Sendfile:  "./internal/send/testfile.txt",
		OutputDir: "./internal/received",
	})
}

func TestSendDir(t *testing.T) {
	suite.Run(t, &SuiteSendDir{
		SendDir:   "internal/senddir",
		OutputDir: "internal/received",
	})
}

func TestStreamDir(t *testing.T) {
	suite.Run(t, &SuiteStreamDir{
		SendDir:   "internal/streamdir",
		OutputDir: "internal/received",
	})
}
