package canonical

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/movsb/taoblog/metrics"
	"github.com/movsb/taoblog/modules/utils"
)

type RedirectFinder interface {
	FindRedirect(sourcePath string) (targetPath string, err error)
}

// Renderer ...
type Renderer interface {
	Exception(w http.ResponseWriter, req *http.Request, e interface{}) bool
	ProcessHomeQueries(w http.ResponseWriter, req *http.Request, query url.Values) bool
	QueryHome(w http.ResponseWriter, req *http.Request) error
	QueryByID(w http.ResponseWriter, req *http.Request, id int64) error
	QueryFile(w http.ResponseWriter, req *http.Request, postID int64, file string)
	QueryByTags(w http.ResponseWriter, req *http.Request, tags []string)
	QueryBySlug(w http.ResponseWriter, req *http.Request, tree string, slug string) (int64, error)
	QueryByPage(w http.ResponseWriter, req *http.Request, parents string, slug string) (int64, error)
	QueryStatic(w http.ResponseWriter, req *http.Request, file string)
	QuerySpecial(w http.ResponseWriter, req *http.Request, file string) bool
}

type Renderers struct {
	themes  map[string]Renderer
	Default string
}

func (r *Renderers) Get(name string) Renderer {
	if name == `` {
		name = r.Default
	}
	return r.themes[name]
}

func (r *Renderers) Add(name string, renderer Renderer) {
	if r.themes == nil {
		r.themes = make(map[string]Renderer)
	}
	r.themes[name] = renderer
}

// Canonical ...
type Canonical struct {
	mux *http.ServeMux
	rs  *Renderers
	mr  *metrics.Registry
}

// New ...
func New(rs *Renderers, mr *metrics.Registry) *Canonical {
	c := &Canonical{
		mux: http.NewServeMux(),
		rs:  rs,
		mr:  mr,
	}
	return c
}

func (c *Canonical) r(w http.ResponseWriter, r *http.Request) Renderer {
	theme := r.URL.Query().Get(`_theme`)
	themeCookie, _ := r.Cookie(`theme`)

	// We prefer theme from query, and save it in cookie if it differs.
	if theme != `` {
		if themeCookie == nil || themeCookie.Value != theme {
			http.SetCookie(w, &http.Cookie{
				Name:     `theme`,
				Value:    theme,
				Path:     `/`,
				MaxAge:   int((time.Hour * 24).Seconds()),
				HttpOnly: true,
			})
		}
	} else {
		if themeCookie != nil && themeCookie.Value != `` {
			theme = themeCookie.Value
		}
	}

	return c.rs.Get(theme)
}

// ServeHTTP implements htt.Handler.
func (c *Canonical) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	renderer := c.r(w, req)
	if renderer == nil {
		log.Println(`unknown theme to use`)
		return
	}

	defer func() {
		if e := recover(); e != nil {
			if !renderer.Exception(w, req, e) {
				panic(e)
			}
		}
	}()

	path := req.URL.Path
	if !isValidPath(path) {
		renderer.Exception(w, req, errors.New(`invalid request path`))
		return
	}

	if req.Method == http.MethodGet || req.Method == http.MethodOptions {
		if regexpHome.MatchString(path) {
			if renderer.ProcessHomeQueries(w, req, req.URL.Query()) {
				return
			}
			renderer.QueryHome(w, req)
			c.mr.CountHome()
			c.mr.UserAgent(req.UserAgent())
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
			renderer.QueryByID(w, req, id)
			c.mr.CountPageView(id)
			c.mr.UserAgent(req.UserAgent())
			return
		}

		if regexpFile.MatchString(path) {
			matches := regexpFile.FindStringSubmatch(path)
			postID := utils.MustToInt64(matches[1])
			file := matches[2]
			renderer.QueryFile(w, req, postID, file)
			return
		}

		if regexpByTags.MatchString(path) {
			matches := regexpByTags.FindStringSubmatch(path)
			tags := strings.Split(matches[1], `+`)
			renderer.QueryByTags(w, req, tags)
			return
		}

		if renderer.QuerySpecial(w, req, path) {
			return
		}

		if regexpBySlug.MatchString(path) && isCategoryPath(path) {
			matches := regexpBySlug.FindStringSubmatch(path)
			tree := matches[1]
			slug := matches[2]
			id, err := renderer.QueryBySlug(w, req, tree, slug)
			if err == nil {
				c.mr.CountPageView(id)
				c.mr.UserAgent(req.UserAgent())
			}
			return
		}

		if regexpByPage.MatchString(path) && isCategoryPath(path) {
			matches := regexpByPage.FindStringSubmatch(path)
			parents := matches[1]
			if parents != "" {
				parents = parents[1:]
			}
			slug := matches[3]
			id, err := renderer.QueryByPage(w, req, parents, slug)
			if err == nil {
				c.mr.CountPageView(id)
				c.mr.UserAgent(req.UserAgent())
			}
			return
		}

		if strings.HasSuffix(path, "/") {
			w.WriteHeader(403)
			return
		}

		renderer.QueryStatic(w, req, path)
		return
	}

	http.NotFound(w, req)
}

// PostFromPath ...
func PostFromPath(path string) (int64, bool) {
	if regexpByID.MatchString(path) {
		matches := regexpByID.FindStringSubmatch(path)
		id := utils.MustToInt64(matches[1])
		return id, true
	}
	return 0, false
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
