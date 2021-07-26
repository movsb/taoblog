package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

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

	r.userAgentCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      `user_agent`,
		},
		[]string{`bot`, `browser_name`, `browser_version`, `mobile`, `os_name`, `os_version`, `platform`},
	)
	r.r.MustRegister(r.userAgentCounter)
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
