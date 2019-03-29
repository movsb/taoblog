package weekly

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"

	"github.com/movsb/taoblog/service"

	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/protocols"
)

// ThemeHeaderData ...
type ThemeHeaderData struct {
	Title  string
	Header func()
}

// HeaderHook ...
func (d *ThemeHeaderData) HeaderHook() string {
	if d.Header != nil {
		d.Header()
	}
	return ""
}

// ThemeFooterData ...
type ThemeFooterData struct {
	Footer func()
}

// FooterHook ...
func (d *ThemeFooterData) FooterHook() string {
	if d.Footer != nil {
		d.Footer()
	}
	return ""
}

// Home ...
type Home struct {
	Title       string
	LatestPosts []*Post
}

// Weekly ...
type Weekly struct {
	service   *service.Service
	templates *template.Template
	router    *gin.RouterGroup
	auth      *auth.Auth
	api       *gin.RouterGroup
	// dynamic files, rather than static files.
	// Not thread safe. Don't write after initializing.
	specialFiles map[string]func(c *gin.Context)
}

// NewWeekly ...
func NewWeekly(service *service.Service, auth *auth.Auth, router *gin.RouterGroup, api *gin.RouterGroup) *Weekly {
	w := &Weekly{
		service:      service,
		router:       router,
		auth:         auth,
		api:          api,
		specialFiles: make(map[string]func(c *gin.Context)),
	}
	w.loadTemplates()
	w.route()
	return w
}

func (w *Weekly) route() {
	w.router.GET("/*path", func(c *gin.Context) {
		path := c.Param("path")
		w.Query(c, path)
	})
}

func (w *Weekly) render(wr io.Writer, name string, data interface{}) {
	if err := w.templates.ExecuteTemplate(wr, name, data); err != nil {
		panic(err)
	}
}

func (w *Weekly) loadTemplates() {
	funcs := template.FuncMap{
		"get_config": func(name string) string {
			return w.service.GetStringOption(name)
		},
	}

	var tmpl *template.Template
	tmpl = template.New("weekly").Funcs(funcs)
	path := filepath.Join("weekly/templates", "*.html")
	tmpl, err := tmpl.ParseGlob(path)
	if err != nil {
		panic(err)
	}
	w.templates = tmpl
}

// Query ...
func (w *Weekly) Query(c *gin.Context, path string) {
	defer func() {
		if e := recover(); e != nil {
			switch e.(type) {
			case *service.PostNotFoundError, *service.TagNotFoundError, *service.CategoryNotFoundError:
				c.Status(404)
				w.render(c.Writer, "404", nil)
				return
			case *PermDeniedError:
				c.Status(403)
				w.render(c.Writer, "403", nil)
				return
			}
			panic(e)
		}
	}()

	if regexpHome.MatchString(path) {
		w.queryHome(c)
		return
	}
	if regexpBySlug.MatchString(path) {
		matches := regexpBySlug.FindStringSubmatch(path)
		slug := matches[1]
		w.queryBySlug(c, slug)
		return
	}
	c.File(filepath.Join("weekly/statics", path))
}

func (w *Weekly) handle304(c *gin.Context, p *protocols.Post) bool {
	if modified := c.GetHeader(`If-Modified-Since`); modified != "" {
		if datetime.My2Gmt(p.Modified) == modified {
			c.Status(304)
			return true
		}
	}
	return false
}

func (w *Weekly) queryHome(c *gin.Context) {
	user := w.auth.AuthCookie(c)
	header := &ThemeHeaderData{
		Title: "",
		Header: func() {
			w.render(c.Writer, "home_header", nil)
		},
	}

	footer := &ThemeFooterData{
		Footer: func() {
			w.render(c.Writer, "home_footer", nil)
		},
	}

	home := &Home{}
	home.LatestPosts = newPosts(w.service.MustListPosts(user.Context(nil),
		&protocols.ListPostsRequest{
			Fields:  "id,title,slug,type",
			Limit:   20,
			OrderBy: "date DESC",
		}), w.service)
	w.render(c.Writer, "header", header)
	w.render(c.Writer, "home", home)
	w.render(c.Writer, "footer", footer)
}

func (w *Weekly) queryBySlug(c *gin.Context, slug string) {
	post := w.service.GetPostBySlug("", slug)
	w.userMustCanSeePost(c, post)
	w.incView(post.ID)
	if w.handle304(c, post) {
		return
	}
	w.tempRenderPost(c, post)
}

func (w *Weekly) incView(id int64) {
	w.service.IncrementPostPageView(id)
}

func (w *Weekly) tempRenderPost(c *gin.Context, p *protocols.Post) {
	post := newPost(p, w.service)
	post.RelatedPosts = w.service.GetRelatedPosts(post.ID)
	post.Tags = w.service.GetPostTags(post.ID)
	c.Header("Last-Modified", datetime.My2Gmt(post.Modified))
	wr := c.Writer
	header := &ThemeHeaderData{
		Title: post.Title,
		Header: func() {
			w.render(wr, "content_header", post)
			fmt.Fprint(wr, post.CustomHeader())
		},
	}
	footer := &ThemeFooterData{
		Footer: func() {
			w.render(wr, "content_footer", post)
			fmt.Fprint(wr, post.CustomFooter())
		},
	}
	w.render(wr, "header", header)
	w.render(wr, "content", post)
	w.render(wr, "footer", footer)
}

func (w *Weekly) userMustCanSeePost(c *gin.Context, post *protocols.Post) {
	user := w.auth.AuthCookie(c)
	if user.IsGuest() && post.Status != "public" {
		panic(&PermDeniedError{})
	}
}
