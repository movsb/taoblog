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
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/phuslu/lru"
)

// Name 只包含最后一部分，与时间一起，足够区分了吧？
type _CacheKey struct {
	Name string
	Time time.Time
}

func (k _CacheKey) String() string {
	return fmt.Sprintf(`%s@%d`, k.Name, int32(k.Time.Unix()))
}

func parseCacheKeyFrom(s string) _CacheKey {
	k := _CacheKey{}
	if p := strings.SplitN(s, `@`, 2); len(p) == 2 {
		k.Name = p[0]
		k.Time = time.Unix(int64(utils.DropLast1(strconv.Atoi(p[1]))), 0)
	}
	return k
}

type InvalidateCacheFor func(id int)

const maxExecutions = 16

type Task struct {
	invalidate   InvalidateCacheFor
	cache        *lru.TTLCache[_CacheKey, string]
	saveDebounce *utils.Debouncer
	storage      utils.PluginStorage

	lock    sync.Mutex
	allKeys []_CacheKey

	// 太多了内存会爆。
	numberOfExecutions atomic.Int32
}

const ttl = time.Hour * 24 * 30

func NewTask(storage utils.PluginStorage, invalidate InvalidateCacheFor) *Task {
	t := &Task{
		storage:    storage,
		invalidate: invalidate,
		cache:      lru.NewTTLCache[_CacheKey, string](1024),
	}
	t.saveDebounce = utils.NewDebouncer(time.Second*10, t.save)
	t.load()
	return t
}

func (t *Task) save() {
	log.Println(`即将保存图片元数据`)

	t.lock.Lock()
	defer t.lock.Unlock()
	m := map[string]string{}
	for _, k := range t.allKeys {
		if value, ok := t.cache.Get(k); ok {
			m[k.String()] = value
		}
	}

	data := string(utils.Must1(json.Marshal(m)))
	t.storage.Set(`cache`, data)
}

func (t *Task) load() {
	cached, err := t.storage.Get(`cache`)
	if err != nil {
		log.Println(err)
		return
	}

	m := map[string]string{}
	if err := json.Unmarshal([]byte(cached), &m); err != nil {
		log.Println(err)
		return
	}
	for k, v := range m {
		p := parseCacheKeyFrom(k)
		t.cache.Set(p, v, ttl)
		t.allKeys = append(t.allKeys, p)
	}
	log.Println(`恢复了图片元数据：`, t.cache.Len())
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

	key := _CacheKey{
		Name: baseName,
		Time: stat.ModTime(),
	}

	if value, ok := t.cache.Get(key); ok {
		return value
	}

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

	return ""
}

func (t *Task) extract(id int, name string, stat fs.FileInfo, key _CacheKey, r io.ReadCloser) {
	defer r.Close()

	cmd := exec.CommandContext(context.TODO(), `exiftool`, `-G`, `-s`, `-json`, `-`)
	cmd.Stdin = r
	output, err := cmd.Output()
	if err != nil {
		log.Println(id, name, err)
		return
	}
	var md []*Metadata
	if err := json.Unmarshal(output, &md); err != nil {
		log.Println(id, name, err)
		return
	}
	if len(md) <= 0 {
		log.Println(id, name, `没有提取到元数据。`)
		return
	}

	md[0].FileName = name
	md[0].FileSize = utils.ByteCountIEC(stat.Size())

	s := string(utils.DropLast1(json.Marshal(md[0].String())))
	t.cache.Set(key, s, ttl)
	t.lock.Lock()
	// TODO 只追加，没清除
	t.allKeys = append(t.allKeys, key)
	t.lock.Unlock()
	t.invalidate(id)
	t.saveDebounce.Enter()
}
