package blog

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/metrics"
	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/themes/data"
)

var (
	regexpHome   = regexp.MustCompile(`^/$`)
	regexpByID   = regexp.MustCompile(`^/(\d+)/$`)
	regexpFile   = regexp.MustCompile(`^/(\d+)/(.+)$`)
	regexpBySlug = regexp.MustCompile(`^/(.+)/([^/]+)\.html$`)
	regexpByTags = regexp.MustCompile(`^/tags/(.*)$`)
	regexpByPage = regexp.MustCompile(`^((/[0-9a-zA-Z\-_]+)*)/([0-9a-zA-Z\-_]+)$`)
)

var nonCategoryNames = map[string]bool{
	"/admin/":   true,
	"/scripts/": true,
	"/images/":  true,
	"/sass/":    true,
	"/tags/":    true,
	"/plugins/": true,
	"/files/":   true,
}

// Blog ...
type Blog struct {
	cfg     *config.Config
	base    string // base directory
	service *service.Service
	router  *gin.RouterGroup
	auth    *auth.Auth
	api     *gin.RouterGroup
	metrics metrics.Server
	// dynamic files, rather than static files.
	// Not thread safe. Don't write after initializing.
	specialFiles map[string]func(c *gin.Context)

	globalTemplates *template.Template
	namedTemplates  map[string]*template.Template
}

// NewBlog ...
func NewBlog(cfg *config.Config, service *service.Service, auth *auth.Auth, router *gin.RouterGroup, api *gin.RouterGroup, metrics metrics.Server, base string) *Blog {
	b := &Blog{
		cfg:          cfg,
		base:         base,
		service:      service,
		router:       router,
		auth:         auth,
		api:          api,
		metrics:      metrics,
		specialFiles: make(map[string]func(c *gin.Context)),
	}
	b.loadTemplates()
	b.route()
	b.specialFiles = map[string]func(c *gin.Context){
		"/sitemap.xml": b.querySitemap,
		"/rss":         b.queryRss,
		"/search":      b.querySearch,
		"/posts":       b.queryPosts,
		"/all-posts.html": func(c *gin.Context) {
			c.Redirect(301, "/posts")
		},
	}
	return b
}

func (b *Blog) route() {
	b.router.GET("/*path", func(c *gin.Context) {
		path := c.Param("path")
		b.Query(c, path)
	})
}

func createMenus(items []config.MenuItem) string {
	menus := bytes.NewBuffer(nil)
	var genSubMenus func(buf *bytes.Buffer, items []config.MenuItem)
	a := func(item config.MenuItem) string {
		s := "<a"
		if item.Blank {
			s += " target=_blank"
		}
		if item.Link != "" {
			// TODO maybe error
			s += fmt.Sprintf(` href="%s"`, html.EscapeString(item.Link))
		}
		s += fmt.Sprintf(`>%s</a>`, html.EscapeString(item.Name))
		return s
	}
	genSubMenus = func(buf *bytes.Buffer, items []config.MenuItem) {
		if len(items) <= 0 {
			return
		}
		buf.WriteString("<ol>\n")
		for _, item := range items {
			if len(item.Items) == 0 {
				buf.WriteString(fmt.Sprintf("<li>%s</li>\n", a(item)))
			} else {
				buf.WriteString("<li>\n")
				buf.WriteString(fmt.Sprintf("%s\n", a(item)))
				genSubMenus(buf, item.Items)
				buf.WriteString("</li>\n")
			}
		}
		buf.WriteString("</ol>\n")
	}
	genSubMenus(menus, items)
	return menus.String()
}

func (b *Blog) loadTemplates() {
	menustr := createMenus(b.cfg.Menus)
	funcs := template.FuncMap{
		// https://github.com/golang/go/issues/14256
		"raw": func(s string) template.HTML {
			return template.HTML(s)
		},
		"get_config": func(name string) string {
			switch name {
			case `blog_name`:
				return b.service.Name()
			case `blog_desc`:
				return b.service.Description()
			default:
				panic(`cannot get this option`)
			}
		},
		"menus": func() template.HTML {
			return template.HTML(menustr)
		},
		"render": func(name string, data *data.Data) error {
			if t := data.Template.Lookup(name); t != nil {
				return t.Execute(data.Writer, data)
			}
			if t := b.globalTemplates.Lookup(name); t != nil {
				return t.Execute(data.Writer, data)
			}
			return nil
		},
	}

	b.globalTemplates = template.New(`global`).Funcs(funcs)
	b.namedTemplates = make(map[string]*template.Template)

	templateFiles, err := filepath.Glob(filepath.Join(b.base, `templates`, `*.html`))
	if err != nil {
		panic(err)
	}

	for _, path := range templateFiles {
		name := filepath.Base(path)
		if name[0] == '_' {
			b.globalTemplates.ParseFiles(path)
		} else {
			tmpl := template.Must(template.New(name).Funcs(funcs).ParseFiles(path))
			b.namedTemplates[tmpl.Name()] = tmpl
		}
	}
}

