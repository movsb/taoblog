package commentgeo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type CommentGeo struct {
	ctx context.Context
	mu  sync.RWMutex
	// 本身带锁的
	lru *lru.Cache[string, *Resp]
}

func NewCommentGeo(ctx context.Context) *CommentGeo {
	cache, err := lru.New[string, *Resp](1 << 10)
	if err != nil {
		panic(err)
	}
	geo := &CommentGeo{
		ctx: ctx,
		lru: cache,
	}
	go geo.clearBad()
	return geo
}

// http://ip-api.com/json/8.8.8.8
/*
{
  "status": "success",
  "country": "United States",
  "countryCode": "US",
  "region": "VA",
  "regionName": "Virginia",
  "city": "Ashburn",
  "zip": "20149",
  "lat": 39.03,
  "lon": -77.5,
  "timezone": "America/New_York",
  "isp": "Google LLC",
  "org": "Google Public DNS",
  "as": "AS15169 Google LLC",
  "query": "8.8.8.8"
}
*/
type Resp struct {
	Status     string `json:"status"`
	Country    string `json:"country"`
	RegionName string `json:"regionName"`
	City       string `json:"city"`
	Isp        string `json:"isp"`
	Message    string `json:"message"`
}

func (r *Resp) valid() bool {
	return r.Status == `success` && r.Country != `` &&
		r.RegionName != `` && r.City != ``
}

func (r *Resp) String() string {
	if r.Message != "" {
		return r.Message
	}
	return fmt.Sprintf(`%s, %s, %s; %s`,
		r.Country, r.RegionName, r.City, r.Isp)
}

func (cg *CommentGeo) Queue(ip string, fn func()) error {
	defer func() {
		if fn != nil {
			fn()
		}
	}()

	if cg.lru.Contains(ip) {
		return nil
	}

	// log.Println(`GeoLocation queue:`, ip)

	// 限流目标网站的请求量。
	cg.mu.Lock()
	defer cg.mu.Unlock()

	if cg.lru.Contains(ip) {
		return nil
	}

	ctx, cancel := context.WithTimeout(cg.ctx, time.Second*3)
	defer cancel()

	u, err := url.Parse(`http://ip-api.com/json`)
	if err != nil {
		return err
	}
	u = u.JoinPath(ip)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if rsp.StatusCode != 200 {
		return fmt.Errorf(`status code = %d != 200`, rsp.StatusCode)
	}

	var r Resp
	if err := json.NewDecoder(rsp.Body).Decode(&r); err != nil {
		return fmt.Errorf(`decoding error: %v`, err)
	}
	if !r.valid() {
		// return fmt.Errorf(`invalid ip response`)
		log.Println(`GeoLocation: error:`, `invalid ip response`)
		// 错的也保存，防止重复查。
		cg.lru.Add(ip, &r)
		return nil
	}

	cg.lru.Add(ip, &r)
	return nil
}

func (cg *CommentGeo) Get(ip string) string {
	if r, ok := cg.lru.Get(ip); ok {
		return r.String()
	}
	return ""
}

func (cg *CommentGeo) GetTimeout(ip string, timeout time.Duration) string {
	ch := make(chan string, 1)
	// 如果线程返回前函数返回了，会写 closed channel
	// 那就不关了，不会泄漏
	// defer close(ch)

	go func() {
		cg.Queue(ip, func() {
			ch <- cg.Get(ip)
		})
	}()

	select {
	case <-time.After(timeout):
		return ""
	case s := <-ch:
		return s
	}
}

func (cg *CommentGeo) clearBad() {

}
