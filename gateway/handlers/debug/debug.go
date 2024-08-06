package debug

import (
	"expvar"
	"net/http"
	_ "net/http/pprof"
	"strconv"
)

func Handler() http.Handler {
	mux := http.NewServeMux()

	mux.Handle(`GET /vars`, expvar.Handler())
	mux.HandleFunc(`POST /vars/{var}`, func(w http.ResponseWriter, r *http.Request) {
		va := expvar.Get(r.PathValue(`var`))
		if va == nil {
			http.NotFound(w, r)
			return
		}
		switch typed := va.(type) {
		case *expvar.Int:
			i, err := strconv.Atoi(r.URL.Query().Get(`v`))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			typed.Set(int64(i))
		default:
			http.Error(w, `unknown var type`, http.StatusBadRequest)
			return
		}
	})

	mux.HandleFunc(`/pprof/`, func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = `/debug` + r.URL.Path
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	return mux
}
