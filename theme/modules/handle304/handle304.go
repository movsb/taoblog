package handle304

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching#force_revalidation
func MustRevalidate(w http.ResponseWriter) {
	w.Header().Add(`Cache-Control`, `private, no-cache, must-revalidate`)
}

type Handler interface {
	Match(w http.ResponseWriter, r *http.Request) bool
	Respond(w http.ResponseWriter)
}

func WithNotModified(t time.Time) Handler {
	return &NotModified{
		Modified: t,
	}
}
func WithEntityTag(fields ...any) Handler {
	return EntityTag{fields: fields}
}

type _Bundle struct {
	h        http.Handler
	handlers []Handler
}

func (b *_Bundle) Match(w http.ResponseWriter, r *http.Request) bool {
	for _, h := range b.handlers {
		if !h.Match(w, r) {
			return false
		}
	}
	w.WriteHeader(http.StatusNotModified)
	return true
}
func (b *_Bundle) Respond(w http.ResponseWriter) {
	for _, h := range b.handlers {
		h.Respond(w)
	}
	MustRevalidate(w)
}

// NOTE: 调用此方法必须 h 不为空。
func (b *_Bundle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if b.Match(w, r) {
		return
	}
	b.Respond(w)
	// TODO 移除到独立的缓存处理器中。
	MustRevalidate(w)
	b.h.ServeHTTP(w, r)
}

type BundleHandler interface {
	Handler
	http.Handler
}

func New(h http.Handler, handlers ...Handler) BundleHandler {
	return &_Bundle{
		h:        h,
		handlers: handlers,
	}
}

// NotModified ...
type NotModified struct {
	Modified time.Time
}

func (nm NotModified) Match(w http.ResponseWriter, r *http.Request) bool {
	h := r.Header.Get(`If-Modified-Since`)
	t, _ := time.ParseInLocation(http.TimeFormat, h, time.UTC)
	return t.Equal(nm.Modified)
}

func (nm NotModified) Respond(w http.ResponseWriter) {
	w.Header().Add(`Last-Modified`, nm.Modified.UTC().Format(http.TimeFormat))
}

// 格式形如：文章的修改日期-博客版本号
// Typically, the ETag value is a hash of the content, a hash of the last modification timestamp, or just a revision number.
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
type EntityTag struct {
	fields []any
}

func (cm EntityTag) format() string {
	var values []string
	for _, v := range cm.fields {
		var val string
		switch typed := v.(type) {
		case time.Time:
			val = fmt.Sprint(typed.Unix())
		case func() time.Time:
			val = fmt.Sprint(typed().Unix())
		default:
			val = fmt.Sprint(typed)
		}
		values = append(values, val)
	}
	return strings.Join(values, "-")
}

func (cm EntityTag) Match(w http.ResponseWriter, r *http.Request) bool {
	haystack := r.Header.Get(`If-None-Match`)
	return haystack == cm.format()
}

func (cm EntityTag) Respond(w http.ResponseWriter) {
	w.Header().Add(`ETag`, cm.format())
}
