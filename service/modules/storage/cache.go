package storage

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sync"
	"time"
)

type FileCache struct {
	root *os.Root

	lock       sync.Mutex
	lastAccess map[string]time.Time
}

func NewFileCache(root *os.Root) *FileCache {
	fc := &FileCache{
		root:       root,
		lastAccess: map[string]time.Time{},
	}
	go fc.clean(context.Background())
	return fc
}

func (fc *FileCache) pathOf(pid int, digest string) string {
	return fmt.Sprintf(`%s%08x%02x`, digest, pid, 1)
}

func (fc *FileCache) Open(pid int, digest string) (fs.File, error) {
	path := fc.pathOf(pid, digest)
	file, err := fc.root.Open(path)
	if err == nil {
		fc.lock.Lock()
		defer fc.lock.Unlock()
		fc.lastAccess[path] = time.Now()
	}
	return file, err
}

func (fc *FileCache) Create(pid int, digest string, data []byte) error {
	path := fc.pathOf(pid, digest)
	fp, err := fc.root.Create(path)
	if err != nil {
		return err
	}

	if _, err := fp.Write(data); err != nil {
		fp.Close()
		return err
	}

	if err := fp.Close(); err != nil {
		return err
	}

	log.Println(`创建临时文件：`, path)

	return nil
}

func (fc *FileCache) clean(ctx context.Context) {
	ticker := time.NewTicker(time.Hour * 24)
	defer ticker.Stop()

	clean := func() {
		const maxAccess = time.Hour * 24

		entries, err := fs.ReadDir(fc.root.FS(), `.`)
		if err != nil {
			log.Println(err)
			return
		}

		// 集中在一起，不加锁删除。
		removes := []string{}

		fc.lock.Lock()

		for _, entry := range entries {
			if !entry.Type().IsRegular() {
				log.Println(`奇怪的文件：`, entry.Name())
				continue
			}

			// 防启动的时候是空的，会清除所有文件。
			if _, ok := fc.lastAccess[entry.Name()]; !ok {
				fc.lastAccess[entry.Name()] = time.Now()
			}

			if t, ok := fc.lastAccess[entry.Name()]; ok && time.Since(t) > maxAccess {
				removes = append(removes, entry.Name())
			}
		}

		fc.lock.Unlock()

		// 删除文件可能会很慢，不要加锁。
		for _, name := range removes {
			err := fc.root.Remove(name)
			log.Println(`清理缓存文件：`, name, err)
		}

		fc.lock.Lock()
		defer fc.lock.Unlock()

		for _, name := range removes {
			delete(fc.lastAccess, name)
		}
	}

	clean()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			clean()
		}
	}
}
