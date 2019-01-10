package front

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/movsb/taoblog/modules/utils"

	"github.com/movsb/taoblog/service"

	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
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

type Home struct {
	Title          string
	PostCount      int64
	PageCount      int64
	CommentCount   int64
	LatestPosts    []*Post
	LatestComments []*Comment
}

type Archives struct {
	Tags  []*models.TagWithCount
	Dates []*models.PostForDate
	Cats  template.HTML
	Title string
}

type QueryTags struct {
	Tag   string
	Posts []*models.PostForArchive
}

type Front struct {
	server    *service.ImplServer
	templates *template.Template
	router    *gin.RouterGroup
	auth      *auth.Auth
	api       *gin.RouterGroup
	// dynamic files, rather than static files.
	// Not thread safe. Don't write after initializing.
	specialFiles map[string]func(c *gin.Context)
}

func NewFront(server *service.ImplServer, auth *auth.Auth, router *gin.RouterGroup, api *gin.RouterGroup) *Front {
	f := &Front{
		server:       server,
		router:       router,
		auth:         auth,
		api:          api,
		specialFiles: make(map[string]func(c *gin.Context)),
	}
	f.loadTemplates()
	f.route()
	f.specialFiles = map[string]func(c *gin.Context){
		"/sitemap.xml": f.GetSitemap,
		"/rss":         f.GetRss,
		"/posts":       f.GetPagePosts,
		"/all-posts.html": func(c *gin.Context) {
			c.Redirect(301, "/posts")
		},
	}
	return f
}

func (f *Front) route() {
	f.router.GET("/*path", func(c *gin.Context) {
		path := c.Param("path")
		f.Query(c, path)
	})

	posts := f.api.Group("/posts")
	posts.GET("/:name/comments", f.listPostComments)
	posts.POST("/:name/comments", f.createPostComment)

	tools := f.api.Group("/tools")
	tools.POST("/aes2htm", aes2htm)
}

func (f *Front) render(w io.Writer, name string, data interface{}) {
	if err := f.templates.ExecuteTemplate(w, name, data); err != nil {
		panic(err)
	}
}

func (f *Front) loadTemplates() {
	funcs := template.FuncMap{
		"get_config": func(name string) string {
			return f.server.GetStringOption(name)
		},
	}

	var tmpl *template.Template
	tmpl = template.New("front").Funcs(funcs)
	path := filepath.Join("front/templates", "*.html")
	tmpl, err := tmpl.ParseGlob(path)
	if err != nil {
		panic(err)
	}
	f.templates = tmpl
}

func (f *Front) Query(c *gin.Context, path string) {
	defer func() {
		if e := recover(); e != nil {
			switch e.(type) {
			case *service.PostNotFoundError, *service.TagNotFoundError, *service.CategoryNotFoundError:
				c.Status(404)
				f.render(c.Writer, "404", nil)
				return
			}
			panic(e)
		}
	}()

	if regexpHome.MatchString(path) {
		if f.processHomeQueries(c) {
			return
		}
		f.queryHome(c)
		return
	}
	if regexpByID.MatchString(path) {
		matches := regexpByID.FindStringSubmatch(path)
		id := utils.MustToInt64(matches[1])
		f.queryByID(c, id)
		return
	}
	if regexpFile.MatchString(path) {
		matches := regexpFile.FindStringSubmatch(path)
		postID := utils.MustToInt64(matches[1])
		file := matches[2]
		f.queryByFile(c, postID, file)
		return
	}
	if regexpByTags.MatchString(path) {
		matches := regexpByTags.FindStringSubmatch(path)
		tags := matches[1]
		f.queryByTags(c, tags)
		return
	}
	if handler, ok := f.specialFiles[path]; ok {
		handler(c)
		return
	}
	if regexpBySlug.MatchString(path) && f.isCategoryPath(path) {
		matches := regexpBySlug.FindStringSubmatch(path)
		tree := matches[1]
		slug := matches[2]
		f.queryBySlug(c, tree, slug)
		return
	}
	if regexpByPage.MatchString(path) && f.isCategoryPath(path) {
		matches := regexpByPage.FindStringSubmatch(path)
		parents := matches[1]
		if parents != "" {
			parents = parents[1:]
		}
		slug := matches[3]
		f.queryByPage(c, parents, slug)
		return
	}
	if strings.HasSuffix(path, "/") {
		c.String(http.StatusForbidden, "403 Forbidden")
		return
	}
	c.File(filepath.Join("front/statics", path))
}

