package fsx

import (
	"io"
	"io/ioutil"
	"os"
)

type FileReader struct {
}

func NewFileReader() *FileReader {
	return &FileReader{}
}

func (s *FileReader) OpenFile(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	return file, err
}

func (s *FileReader) ReadDir(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(path)
}

type FileWriter struct {
}

func NewFileWriter() *FileWriter {
	return &FileWriter{}
}

func (s *FileWriter) OpenFile(path string, mode os.FileMode) (io.WriteCloser, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	return file, err
}

func (s *FileWriter) Mkdir(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}
