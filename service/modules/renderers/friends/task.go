package friends

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/phuslu/lru"
)

type CacheKey struct {
	PostID     int
	FaviconURL string
}

func (k CacheKey) String() string {
	return string(utils.Must1(json.Marshal(k)))
}

func CacheKeyFromString(s string) CacheKey {
	var k CacheKey
	json.Unmarshal([]byte(s), &k)
	return k
}

type CacheValue struct {
	ContentType string
	Content     []byte
}

type Task struct {
	cache      *lru.TTLCache[CacheKey, CacheValue]
	lock       sync.Mutex
	keys       []CacheKey
	store      utils.PluginStorage
	deb        *utils.Debouncer
	invalidate func(postID int)
}

func NewTask(storage utils.PluginStorage, invalidate func(postID int)) *Task {
	t := &Task{
		cache:      lru.NewTTLCache[CacheKey, CacheValue](1024),
		store:      storage,
		invalidate: invalidate,
	}
	t.deb = utils.NewDebouncer(time.Second*10, t.save)
	t.load()
	go t.refreshLoop(context.Background())
	return t
}

const ttl = time.Hour * 24 * 7

func (t *Task) load() {
	cached, err := t.store.Get(`cache`)
	if err != nil {
		log.Println(err)
		return
	}

	m := map[string]CacheValue{}
	if err := json.Unmarshal([]byte(cached), &m); err != nil {
		log.Println(err)
		return
	}

	for k, v := range m {
		ck := CacheKeyFromString(k)
		t.cache.Set(ck, v, ttl)
		t.keys = append(t.keys, ck)
	}

	log.Println(`已恢复朋友头像数据`)
}

func (t *Task) save() {
	t.lock.Lock()
	defer t.lock.Unlock()

	m := map[string]CacheValue{}
	existingKeys := []CacheKey{}
	for _, k := range t.keys {
		if value, ok := t.cache.Get(k); ok {
			m[k.String()] = value
			existingKeys = append(existingKeys, k)
		}
	}

	data := string(utils.Must1(json.Marshal(m)))
	t.store.Set(`cache`, data)
	t.keys = existingKeys

	log.Println(`已存储朋友头像数据`)
}

func (t *Task) Get(postID int, faviconURL string) (string, []byte, bool) {
	if value, found := t.cache.Get(CacheKey{postID, faviconURL}); found {
		return value.ContentType, value.Content, true
	}

	go t.update(postID, faviconURL)

	return ``, nil, false
}

const refreshTTL = time.Hour * 6

func (t *Task) refreshLoop(ctx context.Context) {
	refresh := func() {
		log.Println(`即将更新朋友头像`)
		t.lock.Lock()
		defer t.lock.Unlock()

		for _, k := range t.keys {
			go t.update(k.PostID, k.FaviconURL)
		}
	}

	// refresh()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(refreshTTL):
			refresh()
		}
	}
}

func (t *Task) update(postID int, faviconURL string) {
	contentType, content, err := t.get(faviconURL)
	if err != nil {
		log.Println(faviconURL, err)
		return
	}
	t.cache.Set(CacheKey{postID, faviconURL}, CacheValue{
		ContentType: contentType,
		Content:     content,
	}, ttl)
	t.lock.Lock()
	t.keys = append(t.keys, CacheKey{postID, faviconURL})
	t.lock.Unlock()
	t.invalidate(postID)
	t.deb.Enter()
	log.Println(`已更新朋友头像数据：`, faviconURL)
}

const maxBodySize = 100 << 10

// 返回 [ContentType, Data]
func (t *Task) get(faviconURL string) (string, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, faviconURL, nil)
	if err != nil {
		log.Println(`头像请求失败：`, err)
		return ``, nil, err
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(`头像请求失败：`, err)
		return ``, nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		log.Println(`头像请求失败：`, rsp.Status)
		return ``, nil, fmt.Errorf(`StatusCode: %d`, rsp.StatusCode)
	}
	body, _ := io.ReadAll(io.LimitReader(rsp.Body, maxBodySize))
	contentType, _, _ := mime.ParseMediaType(rsp.Header.Get(`Content-Type`))
	if contentType == "" {
		contentType = http.DetectContentType(body)
	}
	if contentType == "" {
		return ``, nil, fmt.Errorf(`无法识别的内容类型`)
	}
	return contentType, body, nil
}
