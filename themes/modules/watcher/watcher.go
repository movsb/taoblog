package watcher

import (
	"io/ioutil"
	"log"
	"os"
	"sort"
	"time"
)

type FolderChangedWatcher struct {
	root     string
	exts     []string
	contents []os.FileInfo
}

type FolderChangedEvent struct {
}

func NewFolderChangedWatcher(root string, exts []string) *FolderChangedWatcher {
	return &FolderChangedWatcher{
		root: root,
		exts: exts,
	}
}

func (f *FolderChangedWatcher) Watch() <-chan *FolderChangedEvent {
	f.contents = f.snapshot()

	ch := make(chan *FolderChangedEvent)

	go func() {
		ticker := time.NewTicker(time.Second * 3)
		defer ticker.Stop()

		for range ticker.C {
			contents := f.snapshot()
			if f.changed(f.contents, contents) {
				ch <- &FolderChangedEvent{}
			}
			f.contents = contents
		}
	}()

	return ch
}

func (f *FolderChangedWatcher) changed(old, new []os.FileInfo) bool {
	if len(old) != len(new) {
		return true
	}
	for i := 0; i < len(old); i++ {
		l, r := old[i], new[i]
		if l.Name() != r.Name() {
			return true
		}
		if !l.ModTime().Equal(r.ModTime()) {
			return true
		}
		if l.Size() != r.Size() {
			return true
		}
	}
	return false
}

func (f *FolderChangedWatcher) snapshot() []os.FileInfo {
	entries, err := ioutil.ReadDir(f.root)
	if err != nil {
		log.Println(err)
		return nil
	}
	infos := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		if !entry.Mode().IsRegular() {
			continue
		}
		infos = append(infos, entry)
	}
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name() < infos[j].Name()
	})
	return infos
}
