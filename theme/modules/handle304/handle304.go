package handle304

import (
	"net/http"
	"time"

	"github.com/movsb/taoblog/modules/version"
)

const httpgmtFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

// Requester ...
type Requester interface {
	Request(req *http.Request) bool
}

// Responder ...
type Responder interface {
	Response(w http.ResponseWriter)
}

// ArticleRequest ...
func ArticleRequest(w http.ResponseWriter, req *http.Request, modified time.Time) bool {
	handlers := []Requester{
		NotModified{modified},
		CommitMatch{},
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
		CommitMatch{},
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
	t, _ := time.ParseInLocation(httpgmtFormat, h, time.UTC)
	return t.Equal(nm.Modified)
}

// Response ...
func (nm NotModified) Response(w http.ResponseWriter) {
	w.Header().Add(`Last-Modified`, nm.Modified.UTC().Format(httpgmtFormat))
}

// CommitMatch ...
type CommitMatch struct {
}

// Request ...
func (cm CommitMatch) Request(req *http.Request) bool {
	commitCookie, err := req.Cookie(`commit`)
	if err != nil {
		return false
	}
	return commitCookie.Value == version.GitCommit
}

// Response ...
func (cm CommitMatch) Response(w http.ResponseWriter) {
	v := version.GitCommit
	if v == `` {
		v = `HEAD`
	}
	http.SetCookie(w, &http.Cookie{
		Name:     `commit`,
		Value:    v,
		Path:     `/`,
		MaxAge:   0,
		Secure:   false,
		HttpOnly: true,
	})
}
