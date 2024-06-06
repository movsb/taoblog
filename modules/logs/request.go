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

type Option func(*RequestLogger)

func WithSentBytesCounter(a SentBytesCounter) Option {
	return func(rl *RequestLogger) {
		rl.sentBytes = a
	}
}

type RequestLogger struct {
	f *os.File

	// 每隔几秒手动 sync 一下。
	lock    sync.Mutex
	counter atomic.Int32

	sentBytes SentBytesCounter
}

func NewRequestLoggerHandler(path string, options ...Option) func(http.Handler) http.Handler {
	return NewRequestLogger(path, options...).Handler
}

func NewRequestLogger(path string, options ...Option) *RequestLogger {
	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}

	l := &RequestLogger{
		f: fp,
	}

	for _, opt := range options {
		opt(l)
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

type SentBytesCounter interface {
	CountSentBytes(ip string, sentBytes int)
}

// https://blog.twofei.com/909/
type _ResponseWriter struct {
	http.ResponseWriter
	*http.ResponseController

	code      int
	sentBytes int // 服务端 → 客户端
}

func (w *_ResponseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *_ResponseWriter) Write(b []byte) (int, error) {
	// 写没写成功都算。
	w.sentBytes += len(b)

	return w.ResponseWriter.Write(b)
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
		// 仅统计非管理员用户。
		if l.sentBytes != nil && !ac.User.IsAdmin() {
			l.sentBytes.CountSentBytes(ac.RemoteAddr.String(), takeOver.sentBytes)
		}
	})
}
