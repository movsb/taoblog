package avatar

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/movsb/taoblog/gateway/handlers/avatar/github"
	"github.com/movsb/taoblog/gateway/handlers/avatar/gravatar"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/phuslu/lru"
)

type CacheKey struct {
	Email string // 小写的邮箱
}

func (k CacheKey) String() string {
	return k.Email
}

func CacheKeyFromString(s string) CacheKey {
	var k CacheKey
	k.Email = s
	return k
}

type CacheValue struct {
	Content      []byte
	LastModified time.Time
}

type Task struct {
	cache *lru.TTLCache[CacheKey, CacheValue]
	lock  sync.Mutex
	keys  []CacheKey
	store utils.PluginStorage
	deb   *utils.Debouncer
}

func NewTask(storage utils.PluginStorage) *Task {
	t := &Task{
		cache: lru.NewTTLCache[CacheKey, CacheValue](1024),
		store: storage,
	}
	t.deb = utils.NewDebouncer(time.Second*10, t.save)
	t.load()
	go t.refreshLoop(context.Background())
	return t
}

const ttl = time.Hour * 24 * 7

func (t *Task) load() {
	cached, err := t.store.GetString(`cache`)
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

	log.Println(`已恢复头像数据`)
}

func (t *Task) save() {
	t.lock.Lock()
	defer t.lock.Unlock()

	m := map[string]CacheValue{}
	existingKeys := []CacheKey{}
	for _, k := range t.keys {
		if value, _, ok := t.cache.Peek(k); ok {
			m[k.String()] = value
			existingKeys = append(existingKeys, k)
		}
	}

	data := string(utils.Must1(json.Marshal(m)))
	t.store.SetString(`cache`, data)
	t.keys = existingKeys

	log.Println(`已存储头像数据`)
}

func (t *Task) Get(email string) (lastModified time.Time, content []byte, found bool) {
	ck := CacheKey{Email: strings.ToLower(email)}

	value, err, found := t.cache.GetOrLoad(
		context.Background(), ck,
		func(ctx context.Context, ck CacheKey) (CacheValue, time.Duration, error) {
			l, c, err := get(email)
			if err != nil {
				log.Println(err, email)
				return CacheValue{}, ttl, err
			}
			return CacheValue{
				LastModified: l,
				Content:      c,
			}, ttl, nil
		},
	)

	if err != nil {
		return time.Time{}, nil, false
	}

	if !found {
		t.lock.Lock()
		t.keys = append(t.keys, ck)
		t.lock.Unlock()
		t.deb.Enter()
	}

	return value.LastModified, value.Content, true
}

const refreshTTL = time.Hour * 24

// TODO 某些 email 可能因为删除评论而不复存在，但是由于一直在刷新，会导致缓存一直存在，需要清理。
func (t *Task) refreshLoop(ctx context.Context) {
	refresh := func() {
		log.Println(`即将更新头像`)

		t.lock.Lock()
		keys := append([]CacheKey{}, t.keys...)
		t.lock.Unlock()

		for _, k := range keys {
			// 顺序更新，没必要异步？
			func() {
				l, c, err := get(k.Email)
				if err != nil {
					log.Println(err)
					return
				}
				t.cache.Set(CacheKey{Email: k.Email}, CacheValue{
					Content:      c,
					LastModified: l,
				}, ttl)
				t.deb.Enter()
			}()
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(refreshTTL):
			refresh()
		}
	}
}

const maxBodySize = 50 << 10

func get(email string) (_ time.Time, _ []byte, outErr error) {
	rsp, err := github.Get(context.Background(), email)
	if err != nil {
		rsp, err = gravatar.Get(context.Background(), email)
	}
	if err != nil {
		log.Println(`头像获取失败：`, err)
		outErr = err
		return
	}

	lastModified, _ := time.Parse(http.TimeFormat, rsp.Header.Get(`Last-Modified`))

	rc := http.MaxBytesReader(nil, rsp.Body, maxBodySize)
	defer rc.Close()

	body, err := io.ReadAll(rc)
	if err != nil {
		outErr = err
		return
	}

	return lastModified, body, nil
}
