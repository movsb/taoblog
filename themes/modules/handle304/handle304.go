package handle304

import (
	"net/http"
	"time"

	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/modules/version"
)

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
	if h == `` {
		return false
	}
	return datetime.Gmt2Time(h).Equal(nm.Modified)
}

// Response ...
func (nm NotModified) Response(w http.ResponseWriter) {
	w.Header().Add(`Last-Modified`, datetime.Time2Gmt(nm.Modified))
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
