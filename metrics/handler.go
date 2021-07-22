package metrics

import (
	"context"
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

// Handler ...
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.r, promhttp.HandlerOpts{})
}
