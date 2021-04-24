package streamer

import (
	"os"
	"path/filepath"
	"time"
)

// type StreamFile struct {
// 	Name string //Relative path depending on what is sharing
// 	Path string //This should be removed, it is used on sender receiver to keep original path

// 	Size    int64
// 	Mode    os.FileMode
// 	ModTime time.Time
// 	// Data    []byte
// 	// sys     syscall.Stat_t
// 	//Used on receiver side
// 	fullPath string
// }

func (s *StreamFile) IsDir() bool {
	return s.FileMode().IsDir()
}

func (s *StreamFile) FileMode() os.FileMode {
	fmode := os.FileMode(s.Mode)
	return fmode
}

func StreamFile2FileInfo(fi StreamFile) os.FileInfo {
	t, _ := time.Parse(time.RFC3339Nano, fi.ModTime)
	return &FileStat{
		name:    fi.Name,
		size:    fi.Size_,
		mode:    fi.FileMode(),
		modtime: t,
	}
}

func FileInfo2StreamFile(fi os.FileInfo, path string) StreamFile {
	return StreamFile{
		Name:    fi.Name(),
		Path:    filepath.Clean(path),
		Size_:   fi.Size(),
		Mode:    uint32(fi.Mode()),
		ModTime: fi.ModTime().String(),
	}
}
