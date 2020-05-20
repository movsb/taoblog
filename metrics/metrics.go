package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = `taoblog`
)

// Server ...
type Server interface {
	http.Handler
	CountPost(PostID int64, Title string, IP string, userAgent string)
}

type _Server struct {
	handler     http.Handler
	registry    *prometheus.Registry
	counterPost *prometheus.CounterVec
}

// ServeHTTP ...
func (s *_Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.handler.ServeHTTP(w, req)
}

func (s *_Server) CountPost(PostID int64, Title string, IP string, userAgent string) {
	s.counterPost.WithLabelValues(
		fmt.Sprint(PostID),
		Title,
		IP,
		userAgent,
	).Inc()
}

// New ...
func New() Server {
	s := &_Server{
		registry: prometheus.NewRegistry(),
		counterPost: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      `post_counter`,
			Help:      `page view counter for post`,
		},
			[]string{
				`post_id`,
				`title`,
				`ip`,
				`user_agent`,
			},
		),
	}
	s.registry.MustRegister(s.counterPost)
	s.handler = promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{})
	return s
}
