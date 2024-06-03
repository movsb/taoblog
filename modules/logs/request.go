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

func NewRequestLoggerHandler(path string) func(http.Handler) http.Handler {
	return NewRequestLogger(path).Handler
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

// https://blog.twofei.com/909/
type _ResponseWriter struct {
	http.ResponseWriter
	*http.ResponseController
	code int
}

func (w *_ResponseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (l *RequestLogger) Handler(h http.Handler) http.Handler {
	tz := time.FixedZone(`China`, 8*60*60)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		takeOver := &_ResponseWriter{
			ResponseWriter:     w,
			ResponseController: http.NewResponseController(w),
			code:               200,
		}
		h.ServeHTTP(takeOver, r)
		l.counter.Store(0)
		l.lock.Lock()
		defer l.lock.Unlock()
		now := time.Now().In(tz).Format(`2006-01-02 15:04:05`)
		ac := auth.Context(r.Context())
		fmt.Fprintf(l.f,
			"%s %-15s %3d %-8s %-32s %-32s %-32s\n",
			now, ac.RemoteAddr.String(), takeOver.code, r.Method, r.RequestURI, r.Referer(), r.Header.Get(`User-Agent`),
		)
	})
}
