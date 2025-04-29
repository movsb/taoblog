package rss

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/service/modules/dynamic"
	rss_parser "github.com/movsb/taoblog/service/modules/renderers/rss/parser"
)

type Task struct {
	store utils.PluginStorage

	lock sync.Mutex

	posts map[int]map[string][]*PostData

	saveDebouncer *utils.Debouncer
	invalidate    func(pid int)
	refreshNow    chan struct{}
}

func NewTask(ctx context.Context, store utils.PluginStorage, invalidate func(pid int)) *Task {
	t := &Task{
		store:      store,
		posts:      map[int]map[string][]*PostData{},
		invalidate: invalidate,
		refreshNow: make(chan struct{}, 10),
	}
	t.saveDebouncer = utils.NewDebouncer(time.Second*10, t.save)
	go t.load()
	go t.refresh(ctx)
	dynamic.WithHandler(module, NewHandler(t))
	return t
}

type PostData struct {
	SiteName string
	PostName string
	PostURL  string
	PubDate  time.Time
	ReadAt   time.Time
}

func (d *PostData) OpenURL() string {
	u := url.URL{Path: dynamic.URL(`/rss/open`)}
	q := u.Query()
	q.Set(`url`, d.PostURL)
	u.RawQuery = q.Encode()
	return u.String()
}

func (t *Task) load() {
	t.lock.Lock()
	defer t.lock.Unlock()

	var posts map[int]map[string][]*PostData
	str, err := t.store.GetStringDefault(`posts`, `{}`)
	if err != nil {
		log.Println(`rss.load error`, err)
		return
	}
	if err := json.Unmarshal([]byte(str), &posts); err != nil {
		log.Println(`rss.load error`, err)
		return
	}
	t.posts = posts
}

func (t *Task) save() {
	t.lock.Lock()
	defer t.lock.Unlock()
	b := utils.Must1(json.Marshal(t.posts))
	t.store.SetString(`posts`, string(b))
}

func (t *Task) GetLatestPosts(postID int, urls []string) []*PostData {
	t.lock.Lock()
	defer t.lock.Unlock()

	// 删除已经不存在的网站。
	// 但不会马上删除已经阅读的文章链接，防止加回来后被标记为未读。
	sites := t.posts[postID]
	toDelete := []string{}
	for site := range sites {
		if !slices.Contains(urls, site) {
			toDelete = append(toDelete, site)
		}
	}
	for _, d := range toDelete {
		delete(sites, d)
	}
	if len(toDelete) > 0 {
		t.posts[postID] = sites
		t.saveDebouncer.Enter()
	}

	// 加入新的
	toAdd := []string{}
	for _, url := range urls {
		if _, ok := sites[url]; !ok {
			toAdd = append(toAdd, url)
		}
	}
	if len(toAdd) > 0 {
		if sites == nil {
			sites = map[string][]*PostData{}
		}
		for _, a := range toAdd {
			sites[a] = []*PostData{}
		}
		t.posts[postID] = sites
		t.saveDebouncer.Enter()
		log.Println(`立即刷新`)
		t.refreshNow <- struct{}{}
	}

	// 按最新发表时间排序。
	tooOld := time.Now().Add(-time.Minute * 10)
	posts := []*PostData{}
	for _, site := range sites {
		for _, p := range site {
			// 只保留未读的文章。
			if !p.ReadAt.IsZero() && p.ReadAt.Before(tooOld) {
				continue
			}
			posts = append(posts, p)
		}
	}
	slices.SortFunc(posts, func(a, b *PostData) int {
		return -int(a.PubDate.Unix() - b.PubDate.Unix())
	})

	return posts
}

func (t *Task) refresh(ctx context.Context) {
	if version.DevMode() {
		log.Println(`开发环境不运行订阅服务`)
		return
	}
	// 防止第一次等太久。
	go func() {
		time.Sleep(time.Minute * 10)
		t.refreshNow <- struct{}{}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Hour):
			t.doRefresh(ctx)
		case <-t.refreshNow:
			t.doRefresh(ctx)
		}
	}
}

func (t *Task) doRefresh(ctx context.Context) {
	t.lock.Lock()
	defer t.lock.Unlock()

	for pid, urls := range t.posts {
		// 等的目的是按文章刷新缓存。
		wg := sync.WaitGroup{}
		for url, posts := range urls {
			wg.Add(1)
			go func(pid int, url string, posts []*PostData) {
				defer wg.Done()
				t.doRefreshAsync(ctx, pid, url, posts)
			}(pid, url, posts)
		}
		go func(pid int) {
			// ??? 在线程中等？
			// 前面加锁冲突所致。
			wg.Wait()
			t.invalidate(pid)
		}(pid)
	}
}

func (t *Task) doRefreshAsync(ctx context.Context, pid int, url string, old []*PostData) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	data, err := t.fetch(ctxTimeout, url)
	if err != nil {
		log.Println(url, err)
		return err
	}

	copied := slices.Clone(old)

	for _, post := range data.Posts {
		if !slices.ContainsFunc(copied, func(p *PostData) bool {
			// 仅通过 URL 判断，如果文章后续有过修改（更新过发表时间），
			// 其将不会被重新添加到这里来。
			return p.PostURL == post.URL
		}) {
			copied = append(copied, &PostData{
				SiteName: data.Name,
				PostName: post.Name,
				PostURL:  post.URL,
				PubDate:  post.Date,
			})
		}
	}

	// 只保留前 10 篇
	const maxPosts = 10
	toKeep := min(maxPosts, len(copied))
	copied = copied[:toKeep]

	t.lock.Lock()
	defer t.lock.Unlock()

	// 文章和订阅可能已不存在，需要判断
	if _, ok := t.posts[pid]; ok {
		if _, ok := t.posts[pid][url]; ok {
			t.posts[pid][url] = copied
		}
	}

	t.saveDebouncer.Enter()

	return nil
}

type PerSiteData struct {
	Name  string
	Posts []PerSitePostData
}
type PerSitePostData struct {
	Name string
	URL  string
	Date time.Time
}

func (t *Task) fetch(ctx context.Context, url string) (_ *PerSiteData, outErr error) {
	defer utils.CatchAsError(&outErr)
	req := utils.Must1(http.NewRequestWithContext(ctx, http.MethodGet, url, nil))
	// 有点不道德，但是有些网站限制了 go-http-client ……
	req.Header.Add(`User-Agent`, `Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:137.0) Gecko/20100101 Firefox/137.0`)
	rsp := utils.Must1(http.DefaultClient.Do(req))
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		panic(fmt.Sprintf(`rss: statusCode: %s, %s`, rsp.Status, url))
	}
	parsed := utils.Must1(rss_parser.Parse(io.LimitReader(rsp.Body, 10<<20)))
	switch typed := parsed.(type) {
	case *rss_parser.RSS:
		data := PerSiteData{
			Name: typed.Channel.Title.String(),
		}
		for _, item := range typed.Channel.Items {
			data.Posts = append(data.Posts, PerSitePostData{
				Name: item.Title.String(),
				URL:  item.Link.String(),
				Date: item.PubDate.Time,
			})
		}
		return &data, nil
	case *rss_parser.Feed:
		data := PerSiteData{
			Name: typed.Title.String(),
		}
		for _, item := range typed.Entries {
			data.Posts = append(data.Posts, PerSitePostData{
				Name: item.Title.String(),
				URL:  item.Link.Href,
				Date: item.Published.Time,
			})
		}
		return &data, nil
	}
	return nil, fmt.Errorf(`未知错误：%s`, url)
}
