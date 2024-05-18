package commentgeo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/phuslu/lru"
)

type CommentGeo struct {
	ctx   context.Context
	cache *lru.LRUCache[string, *Resp]
}

func New(ctx context.Context) *CommentGeo {
	geo := &CommentGeo{
		ctx:   ctx,
		cache: lru.NewLRUCache[string, *Resp](128),
	}
	return geo
}

func (cg *CommentGeo) Get(ip string) string {
	if r, ok := cg.cache.Get(ip); ok {
		return r.String()
	}
	return ""
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

// 异步入队列。
func (cg *CommentGeo) Queue(ip string) {
	go cg.cache.GetOrLoad(cg.ctx, ip, fetch)
}

func fetch(ctx context.Context, ip string) (*Resp, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	u, err := url.Parse(`http://ip-api.com/json`)
	if err != nil {
		return nil, err
	}
	u = u.JoinPath(ip)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != 200 {
		return nil, fmt.Errorf(`status code = %d != 200`, rsp.StatusCode)
	}

	var r Resp
	if err := json.NewDecoder(rsp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf(`decoding error: %v`, err)
	}
	if !r.valid() {
		return nil, fmt.Errorf(`geo: invalid response`)
	}
	return &r, nil
}
