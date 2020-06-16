package streamer

import (
	"os"
	"path/filepath"
	"time"
)

type StreamFile struct {
	Name     string //Relative path depending on what is sharing
	Path     string //This should be removed, it is used on sender receiver to keep original path
	FullPath string
	Size     int64
	Mode     os.FileMode
	ModTime  time.Time
	// Data    []byte
	// sys     syscall.Stat_t
}

func StreamFile2FileInfo(fi StreamFile) os.FileInfo {
	return &FileStat{
		name:    fi.Name,
		size:    fi.Size,
		mode:    fi.Mode,
		modtime: fi.ModTime,
	}
}

func FileInfo2StreamFile(fi os.FileInfo, path string) StreamFile {
	return StreamFile{
		Name:    fi.Name(),
		Path:    filepath.Clean(path),
		Size:    fi.Size(),
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
	}
}
