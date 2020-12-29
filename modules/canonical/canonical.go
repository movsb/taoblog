package canonical

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/pingback"
)

// Renderer ...
type Renderer interface {
	Exception(w http.ResponseWriter, req *http.Request, e interface{}) bool
	ProcessHomeQueries(w http.ResponseWriter, req *http.Request, query url.Values) bool
	QueryHome(w http.ResponseWriter, req *http.Request)
	QueryByID(w http.ResponseWriter, req *http.Request, id int64)
	QueryFile(w http.ResponseWriter, req *http.Request, postID int64, file string)
	QueryByTags(w http.ResponseWriter, req *http.Request, tags []string)
	QueryBySlug(w http.ResponseWriter, req *http.Request, tree string, slug string)
	QueryByPage(w http.ResponseWriter, req *http.Request, parents string, slug string)
	QueryStatic(w http.ResponseWriter, req *http.Request, file string)
	QuerySpecial(w http.ResponseWriter, req *http.Request, file string) bool
}

// Canonical ...
type Canonical struct {
	// TODO(movsb): hack!
	PingbackURL string
	mux         *http.ServeMux
	renderer    Renderer
}

// New ...
func New(renderer Renderer) *Canonical {
	c := &Canonical{
		mux:      http.NewServeMux(),
		renderer: renderer,
	}
	return c
}

// ServeHTTP implements htt.Handler.
func (c *Canonical) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			if !c.renderer.Exception(w, req, e) {
				panic(e)
			}
		}
	}()

	path := req.URL.Path
	if !isValidPath(path) {
		c.renderer.Exception(w, req, errors.New(`invalid request path`))
		return
	}

	if req.Method == http.MethodGet || req.Method == http.MethodOptions {
		if regexpHome.MatchString(path) {
			if c.renderer.ProcessHomeQueries(w, req, req.URL.Query()) {
				return
			}
			c.renderer.QueryHome(w, req)
			return
		}

		if regexpByID.MatchString(path) {
			matches := regexpByID.FindStringSubmatch(path)
			if slash := matches[2]; slash == `` {
				w.Header().Set(`Location`, matches[1]+`/`)
				w.WriteHeader(301)
				return
			}
			id := utils.MustToInt64(matches[1])
			// TODO(movsb): hack! hack!
			w.Header().Set(pingback.Header, c.PingbackURL)
			c.renderer.QueryByID(w, req, id)
			return
		}

		if regexpFile.MatchString(path) {
			matches := regexpFile.FindStringSubmatch(path)
			postID := utils.MustToInt64(matches[1])
			file := matches[2]
			c.renderer.QueryFile(w, req, postID, file)
			return
		}

		if regexpByTags.MatchString(path) {
			matches := regexpByTags.FindStringSubmatch(path)
			tags := strings.Split(matches[1], `+`)
			c.renderer.QueryByTags(w, req, tags)
			return
		}

		if c.renderer.QuerySpecial(w, req, path) {
			return
		}

		if regexpBySlug.MatchString(path) && isCategoryPath(path) {
			matches := regexpBySlug.FindStringSubmatch(path)
			tree := matches[1]
			slug := matches[2]
			c.renderer.QueryBySlug(w, req, tree, slug)
			return
		}

		if regexpByPage.MatchString(path) && isCategoryPath(path) {
			matches := regexpByPage.FindStringSubmatch(path)
			parents := matches[1]
			if parents != "" {
				parents = parents[1:]
			}
			slug := matches[3]
			c.renderer.QueryByPage(w, req, parents, slug)
			return
		}

		if strings.HasSuffix(path, "/") {
			w.WriteHeader(403)
			return
		}

		c.renderer.QueryStatic(w, req, path)
		return
	}

	http.NotFound(w, req)
}

func isCategoryPath(path string) bool {
	p := strings.IndexByte(path[1:], '/')
	if p == -1 {
		return true
	}
	p++
	first := path[0 : p+1]
	if _, ok := nonCategoryNames[first]; ok {
		return false
	}
	return true
}

func isValidPath(path string) bool {
	if len(path) == 0 || path[0] != '/' {
		return false
	}

	// We reject all requests with path containing `.` or `..` components.
	if p := strings.Index(path, `/.`); p != -1 { // fast path
		if strings.Index(path, `/./`) != -1 || strings.Index(path, `/../`) != -1 {
			return false
		}
		if strings.HasSuffix(path, `/.`) || strings.HasSuffix(path, `/..`) {
			return false
		}
	}

	return true
}
