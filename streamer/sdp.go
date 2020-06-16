package streamer

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	"github.com/klauspost/compress/s2"
)

// StripSDP remove useless elements from an SDP
func StripSDP(originalSDP string) string {
	finalSDP := strings.Replace(originalSDP, "a=group:BUNDLE audio video data", "a=group:BUNDLE data", -1)
	tmp := strings.Split(finalSDP, "m=audio")
	beginningSdp := tmp[0]

	var endSdp string
	if len(tmp) > 1 {
		tmp = strings.Split(tmp[1], "a=end-of-candidates")
		endSdp = strings.Join(tmp[2:], "a=end-of-candidates")
	} else {
		endSdp = strings.Join(tmp[1:], "a=end-of-candidates")
	}

	finalSDP = beginningSdp + endSdp
	finalSDP = strings.Replace(finalSDP, "\r\n\r\n", "\r\n", -1)
	finalSDP = strings.Replace(finalSDP, "\n\n", "\n", -1)
	return finalSDP
}

// Encode encodes the input in base64
// It can optionally zip the input before encoding
func Encode(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	var gzbuff bytes.Buffer
	gz, err := gzip.NewWriterLevel(&gzbuff, gzip.BestCompression)
	if err != nil {
		return "", err
	}
	if _, err := gz.Write(b); err != nil {
		return "", err
	}
	if err := gz.Flush(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(gzbuff.Bytes())

	// var s2buff bytes.Buffer
	// S2EncodeStream(b, &s2buff)

	// log.Info("JSON ", len(b))
	// log.Info("Compressed ", len(gzbuff.Bytes()))
	// log.Info("base64 ", len(encoded))
	// log.Info("s2 ", len(s2buff.Bytes()))

	return encoded, nil
}

// Decode decodes the input from base64
// It can optionally unzip the input after decoding
func Decode(in string, obj interface{}) error {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return err
	}

	gz, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer gz.Close()
	s, err := ioutil.ReadAll(gz)
	if err != nil {
		return err
	}

	return json.Unmarshal(s, obj)
}

func S2EncodeStream(data []byte, dst io.Writer) error {
	// src := bytes.NewReader([]byte)
	// dst := bytes.NewWriter([]byte)

	enc := s2.NewWriter(dst)
	err := enc.EncodeBuffer(data)
	// _, err := io.Copy(enc, src)
	if err != nil {
		enc.Close()
		return err
	}
	// Blocks until compression is done.
	return enc.Close()
}