func (f *Front) handle304(c *gin.Context, p *protocols.Post) bool {
	if modified := c.GetHeader(`If-Modified-Since`); modified != "" {
		ht := datetime.Http2Time(modified)
		pt := datetime.My2Time(p.Modified)
		if ht.Equal(pt) {
			c.Status(304)
			return true
		}
	}
	return false
}

func (f *Front) isCategoryPath(path string) bool {
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

func (f *Front) processHomeQueries(c *gin.Context) bool {
	if p, ok := c.GetQuery("p"); ok && p != "" {
		c.Redirect(301, fmt.Sprintf("/%s/", p))
		return true
	}
	return false
}

func (f *Front) queryHome(c *gin.Context) {
	header := &ThemeHeaderData{
		Title: "",
		Header: func() {
			f.render(c.Writer, "home_header", nil)
		},
	}

	footer := &ThemeFooterData{
		Footer: func() {
			f.render(c.Writer, "home_footer", nil)
		},
	}

	home := &Home{
		PostCount:    f.server.GetDefaultIntegerOption("post_count", 0),
		PageCount:    f.server.GetDefaultIntegerOption("page_count", 0),
		CommentCount: f.server.GetDefaultIntegerOption("comment_count", 0),
	}
	home.LatestPosts = newPosts(f.server.MustListPosts(&protocols.ListPostsRequest{
		Fields:  "id,title,type",
		Limit:   20,
		OrderBy: "date DESC",
	}), f.server)
	home.LatestComments = newComments(f.server.ListComments(&protocols.ListCommentsRequest{
		Ancestor: -1,
		Limit:    10,
		OrderBy:  "date DESC",
	}), f.server)

	f.render(c.Writer, "header", header)
	f.render(c.Writer, "home", home)
	f.render(c.Writer, "footer", footer)
}

func (f *Front) queryByID(c *gin.Context, id int64) {
	post := f.server.GetPostByID(id)
	f.incView(post.ID)
	if f.handle304(c, post) {
		return
	}
	f.tempRenderPost(c, post)
}

func (f *Front) incView(id int64) {
	f.server.IncrementPostPageView(id)
}

func (f *Front) queryBySlug(c *gin.Context, tree string, slug string) {
	post := f.server.GetPostBySlug(tree, slug)
	f.incView(post.ID)
	if f.handle304(c, post) {
		return
	}
	f.tempRenderPost(c, post)
}

func (f *Front) queryByPage(c *gin.Context, parents string, slug string) {
	post := f.server.GetPostByPage(parents, slug)
	f.incView(post.ID)
	if f.handle304(c, post) {
		return
	}
	f.tempRenderPost(c, post)
}

func (f *Front) tempRenderPost(c *gin.Context, p *protocols.Post) {
	post := newPost(p, f.server)
	post.RelatedPosts = f.server.GetRelatedPosts(post.ID)
	post.Tags = f.server.GetPostTags(post.ID)
	c.Header("Last-Modified", datetime.My2Gmt(post.Modified))
	w := c.Writer
	header := &ThemeHeaderData{
		Title: post.Title,
		Header: func() {
			f.render(w, "content_header", post)
			fmt.Fprint(w, post.CustomHeader())
		},
	}
	footer := &ThemeFooterData{
		Footer: func() {
			f.render(w, "content_footer", post)
			fmt.Fprint(w, post.CustomFooter())
		},
	}
	f.render(w, "header", header)
	f.render(w, "content", post)
	f.render(w, "footer", footer)
}

func (f *Front) queryByTags(c *gin.Context, tags string) {
	posts := f.server.GetPostsByTags(tags)
	in := QueryTags{Posts: posts, Tag: tags}
	f.render(c.Writer, "tags", &in)
}
