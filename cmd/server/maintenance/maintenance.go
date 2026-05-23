package maintenance

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/movsb/taoblog/modules/utils"
)

type MaintenanceMode interface {
	Enter(reason string, duration time.Duration)
	Leave()
	// 如果正值维护模式，返回 true 的 exception 将可以正确请求。
	Handler(exception func(ctx context.Context, r *http.Request) bool) func(http.Handler) http.Handler
}

var _ MaintenanceMode = (*Maintenance)(nil)

func New() *Maintenance {
	// 为了测试不重复注册，特地注册为全局。
	// TODO 解决测试重复注册的问题
	// if version.DevMode() {
	enabled := expvar.NewInt(fmt.Sprintf(`%s:%d`, `maintenance:`, time.Now().UnixNano()))
	// } else {
	// 	enabled = expvar.NewInt(`maintenance`)
	// }
	return &Maintenance{
		enabled: enabled,
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

type _Snapshot struct {
	Message   string
	Estimated time.Duration
	In        bool
}

func (s _Snapshot) MessageString() string {
	return s.Message
}
func (s _Snapshot) EstimatedString() string {
	if s.Estimated < 0 {
		return `(未知)`
	}
	return time.Now().Add(s.Estimated).Format(time.RFC3339)
}

func (m *Maintenance) snapshot() _Snapshot {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return _Snapshot{
		Message:   m.Message,
		Estimated: m.Estimated,
		In:        m.in(),
	}
}

func (m *Maintenance) Handler(exception func(ctx context.Context, r *http.Request) bool) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		tmpl := utils.Must1(template.New("").Parse(`网站不可用，请稍候再试。

原因：{{.MessageString}}
时间：{{.EstimatedString}}
`))
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			snapshot := m.snapshot()
			if snapshot.In && (exception == nil || !exception(r.Context(), r)) {
				if snapshot.Estimated > 0 {
					w.Header().Add(`Retry-After`, fmt.Sprint(int32(snapshot.Estimated.Seconds())))
				}
				w.WriteHeader(http.StatusServiceUnavailable)
				tmpl.Execute(w, snapshot)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}
