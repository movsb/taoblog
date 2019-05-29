package blog

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
	TemplateCommon
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
	TemplateCommon
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
	Title          string
	PostCount      int64
	PageCount      int64
	CommentCount   int64
	LatestPosts    []*Post
	LatestComments []*Comment
}

// Archives ...
type Archives struct {
	Tags  []*models.TagWithCount
	Dates []*models.PostForDate
	Cats  template.HTML
	Title string
}

// QueryTags ...
type QueryTags struct {
	Tag   string
	Posts []*models.PostForArchive
}

// Blog ...
type Blog struct {
	base      string // base directory
	service   *service.Service
	templates *template.Template
	router    *gin.RouterGroup
	auth      *auth.Auth
	api       *gin.RouterGroup
	// dynamic files, rather than static files.
	// Not thread safe. Don't write after initializing.
	specialFiles map[string]func(c *gin.Context)
}

// NewBlog ...
func NewBlog(service *service.Service, auth *auth.Auth, router *gin.RouterGroup, api *gin.RouterGroup, base string) *Blog {
	b := &Blog{
		base:         base,
		service:      service,
		router:       router,
		auth:         auth,
		api:          api,
		specialFiles: make(map[string]func(c *gin.Context)),
	}
	b.loadTemplates()
	b.route()
	b.specialFiles = map[string]func(c *gin.Context){
		"/sitemap.xml": b.GetSitemap,
		"/rss":         b.GetRss,
		"/posts":       b.getPagePosts,
		"/all-posts.html": func(c *gin.Context) {
			c.Redirect(301, "/posts")
		},
		"/search": b.getPageSearch,
	}
	return b
}

func (b *Blog) route() {
	b.router.GET("/*path", func(c *gin.Context) {
		path := c.Param("path")
		b.Query(c, path)
	})

	posts := b.api.Group("/posts")
	posts.GET("/:name/comments", b.listPostComments)
	posts.POST("/:name/comments", b.createPostComment)

	tools := b.api.Group("/tools")
	tools.POST("/aes2htm", aes2htm)
}

func (b *Blog) render(w io.Writer, name string, data interface{}) {
	if err := b.templates.ExecuteTemplate(w, name, data); err != nil {
		panic(err)
	}
}

func (b *Blog) loadTemplates() {
	funcs := template.FuncMap{
		"get_config": func(name string) string {
			return b.service.GetStringOption(name)
		},
	}

	var tmpl *template.Template
	tmpl = template.New("blog").Funcs(funcs)
	path := filepath.Join(b.base, "templates", "*.html")
	tmpl, err := tmpl.ParseGlob(path)
	if err != nil {
		panic(err)
	}
	b.templates = tmpl
}

// Query ...
func (b *Blog) Query(c *gin.Context, path string) {
	defer func() {
		if e := recover(); e != nil {
			switch e.(type) {
			case *service.PostNotFoundError, *service.TagNotFoundError, *service.CategoryNotFoundError:
				c.Status(404)
				b.render(c.Writer, "404", nil)
				return
			case *PermDeniedError:
				c.Status(403)
				b.render(c.Writer, "403", nil)
				return
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
		if datetime.My2Gmt(p.Modified) == modified {
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
	user := b.auth.AuthCookie(c)
	tc := TemplateCommon{
		User: user,
	}
	header := &ThemeHeaderData{
		TemplateCommon: tc,
		Title:          "",
		Header: func() {
			b.render(c.Writer, "home_header", nil)
		},
	}

	footer := &ThemeFooterData{
		TemplateCommon: tc,
		Footer: func() {
			b.render(c.Writer, "home_footer", nil)
		},
	}

	home := &Home{
		PostCount:    b.service.GetDefaultIntegerOption("post_count", 0),
		PageCount:    b.service.GetDefaultIntegerOption("page_count", 0),
		CommentCount: b.service.GetDefaultIntegerOption("comment_count", 0),
	}
	home.LatestPosts = newPosts(b.service.MustListPosts(user.Context(nil),
		&protocols.ListPostsRequest{
			Fields:  "id,title,type",
			Limit:   20,
			OrderBy: "date DESC",
		}), b.service)
	home.LatestComments = newComments(b.service.ListComments(user.Context(nil),
		&protocols.ListCommentsRequest{
			Limit:    10,
			OrderBy:  "date DESC",
		}), b.service)

	b.render(c.Writer, "header", header)
	b.render(c.Writer, "home", home)
	b.render(c.Writer, "footer", footer)
}

func (b *Blog) queryByID(c *gin.Context, id int64) {
	post := b.service.GetPostByID(id)
	b.userMustCanSeePost(c, post)
	b.incView(post.ID)
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
	b.incView(post.ID)
	if b.handle304(c, post) {
		return
	}
	b.tempRenderPost(c, post)
}

func (b *Blog) queryByPage(c *gin.Context, parents string, slug string) {
	post := b.service.GetPostByPage(parents, slug)
	b.userMustCanSeePost(c, post)
	b.incView(post.ID)
	if b.handle304(c, post) {
		return
	}
	b.tempRenderPost(c, post)
}

func (b *Blog) tempRenderPost(c *gin.Context, p *protocols.Post) {
	post := newPost(p, b.service)
	post.RelatedPosts = b.service.GetRelatedPosts(post.ID)
	post.Tags = b.service.GetPostTags(post.ID)
	c.Header("Last-Modified", datetime.My2Gmt(post.Modified))
	w := c.Writer
	tc := TemplateCommon{
		User: b.auth.AuthCookie(c),
	}
	header := &ThemeHeaderData{
		TemplateCommon: tc,
		Title:          post.Title,
		Header: func() {
			b.render(w, "content_header", post)
			fmt.Fprint(w, post.CustomHeader())
		},
	}
	footer := &ThemeFooterData{
		TemplateCommon: tc,
		Footer: func() {
			b.render(w, "content_footer", post)
			fmt.Fprint(w, post.CustomFooter())
		},
	}
	b.render(w, "header", header)
	b.render(w, "content", post)
	b.render(w, "footer", footer)
}

func (b *Blog) queryByTags(c *gin.Context, tags string) {
	posts := b.service.GetPostsByTags(tags)
	in := QueryTags{Posts: posts, Tag: tags}
	b.render(c.Writer, "tags", &in)
}

func (b *Blog) userMustCanSeePost(c *gin.Context, post *protocols.Post) {
	user := b.auth.AuthCookie(c)
	if user.IsGuest() && post.Status != "public" {
		panic(&PermDeniedError{})
	}
}
