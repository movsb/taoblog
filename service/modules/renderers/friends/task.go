package friends

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"slices"
	"time"

	"github.com/movsb/taoblog/service/modules/cache"
)

type _CacheKey struct {
	PostID     int
	FaviconURL string
}

type _CacheValue struct {
	ContentType string
	Content     []byte
}

type Task struct {
	cache      *cache.FileCache
	invalidate func(postID int)
}

func NewTask(cache *cache.FileCache, invalidate func(postID int)) *Task {
	t := &Task{
		cache:      cache,
		invalidate: invalidate,
	}
	go t.refreshLoop(context.Background())
	return t
}

const ttl = time.Hour * 24 * 7

func (t *Task) Get(postID int, faviconURL string) (string, []byte, bool) {
	var value _CacheValue
	if err := t.cache.GetOrLoad(
		_CacheKey{postID, faviconURL}, ttl, &value,
		func() (any, error) {
			go t.update(postID, faviconURL)
			return nil, fmt.Errorf(`async`)
		},
	); err != nil {
		return ``, nil, false
	}
	return value.ContentType, value.Content, true
}

func (t *Task) KeepInUse(postID int, urls []string) {
	var keys []_CacheKey
	t.cache.GetAllKeysFor(&keys)
	for _, k := range keys {
		if k.PostID != postID {
			continue
		}
		if !slices.Contains(urls, k.FaviconURL) {
			t.cache.Delete(k)
			log.Println(`删除不存在的头像：`, k)
		}
	}
}

const refreshTTL = time.Hour * 24

func (t *Task) refreshLoop(ctx context.Context) {
	refresh := func() {
		log.Println(`即将更新朋友头像`)

		var keys []_CacheKey
		t.cache.GetAllKeysFor(&keys)

		for _, k := range keys {
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
	var (
		contentType string
		content     []byte
		err         error
	)
	contentType, content, err = t.get(faviconURL)
	if err != nil {
		log.Println(faviconURL, err)
		return
	}
	t.cache.Set(
		_CacheKey{postID, faviconURL},
		_CacheValue{
			ContentType: contentType,
			Content:     content,
		}, ttl,
	)
	t.invalidate(postID)
	log.Println(`已更新朋友头像数据：`, faviconURL)
}

const maxBodySize = 200 << 10

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
	body, err := io.ReadAll(io.LimitReader(rsp.Body, maxBodySize))
	if err != nil {
		log.Println(`读取头像 body 时出错：`, err)
		return ``, nil, err
	}
	contentType, _, _ := mime.ParseMediaType(rsp.Header.Get(`Content-Type`))
	if contentType == "" {
		contentType = http.DetectContentType(body)
	}
	if contentType == "" {
		return ``, nil, fmt.Errorf(`无法识别的内容类型`)
	}
	return contentType, body, nil
}
