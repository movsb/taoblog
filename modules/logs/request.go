package logs

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/movsb/taoblog/modules/auth"
)

type RequestLogger struct {
	f *os.File

	// 每隔几秒手动 sync 一下。
	lock    sync.Mutex
	counter atomic.Int32
}

func NewRequestLogger(path string) *RequestLogger {
	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}

	l := &RequestLogger{
		f: fp,
	}

	go func() {
		const interval = 3
		t := time.NewTicker(time.Second * interval)
		defer t.Stop()
		for range t.C {
			l.counter.Add(interval)
			if n := l.counter.Load(); n >= 10 {
				l.lock.Lock()
				l.f.Sync()
				l.lock.Unlock()
				l.counter.Store(0)
			}
		}
	}()

	return l
}

// TODO：Go 的内嵌接口不支持展开，
// 比如 http.ResponseWriter 如果实现了 Hijacker，
// _ResponseWriter 虽然内嵌了 http.ResponseWriter，
// 但是不会自动实现这个 Hijacker 接口。有点儿无解。即便换成 any 也不行。
// 所以暂时不处理 WebSocket 的返回码。
// 更新：先内嵌部分已知的接口以缓解这个问题。
// https://blog.twofei.com/909/
type _ResponseWriter struct {
	http.ResponseWriter
	http.Hijacker
	code int
}

func (w *_ResponseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (l *RequestLogger) Handler(h http.Handler) http.Handler {
	tz := time.FixedZone(`China`, 8*60*60)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var hijacker http.Hijacker
		if h, ok := w.(http.Hijacker); ok {
			hijacker = h
		}
		ww := &_ResponseWriter{w, hijacker, 200}
		h.ServeHTTP(ww, r)
		l.counter.Store(0)
		l.lock.Lock()
		defer l.lock.Unlock()
		now := time.Now().In(tz).Format(`2006-01-02 15:04:05`)
		ac := auth.Context(r.Context())
		fmt.Fprintf(l.f,
			"%s %-15s %3d %-8s %-32s %-32s %-32s\n",
			now, ac.RemoteAddr.String(), ww.code, r.Method, r.RequestURI, r.Referer(), r.Header.Get(`User-Agent`),
		)
	})
}
