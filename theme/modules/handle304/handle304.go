package handle304

import (
	"fmt"
	"net/http"
	"time"

	"github.com/movsb/taoblog/modules/version"
)

const httpGmtFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

// Requester ...
type Requester interface {
	Request(req *http.Request) bool
}

// Responder ...
type Responder interface {
	Response(w http.ResponseWriter)
}

// ArticleRequest 判断是否符合 304 请求。
// 如果符合，返回 true；否则，返回 false。
func ArticleRequest(w http.ResponseWriter, req *http.Request, modified time.Time) bool {
	handlers := []Requester{
		NotModified{modified},
		EntityTag{
			modified: modified,
			version:  version.GitCommit,
		},
	}
	for _, handler := range handlers {
		if !handler.Request(req) {
			return false
		}
	}
	w.WriteHeader(http.StatusNotModified)
	return true
}

// ArticleResponse ...
func ArticleResponse(w http.ResponseWriter, modified time.Time) {
	handlers := []Responder{
		NotModified{modified},
		EntityTag{
			modified: modified,
			version:  version.GitCommit,
		},
	}
	for _, handler := range handlers {
		handler.Response(w)
	}
}

// NotModified ...
type NotModified struct {
	Modified time.Time
}

// Request ...
func (nm NotModified) Request(req *http.Request) bool {
	h := req.Header.Get(`If-Modified-Since`)
	t, _ := time.ParseInLocation(httpGmtFormat, h, time.UTC)
	return t.Equal(nm.Modified)
}

// Response ...
func (nm NotModified) Response(w http.ResponseWriter) {
	w.Header().Add(`Last-Modified`, nm.Modified.UTC().Format(httpGmtFormat))
}

// 格式：文章的修改日期-博客版本号
// Typically, the ETag value is a hash of the content, a hash of the last modification timestamp, or just a revision number.
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
type EntityTag struct {
	modified time.Time
	version  string
}

func (cm EntityTag) format() string {
	return fmt.Sprintf(`"%d-%s"`, cm.modified.Unix(), cm.version)
}

// Request ...
func (cm EntityTag) Request(req *http.Request) bool {
	haystack := req.Header.Get(`If-None-Match`)
	return haystack == cm.format()
}

// Response ...
func (cm EntityTag) Response(w http.ResponseWriter) {
	w.Header().Add(`ETag`, cm.format())
}
