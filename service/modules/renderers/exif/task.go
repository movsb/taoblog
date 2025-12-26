package exif

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/url"
	"path"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/cache"
	"github.com/movsb/taoblog/service/modules/renderers/exif/exif_exports"
)

// Name 只包含最后一部分，与时间一起，足够区分了吧？
type _CacheKey struct {
	Name string
	Time int64
}

type _CacheValue struct {
	Metadata exif_exports.Metadata
}

type InvalidateCacheFor func(id int)

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
// url 仅用来获取文件名并用作缓存键。
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

	time := stat.ModTime().Unix()
	if file, ok := stat.Sys().(*models.File); ok {
		time = file.UpdatedAt
	}

	key := _CacheKey{Name: baseName, Time: time}
	value := _CacheValue{}
	if err := t.cache.GetOrLoad(
		key, ttl, &value,
		func() (any, error) {
			shouldCloseFile = false
			// 为了快速返回，放线程中执行。
			go utils.LimitExec(`exif`, &t.numberOfExecutions, runtime.NumCPU(), func() {
				t.extract(id, baseName, stat, key, f)
			})
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

	md, err := exif_exports.Extract(r)
	if err != nil {
		log.Println(`exif.task.extract`, err, id, name, stat.Name())
		return
	}
	md.FileName = name
	md.FileSize = utils.ByteCountIEC(stat.Size())

	t.cache.Set(key, _CacheValue{Metadata: *md}, ttl)
	log.Println(`更新图片元数据：`, key)

	t.invalidate(id)
}