// Query ...
func (b *Blog) Query(c *gin.Context, path string) {
	defer func() {
		if e := recover(); e != nil {
			switch te := e.(type) {
			case *service.PostNotFoundError, *service.TagNotFoundError, *service.CategoryNotFoundError:
				c.Status(404)
				b.namedTemplates[`404.html`].Execute(c.Writer, nil)
				return
			case string: // hack hack
				switch te {
				case "403":
					c.Status(403)
					b.namedTemplates[`403.html`].Execute(c.Writer, nil)
					return
				default:
					panic(te)
				}
			}
			panic(e)
		}
	}()

	if regexpHome.MatchString(path) {
		if b.processHomeQueries(c) {
			return
		}
		b.queryHome(c)
		return
	}
	if regexpByID.MatchString(path) {
		matches := regexpByID.FindStringSubmatch(path)
		id := utils.MustToInt64(matches[1])
		b.queryByID(c, id)
		return
	}
	if regexpFile.MatchString(path) {
		matches := regexpFile.FindStringSubmatch(path)
		postID := utils.MustToInt64(matches[1])
		file := matches[2]
		b.queryByFile(c, postID, file)
		return
	}
	if regexpByTags.MatchString(path) {
		matches := regexpByTags.FindStringSubmatch(path)
		tags := matches[1]
		b.queryByTags(c, tags)
		return
	}
	if handler, ok := b.specialFiles[path]; ok {
		handler(c)
		return
	}
	if regexpBySlug.MatchString(path) && b.isCategoryPath(path) {
		matches := regexpBySlug.FindStringSubmatch(path)
		tree := matches[1]
		slug := matches[2]
		b.queryBySlug(c, tree, slug)
		return
	}
	if regexpByPage.MatchString(path) && b.isCategoryPath(path) {
		matches := regexpByPage.FindStringSubmatch(path)
		parents := matches[1]
		if parents != "" {
			parents = parents[1:]
		}
		slug := matches[3]
		b.queryByPage(c, parents, slug)
		return
	}
	if strings.HasSuffix(path, "/") {
		c.String(http.StatusForbidden, "403 Forbidden")
		return
	}
	c.File(filepath.Join(b.base, "statics", path))
}

func (b *Blog) handle304(c *gin.Context, p *protocols.Post) bool {
	if modified := c.GetHeader(`If-Modified-Since`); modified != "" {
		if datetime.My2Gmt(datetime.Proto2My(*p.Modified)) == modified {
			c.Status(304)
			return true
		}
	}
	return false
}

