package service

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type _Exporter struct {
	certDaysLeft   prometheus.Gauge
	domainDaysLeft prometheus.Gauge
	lock           sync.Mutex
	s              *Service
}

func (e *_Exporter) Describe(w chan<- *prometheus.Desc) {
	e.certDaysLeft.Describe(w)
	e.domainDaysLeft.Describe(w)
}

func (e *_Exporter) Collect(w chan<- prometheus.Metric) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.certDaysLeft.Collect(w)
	e.domainDaysLeft.Collect(w)
}

func _NewExporter(s *Service) *_Exporter {
	e := &_Exporter{s: s}
	e.certDaysLeft = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: `taoblog`,
			Subsystem: `domain`,
			Name:      `cert_days_left`,
		},
	)
	e.certDaysLeft.Set(float64(s.certDaysLeft.Load()))
	e.domainDaysLeft = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: `taoblog`,
			Subsystem: `domain`,
			Name:      `domain_days_left`,
		},
	)
	e.domainDaysLeft.Set(float64(s.domainExpirationDaysLeft.Load()))
	return e
}

func (s *Service) Exporter() prometheus.Collector {
	return s.exporter
}
