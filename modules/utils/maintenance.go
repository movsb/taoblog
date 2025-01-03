package utils

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"sync"
	"text/template"
	"time"
)

func NewMaintenance() *Maintenance {
	return &Maintenance{
		enabled: expvar.NewInt(`maintenance`),
	}
}

type Maintenance struct {
	Message   string
	Estimated time.Duration
	lock      sync.RWMutex

	enabled *expvar.Int
}

func (m *Maintenance) Enter(message string, estimated time.Duration) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Message = message
	m.Estimated = estimated
	m.enabled.Set(1)
}

func (m *Maintenance) in() bool {
	return m.Estimated != 0 || m.enabled.Value() > 0
}

func (m *Maintenance) Leave() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Estimated = 0
	m.enabled.Set(0)
}

func (m *Maintenance) MessageString() string {
	return m.Message
}

func (m *Maintenance) EstimatedString() string {
	if m.Estimated < 0 {
		return `(未知)`
	}
	return time.Now().Add(m.Estimated).Format(time.RFC3339)
}

func (m *Maintenance) Handler(exception func(ctx context.Context) bool) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		tmpl := Must1(template.New("").Parse(`网站不可用，请稍候再试。

原因：{{.MessageString}}
时间：{{.EstimatedString}}
`))
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.lock.RLock()
			copy := *m //no warn
			m.lock.RUnlock()
			if copy.in() && (exception == nil || !exception(r.Context())) {
				if m.Estimated > 0 {
					w.Header().Add(`Retry-After`, fmt.Sprint(int32(m.Estimated.Seconds())))
				}
				w.WriteHeader(http.StatusServiceUnavailable)
				tmpl.Execute(w, m)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}
