package hostdare

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func New(username, password string) (prometheus.Collector, error) {
	hd, err := NewHostDare(username, password)
	if err != nil {
		return nil, err
	}
	return NewHostDareExporter(hd), nil
}

type HostDare struct {
	client   *http.Client
	redirect *url.URL
}

func NewHostDare(username, password string) (*HostDare, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{
		Jar: jar,
	}
	hd := &HostDare{
		client: httpClient,
	}
	if err := hd.login(username, password); err != nil {
		return nil, fmt.Errorf(`error logging in: %v`, err)
	}
	return hd, nil
}

type LoginResponse struct {
	Redirect string `json:"redirect"`
}

func (hd *HostDare) login(username, password string) error {
	values := url.Values{}
	values.Set(`username`, username)
	values.Set(`password`, password)
	values.Set(`login`, `1`)
	rsp, err := hd.client.PostForm(`https://vps.hostdare.com/index.php?api=json&act=login`, values)
	if err != nil {
		return fmt.Errorf(`error posting form: %v`, err)
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		return fmt.Errorf(`unexpected status code: %v`, rsp.Status)
	}
	var lr LoginResponse
	if err := json.NewDecoder(rsp.Body).Decode(&lr); err != nil {
		return fmt.Errorf(`error decoding response json: %v`, err)
	}
	if lr.Redirect == `` || !strings.HasPrefix(lr.Redirect, `/sess`) {
		return fmt.Errorf(`no valid redirect was returned: %s`, lr.Redirect)
	}
	ru, err := url.Parse(lr.Redirect)
	if err != nil {
		return fmt.Errorf(`error parsing redirect url: [%s] %v`, lr.Redirect, err)
	}
	hd.redirect = ru
	return nil
}

type Status int

const (
	StatusUp Status = 1
)

// buggy: bits not standard compliant, but most time works well.
type IntAsString int64

func (i IntAsString) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprint(i)), nil
}

func (i *IntAsString) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	n, err := strconv.ParseInt(s, 10, 53)
	if err != nil {
		return err
	}
	*(*int64)(i) = n
	return nil
}

type ListServersResponse struct {
	Counts struct {
		VPS IntAsString `json:"vps"`
	} `json:"counts"`
	Status  map[int]Status `json:"status"`
	Servers map[int]struct {
		ID       IntAsString `json:"vpsid"`
		HostName string      `json:"hostname"`
	} `json:"vs"`
}

// 不支持分页。
func (hd *HostDare) ListServers() (*ListServersResponse, error) {
	u, err := url.Parse(`https://vps.hostdare.com/sessXXX/index.php?api=json&act=listvs&&random=0.2665219561593345`)
	if err != nil {
		return nil, fmt.Errorf(`invalid reference url: %v`, err)
	}
	u.Path = hd.redirect.Path
	v := u.Query()
	v.Set(`random`, fmt.Sprint(rand.Float32()))
	u.RawQuery = v.Encode()

	rsp, err := hd.client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf(`error listing servers: %v`, err)
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`invalid status code: %v`, rsp.Status)
	}

	var lsr ListServersResponse
	if err := json.NewDecoder(rsp.Body).Decode(&lsr); err != nil {
		return nil, fmt.Errorf(`error parsing json: %v`, err)
	}

	return &lsr, nil
}

type StatsServerResponse struct {
	Info struct {
		ID        IntAsString `json:"vpsid"`
		Bandwidth struct {
			Limit float64 `json:"limit"`
			Used  float64 `json:"used"`
			// Free  float64         `json:"free"`
			Usage map[int]float64 `json:"usage"`
			In    map[int]float64 `json:"in"`
			Out   map[int]float64 `json:"out"`
		} `json:"bandwidth"`
	} `json:"info"`
}

// https://vps.hostdare.com/sessXXX/index.php?api=json&act=vpsmanage&stats=1&svs=43546
func (hd *HostDare) Stats(vpsID int) (*StatsServerResponse, error) {
	u, err := url.Parse(`https://vps.hostdare.com/sessXXX/index.php?api=json&act=vpsmanage&stats=1&svs=43546&random=0.2665219561593345`)
	if err != nil {
		return nil, fmt.Errorf(`invalid reference url: %v`, err)
	}
	u.Path = hd.redirect.Path
	v := u.Query()
	v.Set(`random`, fmt.Sprint(rand.Float32()))
	v.Set(`svs`, fmt.Sprint(vpsID))
	u.RawQuery = v.Encode()

	rsp, err := hd.client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf(`error listing servers: %v`, err)
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`invalid status code: %v`, rsp.Status)
	}

	var ssr StatsServerResponse
	if err := json.NewDecoder(rsp.Body).Decode(&ssr); err != nil {
		return nil, fmt.Errorf(`error parsing json: %v`, err)
	}

	return &ssr, nil
}

type HostDareExporter struct {
	hd *HostDare

	bandwidthQuota *prometheus.GaugeVec
	bandwidthUsed  *prometheus.GaugeVec

	lock sync.Mutex
}

func NewHostDareExporter(hd *HostDare) *HostDareExporter {
	hde := HostDareExporter{
		hd: hd,
	}

	hde.bandwidthQuota = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: `vps`,
			Subsystem: `bandwidth`,
			Name:      `quota`,
			Help:      `Bandwidth quota`,
		},
		[]string{`vendor`, `name`},
	)
	hde.bandwidthUsed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: `vps`,
			Subsystem: `bandwidth`,
			Name:      `used`,
			Help:      `Bandwidth used`,
		},
		[]string{`vendor`, `name`},
	)

	go hde.run(context.TODO())

	return &hde
}

func (hde *HostDareExporter) updateStats() error {
	lsr, err := hde.hd.ListServers()
	if err != nil {
		return err
	}

	stats := []*StatsServerResponse{}
	for _, s := range lsr.Servers {
		stat, err := hde.hd.Stats(int(s.ID))
		if err != nil {
			log.Println(`error updating:`, err)
			return err
		}
		stats = append(stats, stat)
	}

	for _, server := range lsr.Servers {
		for _, stat := range stats {
			if server.ID == stat.Info.ID {
				hde.bandwidthQuota.WithLabelValues(`hostdare`, server.HostName).Set(stat.Info.Bandwidth.Limit * (1 << 20))
				hde.bandwidthUsed.WithLabelValues(`hostdare`, server.HostName).Set(stat.Info.Bandwidth.Used * (1 << 20))
			}
		}
	}

	return nil
}

func (hde *HostDareExporter) Describe(w chan<- *prometheus.Desc) {
	hde.bandwidthQuota.Describe(w)
	hde.bandwidthUsed.Describe(w)
}

func (hde *HostDareExporter) run(ctx context.Context) {
	time.Sleep(time.Second * 10)
	hde.updateStats()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Minute * 3):
			if err := hde.updateStats(); err != nil {
				log.Println(`error describing:`, err)
			}
		}
	}
}

func (hde *HostDareExporter) Collect(w chan<- prometheus.Metric) {
	hde.lock.Lock()
	defer hde.lock.Unlock()

	hde.bandwidthQuota.Collect(w)
	hde.bandwidthUsed.Collect(w)
}
