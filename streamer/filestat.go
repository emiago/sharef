package streamer

import (
	"os"
	"time"
)

type FileStat struct {
	name    string
	size    int64
	mode    os.FileMode
	modtime time.Time
}

func NewFileStat(name string, size int64, mode os.FileMode, modtime time.Time) *FileStat {
	return &FileStat{name, size, mode, modtime}
}

func (s *FileStat) Name() string {
	return s.name
}

func (s *FileStat) Size() int64 {
	return s.size
}

func (s *FileStat) Mode() os.FileMode {
	return s.mode
}

func (s *FileStat) ModTime() time.Time {
	return s.modtime
}

func (s *FileStat) IsDir() bool {
	return s.mode.IsDir()
}

func (s *FileStat) Sys() interface{} {
	return nil
}
