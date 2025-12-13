package debug

import (
	"crypto/rand"
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"sync"
)

var (
	sessions = make(map[string]struct{})
	lock     sync.Mutex
)

func Enter() string {
	lock.Lock()
	defer lock.Unlock()

	buf := make([]byte, 16)
	rand.Read(buf)

	s := fmt.Sprintf(`%x`, buf)
	sessions[s] = struct{}{}
	return s
}

func Leave(s string) {
	lock.Lock()
	defer lock.Unlock()
	delete(sessions, s)
}

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

	mux.HandleFunc(`/`, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `<!DOCTYPE html>
<html>
<head>
</head>
<body>
	<ul>
		<li><a href="vars">vars</a></li>
		<li><a href="pprof/">pprof</a></li>
	</ul>
</body>
</html>
 `)
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := r.PathValue(`session`)
		lock.Lock()
		if _, found := sessions[session]; !found {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			lock.Unlock()
			return
		}
		lock.Unlock()

		// Path: /debug/session/...
		http.StripPrefix(fmt.Sprintf(`/debug/%s`, session), mux).ServeHTTP(w, r)
	})
}
