// +build integration

package itests

import (
	"fmt"
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SuiteApiSender struct {
	suite.Suite

	Sendfile  string
	OutputDir string

	//Do not initialize this
	SenderReceiverConnector
	outputFile string
}

func (suite *SuiteApiSender) SetupTest() {
	t := suite.T()
	outputDir := suite.OutputDir
	if err := suite.SetupConnection(outputDir); err != nil {
		t.Fatal(err)
	}
	suite.outputFile = fmt.Sprintf("%s/%s", outputDir, path.Base(suite.Sendfile))
}

func (suite *SuiteApiSender) TestReceiveFile() {
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

func TestApiSenderSendFile(t *testing.T) {
	suite.Run(t, &SuiteStreamFile{
		Sendfile:  "./internal/send/testfile.txt",
		OutputDir: "./internal/received",
	})
}
