package metrics

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mssola/user_agent"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry ...
type Registry struct {
	ctx context.Context

	r *prometheus.Registry

	uptimeCounter prometheus.Counter

	homeCounter      prometheus.Counter
	pageViewCounter  *prometheus.CounterVec
	userAgentCounter *prometheus.CounterVec
}

// NewRegistry ...
func NewRegistry(ctx context.Context) *Registry {
	r := &Registry{
		ctx: ctx,
		r:   prometheus.NewRegistry(),
	}
	r.init()
	return r
}

// Handler ...
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.r, promhttp.HandlerOpts{})
}

// CountHome ...
func (r *Registry) CountHome() {
	r.homeCounter.Inc()
}

// CountPageView ...
func (r *Registry) CountPageView(id int64) {
	r.pageViewCounter.WithLabelValues(fmt.Sprint(id)).Inc()
}

// UserAgent ...
func (r *Registry) UserAgent(s string) {
	ua := user_agent.New(s)
	bn, bv := ua.Browser()
	os := ua.OSInfo()

	r.userAgentCounter.WithLabelValues(
		bool2string(ua.Bot()),
		strings.ToLower(bn), bv,
		bool2string(ua.Mobile()),
		strings.ToLower(os.Name), os.Version,
		strings.ToLower(ua.Platform()),
	).Inc()
}
