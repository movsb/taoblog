package exif

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/url"
	"os/exec"
	"path"
	"sync/atomic"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/cache"
)

// Name 只包含最后一部分，与时间一起，足够区分了吧？
type _CacheKey struct {
	Name string
	Time int64
}

type _CacheValue struct {
	Metadata Metadata
}

type InvalidateCacheFor func(id int)

const maxExecutions = 16

type Task struct {
	invalidate InvalidateCacheFor
	cache      *cache.FileCache

	// 太多了内存会爆。
	numberOfExecutions atomic.Int32
}

const ttl = time.Hour * 24 * 30

func NewTask(cache *cache.FileCache, invalidate InvalidateCacheFor) *Task {
	t := &Task{
		invalidate: invalidate,
		cache:      cache,
	}
	return t
}

// 负责关闭文件。
func (t *Task) get(id int, u string, f fs.File) string {
	shouldCloseFile := true
	defer func() {
		if shouldCloseFile {
			f.Close()
		}
	}()

	parsed, err := url.Parse(u)
	if err != nil {
		log.Println(err)
		return ""
	}

	baseName := path.Base(parsed.Path)

	stat, err := f.Stat()
	if err != nil {
		log.Println(err)
		return ""
	}

	key := _CacheKey{Name: baseName, Time: stat.ModTime().Unix()}
	value := _CacheValue{}
	if err := t.cache.GetOrLoad(
		key, ttl, &value,
		func() (any, error) {
			shouldCloseFile = false
			go func() {
				for t.numberOfExecutions.Add(+1) > maxExecutions {
					t.numberOfExecutions.Add(-1)
					log.Println(`任务太多，等待中...`)
					time.Sleep(time.Second)
				}
				defer t.numberOfExecutions.Add(-1)
				t.extract(id, baseName, stat, key, f)
			}()
			return nil, fmt.Errorf(`async`)
		},
	); err != nil {
		return ""
	}

	str, _ := json.Marshal(value.Metadata.String())
	return string(str)
}

func (t *Task) extract(id int, name string, stat fs.FileInfo, key _CacheKey, r io.ReadCloser) {
	defer r.Close()

	md, err := extract(r)
	if err != nil {
		log.Println(`exif.task.extract`, err)
		return
	}
	md.FileName = name
	md.FileSize = utils.ByteCountIEC(stat.Size())

	t.cache.Set(key, _CacheValue{Metadata: *md}, ttl)
	log.Println(`更新图片元数据：`, key)

	t.invalidate(id)
}

func extract(r io.ReadCloser) (*Metadata, error) {
	cmd := exec.CommandContext(context.TODO(), `exiftool`, `-G`, `-s`, `-json`, `-`)
	cmd.Stdin = r
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var md []*Metadata
	if err := json.Unmarshal(output, &md); err != nil {
		return nil, err
	}
	if len(md) <= 0 {
		return nil, fmt.Errorf(`没有提取到元数据`)
	}
	return md[0], nil
}
