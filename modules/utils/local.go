package utils

import (
	"io/fs"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

type FsWithChangeNotify interface {
	fs.FS
	Changed() <-chan fsnotify.Event
}

type Local struct {
	root string
	fs.FS
	ch chan fsnotify.Event
}

var _ FsWithChangeNotify = (*Local)(nil)

func NewLocal(root string) fs.FS {
	l := &Local{
		root: root,
		FS:   os.DirFS(root),
	}
	l.ch = l.watch()
	return l
}

func (l *Local) Changed() <-chan fsnotify.Event {
	return l.ch
}

func (l *Local) watch() chan fsnotify.Event {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
		return nil
	}
	// defer watcher.Close()

	ch := make(chan fsnotify.Event)

	go func() {
		for {
			select {
			case err := <-watcher.Errors:
				log.Println(err)
				return
			case event := <-watcher.Events:
				// log.Println(event)
				ch <- event
			}
		}
	}()

	if err := watcher.Add(l.root); err != nil {
		panic(err)
	} else {
		log.Println(`Started watching`, l.root)
	}

	return ch
}
