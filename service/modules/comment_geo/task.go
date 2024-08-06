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

	"github.com/movsb/taoblog/modules/utils"
	"github.com/phuslu/lru"
)

type Task struct {
	store utils.PluginStorage
	deb   *utils.Debouncer
	cache *lru.TTLCache[string, string]
	lock  sync.Mutex
	keys  []string
	ch    chan string
}

func NewTask(store utils.PluginStorage) *Task {
	t := &Task{
		store: store,
		cache: lru.NewTTLCache[string, string](1024 * 64),
		ch:    make(chan string, 1024),
	}
	t.deb = utils.NewDebouncer(time.Second*10, t.save)
	t.load()
	go t.fetch()
	return t
}

const ttl = time.Hour * 24 * 30

func (t *Task) load() {
	cached, err := t.store.Get(`cache`)
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
		t.cache.Set(k, v, ttl)
		t.keys = append(t.keys, k)
	}
	log.Println(`恢复了 IP 地理位置数据`)
}

func (t *Task) save() {
	log.Println(`即将保存 IP 地理位置数据`)
	t.lock.Lock()
	defer t.lock.Unlock()
	m := map[string]string{}
	for _, k := range t.keys {
		if value, ok := t.cache.Get(k); ok {
			m[k] = value
		}
	}
	data := string(utils.Must1(json.Marshal(m)))
	t.store.Set(`cache`, data)
}

func (t *Task) Get(ip string) string {
	if value, ok := t.cache.Get(ip); ok && value != "" {
		return value
	}

	select {
	case t.ch <- ip:
	default:
		log.Println(`队列溢出了`)
	}
	return ""
}

func (t *Task) fetch() {
	for {
		// This endpoint is limited to 45 requests per minute from an IP address.
		time.Sleep(time.Second * 2)

		ip := <-t.ch

		rsp, err := fetch(context.TODO(), ip)
		if err != nil {
			log.Println(ip, err)
			t.lock.Lock()
			t.cache.Set(ip, err.Error(), time.Hour)
			t.lock.Unlock()
		} else {
			t.lock.Lock()
			t.cache.Set(ip, rsp.String(), ttl)
			log.Println(ip, rsp.String())
			t.keys = append(t.keys, ip)
			t.deb.Enter()
			t.lock.Unlock()
		}
	}
}

// This endpoint is limited to 45 requests per minute from an IP address.
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

func fetch(ctx context.Context, ip string) (*Resp, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	u, err := url.Parse(`http://ip-api.com/json?lang=zh-CN`)
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
		return nil, fmt.Errorf(`geo: invalid response: %v`, ip)
	}
	return &r, nil
}
