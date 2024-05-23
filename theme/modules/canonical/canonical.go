package canonical

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/movsb/taoblog/modules/metrics"
	"github.com/movsb/taoblog/modules/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Renderer ...
type Renderer interface {
	Exception(w http.ResponseWriter, req *http.Request, e any) bool
	ProcessHomeQueries(w http.ResponseWriter, req *http.Request, query url.Values) bool
	QueryHome(w http.ResponseWriter, req *http.Request) error
	QueryByID(w http.ResponseWriter, req *http.Request, id int64)
	QueryFile(w http.ResponseWriter, req *http.Request, postID int64, file string)
	QueryByTags(w http.ResponseWriter, req *http.Request, tags []string)
	QueryByPage(w http.ResponseWriter, req *http.Request, path string) (int64, error)
	QueryStatic(w http.ResponseWriter, req *http.Request, file string)
	QuerySpecial(w http.ResponseWriter, req *http.Request, file string) bool
}

// Canonical ...
type Canonical struct {
	mux      *http.ServeMux
	renderer Renderer
	mr       *metrics.Registry
}

// New ...
func New(renderer Renderer, mr *metrics.Registry) *Canonical {
	c := &Canonical{
		mux:      http.NewServeMux(),
		renderer: renderer,
		mr:       mr,
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
			c.mr.CountHome()
			c.mr.UserAgent(req.UserAgent())
			return
		}

		if regexpByID.MatchString(path) {
			matches := regexpByID.FindStringSubmatch(path)
			if slash := matches[2]; slash == `` {
				http.Redirect(w, req,
					fmt.Sprintf(`/%s/`, matches[1]),
					http.StatusPermanentRedirect,
				)
				return
			}
			id := utils.MustToInt64(matches[1])
			if id <= 0 {
				panic(status.Error(codes.NotFound, ""))
			}
			c.renderer.QueryByID(w, req, id)
			c.mr.CountPageView(id)
			c.mr.UserAgent(req.UserAgent())
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

		if regexpByPage.MatchString(path) && isCategoryPath(path) {
			matches := regexpByPage.FindStringSubmatch(path)
			id, err := c.renderer.QueryByPage(w, req, matches[0])
			if err == nil {
				c.mr.CountPageView(id)
				c.mr.UserAgent(req.UserAgent())
			}
			return
		}

		if strings.HasSuffix(path, "/") {
			w.WriteHeader(http.StatusForbidden)
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
		if strings.Contains(path, `/./`) || strings.Contains(path, `/../`) {
			return false
		}
		if strings.HasSuffix(path, `/.`) || strings.HasSuffix(path, `/..`) {
			return false
		}
	}

	return true
}
