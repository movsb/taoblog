package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry ...
type Registry struct {
	ctx context.Context

	r *prometheus.Registry

	uptimeCounter prometheus.Counter

	homeCounter     prometheus.Counter
	pageViewCounter *prometheus.CounterVec
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

func (r *Registry) init() {
	r.uptimeCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      `uptime`,
			Help:      `count seconds running`,
		},
	)
	r.startUptimeCounterAsync(time.Second * 15)
	r.r.MustRegister(r.uptimeCounter)

	r.homeCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      `home_counter`,
		},
	)
	r.r.MustRegister(r.homeCounter)

	r.pageViewCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      `page_view`,
		},
		[]string{`id`},
	)
	r.r.MustRegister(r.pageViewCounter)
}

func (r *Registry) startUptimeCounterAsync(interval time.Duration) {
	lastTime := time.Now()

	count := func() {
		now := time.Now()
		elapsed := now.Sub(lastTime)
		r.uptimeCounter.Add(elapsed.Seconds())
		lastTime = now
	}

	loop := func() {
		t := time.NewTicker(interval)
		defer t.Stop()

		for {
			select {
			case <-r.ctx.Done():
				return
			case <-t.C:
				count()
			}
		}
	}

	go loop()
}

// CountHome ...
func (r *Registry) CountHome() {
	r.homeCounter.Inc()
}

// CountPageView ...
func (r *Registry) CountPageView(id int64) {
	r.pageViewCounter.WithLabelValues(fmt.Sprint(id)).Inc()
}
