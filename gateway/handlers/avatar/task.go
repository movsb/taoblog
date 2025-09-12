package avatar

import (
	"context"
	_ "embed"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/gateway/handlers/avatar/github"
	"github.com/movsb/taoblog/gateway/handlers/avatar/gravatar"
	"github.com/movsb/taoblog/gateway/handlers/avatar/qq"
	"github.com/movsb/taoblog/service/modules/cache"
	runtime_config "github.com/movsb/taoblog/service/modules/runtime"
)

type CacheKey struct {
	Email string // 小写的邮箱
}

type CacheValue struct {
	Content      []byte
	LastModified time.Time
}

type RuntimeConfig struct {
	RefreshNow bool `yaml:"refresh_now"`

	refreshNow chan struct{}
	config.Saver
}

func (c *RuntimeConfig) AfterSet(paths config.Segments, obj any) {
	switch paths.At(0).Key {
	case `refresh_now`:
		c.refreshNow <- struct{}{}
	}
}

type Task struct {
	cache *cache.FileCache
	rc    *RuntimeConfig
}

func NewTask(ctx context.Context, cache *cache.FileCache) *Task {
	t := &Task{
		cache: cache,
		rc: &RuntimeConfig{
			refreshNow: make(chan struct{}),
		},
	}
	if r := runtime_config.FromContext(ctx); r != nil {
		r.Register(`avatar`, t.rc)
	}
	go t.refreshLoop(context.Background())
	return t
}

func (t *Task) Get(email string, force bool) (lastModified time.Time, content []byte, found bool) {
	ck := CacheKey{Email: strings.ToLower(email)}

	val := CacheValue{}

	if force {
		t.cache.Delete(ck)
	}

	if err := t.cache.GetOrLoad(ck, cacheTTL, &val,
		func() (any, error) {
			l, c, err := get(context.Background(), email)
			if err != nil {
				log.Println(err, email)
				return CacheValue{}, err
			}
			return CacheValue{
				LastModified: l,
				Content:      c,
			}, nil
		},
	); err != nil {
		return time.Time{}, nil, false
	}

	return val.LastModified, val.Content, true
}

const refreshTTL = time.Hour * 24
const cacheTTL = refreshTTL * 7

func (t *Task) refreshLoop(ctx context.Context) {
	refresh := func() {
		log.Println(`即将更新头像`)
		defer log.Println(`头像更新全部结束`)

		var keys []CacheKey
		t.cache.GetAllKeysFor(&keys)

		// 顺序更新，没必要异步？
		for _, k := range keys {
			expiringAt := time.Time{}
			val := CacheValue{}
			if err := t.cache.Peek(k, &val, &expiringAt); err != nil {
				log.Println(err)
				continue
			}

			// 某些 email 可能因为删除评论而不复存在，但是由于一直在刷新，会导致缓存一直存在，需要清理。
			// 这里的做法：只刷新最近有 Get 过的头像，其它的任由过过期自动删除。
			// 因为：每成功 Get/刷新 都会使过期时间延迟缓存时间。
			// 最近有 Get 过：过期剩余时间 > 刷新时间。
			if time.Until(expiringAt) < refreshTTL {
				log.Println(`即将过期或已经过期，不再刷新：`, k)
				continue
			}

			l, c, err := get(ctx, k.Email)
			if err != nil {
				log.Println(err, k.Email)
				return
			}
			if val.LastModified.Equal(l) {
				continue
			}
			t.cache.Set(k, CacheValue{
				Content:      c,
				LastModified: l,
			}, cacheTTL)
			log.Println(`保存头像：`, k)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(refreshTTL):
			refresh()
		case <-t.rc.refreshNow:
			refresh()
		}
	}
}

const maxBodySize = 100 << 10

func get(ctx context.Context, email string) (_ time.Time, _ []byte, outErr error) {
	rsp, err := get2(ctx, email)
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

func get2(ctx context.Context, email string) (*http.Response, error) {
	// 喜欢用 QQ 邮箱？那就先尝试 QQ 头像吧。
	rsp, err0 := qq.Get(ctx, email)
	if err0 == nil {
		return rsp, nil
	}
	rsp, err1 := github.Get(ctx, email)
	if err1 == nil {
		return rsp, nil
	}
	rsp, err2 := gravatar.Get(ctx, email)
	if err2 == nil {
		return rsp, nil
	}

	return nil, errors.Join(err0, err1, err2)
}
