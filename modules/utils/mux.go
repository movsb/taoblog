package utils

import (
	"log"
	"net/http"
	"sync"
)

// ServeMuxWithMethod ...
type ServeMuxWithMethod struct {
	lock    sync.RWMutex
	methods map[string]*http.ServeMux
}

// NewServeMuxWithMethod ...
func NewServeMuxWithMethod() *ServeMuxWithMethod {
	mux := &ServeMuxWithMethod{
		methods: make(map[string]*http.ServeMux),
	}
	return mux
}

func (mux *ServeMuxWithMethod) getMuxLocked(method string) *http.ServeMux {
	mux.lock.RLock()
	hm := mux.methods[method]
	mux.lock.RUnlock()
	if hm == nil {
		mux.lock.Lock()
		hm = http.NewServeMux()
		mux.methods[method] = hm // explode?
		mux.lock.Unlock()
	}
	return hm
}

// ServeHTTP ...
func (mux *ServeMuxWithMethod) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := mux.getMuxLocked(r.Method)
	h, _ := m.Handler(r)
	log.Printf("admin: %-8s %s\n", r.Method, r.RequestURI)
	h.ServeHTTP(w, r)
}

func (mux *ServeMuxWithMethod) Handle(method, pattern string, handler http.Handler) {
	m := mux.getMuxLocked(method)
	m.Handle(pattern, handler)
}

func (mux *ServeMuxWithMethod) HandleFunc(method, pattern string, handler func(http.ResponseWriter, *http.Request)) {
	mux.Handle(method, pattern, http.HandlerFunc(handler))
}