func (b *Blog) isCategoryPath(path string) bool {
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

func (b *Blog) processHomeQueries(c *gin.Context) bool {
	if p, ok := c.GetQuery("p"); ok && p != "" {
		c.Redirect(301, fmt.Sprintf("/%s/", p))
		return true
	}
	return false
}

func (b *Blog) queryHome(c *gin.Context) {
	d := data.NewDataForHome(b.cfg, b.auth.AuthCookie(c), b.service)
	t := b.namedTemplates[`home.html`]
	d.Template = t
	d.Writer = c.Writer
	if err := t.Execute(c.Writer, d); err != nil {
		panic(err)
	}
	b.metrics.CountPost(0, `首页`, c.ClientIP(), c.Request.UserAgent())
}

func (b *Blog) queryRss(c *gin.Context) {
	d := data.NewDataForRss(b.cfg, b.auth.AuthContext(c), b.service)
	t := b.namedTemplates[`rss.html`]
	d.Template = t
	d.Writer = c.Writer

	c.Header("Content-Type", "application/xml")

	// TODO turn on 304 or off from config.
	if modified := b.service.GetDefaultStringOption("last_post_time", ""); modified != "" {
		c.Header("Last-Modified", datetime.My2Gmt(modified))
	}

	c.Writer.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")

	if err := t.Execute(c.Writer, d); err != nil {
		panic(err)
	}

	// TODO metricss for rss crawler.
}

func (b *Blog) querySitemap(c *gin.Context) {
	d := data.NewDataForSitemap(b.cfg, b.auth.AuthCookie(c), b.service)
	t := b.namedTemplates[`sitemap.html`]
	d.Template = t
	d.Writer = c.Writer

	c.Header("Content-Type", "application/xml")

	// TODO turn on 304 or off from config.
	if modified := b.service.GetDefaultStringOption("last_post_time", ""); modified != "" {
		c.Header("Last-Modified", datetime.My2Gmt(modified))
	}

	c.Writer.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")

	if err := t.Execute(c.Writer, d); err != nil {
		panic(err)
	}

	// TODO metricss for rss crawler.
}

func (b *Blog) querySearch(c *gin.Context) {
	d := data.NewDataForSearch(b.cfg, b.auth.AuthCookie(c), b.service)
	t := b.namedTemplates[`search.html`]
	d.Template = t
	d.Writer = c.Writer

	// TODO turn on 304 or off from config.
	if modified := b.service.GetDefaultStringOption("last_post_time", ""); modified != "" {
		c.Header("Last-Modified", datetime.My2Gmt(modified))
	}

	if err := t.Execute(c.Writer, d); err != nil {
		panic(err)
	}

	// TODO metricss.
}

func (b *Blog) queryPosts(c *gin.Context) {
	d := data.NewDataForPosts(b.cfg, b.auth.AuthCookie(c), b.service, c)
	t := b.namedTemplates[`posts.html`]
	d.Template = t
	d.Writer = c.Writer

	// TODO turn on 304 or off from config.
	if modified := b.service.GetDefaultStringOption("last_post_time", ""); modified != "" {
		c.Header("Last-Modified", datetime.My2Gmt(modified))
	}

	if err := t.Execute(c.Writer, d); err != nil {
		panic(err)
	}

	// TODO metricss.
}

func (b *Blog) queryByID(c *gin.Context, id int64) {
	post := b.service.GetPostByID(id)
	b.userMustCanSeePost(c, post)
	b.incView(post.Id)
	b.metrics.CountPost(id, post.Title, c.ClientIP(), c.Request.UserAgent())
	if b.handle304(c, post) {
		return
	}
	b.tempRenderPost(c, post)
}

func (b *Blog) incView(id int64) {
	b.service.IncrementPostPageView(id)
}

func (b *Blog) queryBySlug(c *gin.Context, tree string, slug string) {
	post := b.service.GetPostBySlug(tree, slug)
	b.userMustCanSeePost(c, post)
	b.incView(post.Id)
	if b.handle304(c, post) {
		return
	}
	b.tempRenderPost(c, post)
}

func (b *Blog) queryByPage(c *gin.Context, parents string, slug string) {
	post := b.service.GetPostByPage(parents, slug)
	b.userMustCanSeePost(c, post)
	b.incView(post.Id)
	if b.handle304(c, post) {
		return
	}
	b.tempRenderPost(c, post)
}

func (b *Blog) tempRenderPost(c *gin.Context, p *protocols.Post) {
	d := data.NewDataForPost(b.cfg, b.auth.AuthCookie(c), b.service, p)
	t := b.namedTemplates[`post.html`]
	d.Template = t
	d.Writer = c.Writer
	c.Header("Last-Modified", datetime.My2Gmt(datetime.Proto2My(*d.Post.Post.Modified)))
	if err := t.Execute(c.Writer, d); err != nil {
		panic(err)
	}
}

func (b *Blog) queryByTags(c *gin.Context, tags string) {
	d := data.NewDataForTags(b.cfg, b.auth.AuthCookie(c), b.service, tags)
	t := b.namedTemplates[`tags.html`]
	d.Template = t
	d.Writer = c.Writer
	if err := t.Execute(c.Writer, d); err != nil {
		panic(err)
	}
}

func (b *Blog) queryByFile(c *gin.Context, postID int64, file string) {
	path := b.service.GetFile(postID, file)

	redir := true

	fileHost := b.cfg.Data.File.Mirror

	// remote isn't enabled, use local only
	if redir && fileHost == "" {
		redir = false
	}
	// when logged in, see the newest-uploaded file
	if redir && b.auth.AuthCookie(c).IsAdmin() {
		redir = false
	}
	// if no referer, don't let them know we're using file host
	if redir && c.GetHeader("Referer") == "" {
		redir = false
	}
	// if file isn't in local, we should redirect
	if !redir {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			redir = true
		}
	}

	if redir {
		remotePath := fileHost + "/" + path
		c.Redirect(307, remotePath)
	} else {
		c.File(path)
	}
}
func (b *Blog) userMustCanSeePost(c *gin.Context, post *protocols.Post) {
	user := b.auth.AuthCookie(c)
	if user.IsGuest() && post.Status != "public" {
		panic("403")
	}
}
