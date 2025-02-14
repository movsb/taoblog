package addons

import "net/http"

type _AddonRegistry struct {
	*http.ServeMux
}

var registry *_AddonRegistry

func init() {
	registry = New()
}

func New() *_AddonRegistry {
	registry = &_AddonRegistry{
		ServeMux: http.NewServeMux(),
	}
	return registry
}

func Handler() http.Handler {
	return registry
}

func Handle(pattern string, handler http.Handler) {
	registry.Handle(pattern, handler)
}
