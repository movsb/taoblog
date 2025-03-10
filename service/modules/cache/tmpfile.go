package cache

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type TmpFiles struct {
	dir  string
	keep time.Duration
	lock sync.Mutex
}

// TODO 支持内存文件系统以方便测试。
func NewTmpFiles(dir string, keep time.Duration) *TmpFiles {
	if keep < time.Minute {
		panic(`cache time too small`)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	t := &TmpFiles{
		dir:  dir,
		keep: keep,
	}

	go func() {
		for {
			time.Sleep(keep / 6)
			t.clean()
		}
	}()

	return t
}

func (t *TmpFiles) nameFor(key string) string {
	hash := fmt.Sprintf(`%x`, md5.Sum([]byte(key)))
	return filepath.Join(t.dir, fmt.Sprintf(`taoblog-cache-%s`, hash))
}

// NOTE loader 是串行调用的。
func (t *TmpFiles) GetOrLoad(key string, loader func(key string) (io.ReadCloser, error)) (io.ReadCloser, error) {
	t.lock.Lock()
	fp, err := os.Open(t.nameFor(key))
	if err == nil {
		t.lock.Unlock()
		return fp, nil
	}
	r, err := loader(key)
	if err != nil {
		t.lock.Unlock()
		return nil, err
	}
	if err := t.cache(key, r); err != nil {
		t.lock.Unlock()
		// 失败直接用。
		return loader(key)
	}
	t.lock.Unlock()
	return os.Open(t.nameFor(key))
}

func (t *TmpFiles) cache(key string, r io.ReadCloser) error {
	fp, err := os.Create(t.nameFor(key))
	if err != nil {
		return nil
	}
	if _, err := io.Copy(fp, r); err != nil {
		fp.Close()
		os.Remove(fp.Name())
		return err
	}
	if err := fp.Close(); err != nil {
		return err
	}
	r.Close()
	now := time.Now()
	os.Chtimes(fp.Name(), now, now)
	log.Println(`New file cache:`, key, fp.Name())
	return nil
}

func (t *TmpFiles) clean() {
	entries, err := os.ReadDir(t.dir)
	if err != nil {
		log.Println(err)
		return
	}
	for _, entry := range entries {
		if info, err := entry.Info(); err == nil {
			if time.Since(info.ModTime()) > t.keep {
				if err := os.Remove(entry.Name()); err != nil {
					log.Println(err)
				} else {
					log.Println(`Removed tmp file cache:`, entry.Name())
				}
			}
		}
	}
}
