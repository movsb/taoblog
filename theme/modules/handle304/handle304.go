package handle304

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching#force_revalidation
func MustRevalidate(w http.ResponseWriter) {
	w.Header().Add(`Cache-Control`, `max-age=0, must-revalidate`)
}

func CacheShortly(w http.ResponseWriter) {
	w.Header().Add(`Cache-Control`, `max-age=600, must-revalidate`)
}

type Handler interface {
	Match(w http.ResponseWriter, r *http.Request) bool
	Response(w http.ResponseWriter)
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
func (b *_Bundle) Response(w http.ResponseWriter) {
	for _, h := range b.handlers {
		h.Response(w)
	}
}

func New(handlers ...Handler) Handler {
	return &_Bundle{handlers: handlers}
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

func (nm NotModified) Response(w http.ResponseWriter) {
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

func (cm EntityTag) Response(w http.ResponseWriter) {
	w.Header().Add(`ETag`, cm.format())
}
