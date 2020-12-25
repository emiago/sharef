package watcher

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
)

type Watcher struct {
	streamPath string
	streamInfo os.FileInfo
	log        logrus.FieldLogger
}

type OnChangeFile func(fi os.FileInfo, path string) error

func New(fullpath string, fi os.FileInfo) *Watcher {
	return &Watcher{
		streamPath: fullpath,
		streamInfo: fi,
		log:        logrus.WithField("prefix", "watcher"),
	}
}

func (s *Watcher) ListenChangeFile(ctx context.Context, f OnChangeFile) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		s.log.Println("ERROR", err)
		return
	}
	defer watcher.Close()

	if err := watcher.Add(s.streamPath); err != nil {
		s.log.Info("ERROR ", err)
	}

	for {
		select {
		// watch for events
		case event := <-watcher.Events:
			s.log.Infof("EVENT! %s\n", event.String())
			s.checkFileChanges(event, watcher, f)

			// watch for errors
		case err := <-watcher.Errors:
			s.log.Info("ERROR ", err)

		case <-ctx.Done():
			return
		}
	}
}

func (s *Watcher) checkFileChanges(event fsnotify.Event, watcher *fsnotify.Watcher, f OnChangeFile) {
	// if s.bytesWritten < s.streamInfo.Size {
	// 	return
	// }

	switch {
	case event.Op&fsnotify.Write == fsnotify.Write:
	case event.Op&fsnotify.Create == fsnotify.Create:
		if !s.streamInfo.IsDir() { //Only streaming dir we follow create changes
			return
		}
	default:
		return
	}

	path := event.Name

	fi, err := os.Stat(path)
	if err != nil {
		s.log.Error(err)
		return
	}

	s.log.WithField("path", path).Info("Sending file changes")

	if err := f(fi, path); err != nil {
		s.log.Error(err)
		return
	}

	if fi.IsDir() {
		//Add tracking changes for this dir
		if err := watcher.Add(path); err != nil {
			s.log.Error(err)
		}
	}
}
