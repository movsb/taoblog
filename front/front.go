package front

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/movsb/taoblog/service"

	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/service/models"
)

var (
	regexpHome       = regexp.MustCompile(`^/$`)
	regexpByID       = regexp.MustCompile(`/(\d+)/$`)
	regexpBySlug     = regexp.MustCompile(`^/(.+)/([^/]+)\.html$`)
	regexpByTags     = regexp.MustCompile(`^/tags/(.*)$`)
	regexpByArchives = regexp.MustCompile(`^/archives$`)
	regexpByPage     = regexp.MustCompile(`^((/[0-9a-zA-Z\-_]+)*)/([0-9a-zA-Z\-_]+)$`)
)

var nonCategoryNames = map[string]bool{
	"/admin/":    true,
	"/emotions/": true,
	"/scripts/":  true,
	"/images/":   true,
	"/sass/":     true,
	"/tags/":     true,
	"/plugins/":  true,
}

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
	LatestPosts    []*models.Post
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
}

func NewFront(server *service.ImplServer, router *gin.RouterGroup) *Front {
	f := &Front{
		server: server,
		router: router,
	}
	f.loadTemplates()
	f.route()
	return f
}

func (f *Front) route() {
	g := f.router.Group("/blog")
	g.GET("/*path", func(c *gin.Context) {
		path := c.Param("path")
		f.Query(c, path)
	})
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
	path := filepath.Join(os.Getenv("BASE"), "front/templates", "*.html")
	tmpl, err := tmpl.ParseGlob(path)
	if err != nil {
		panic(err)
	}
	f.templates = tmpl
}

func (f *Front) Query(c *gin.Context, path string) {
	if regexpHome.MatchString(path) {
		if f.processHomeQueries(c) {
			return
		}
		f.queryHome(c)
		return
	}
	if regexpByID.MatchString(path) {
		matches := regexpByID.FindStringSubmatch(path)
		id, _ := strconv.ParseInt(matches[1], 10, 64)
		f.queryByID(c, id)
		return
	}
	if regexpByTags.MatchString(path) {
		//matches := regexpByTags.FindStringSubmatch(path)
		//tags := matches[1]
		// TODO
		//f.queryByTags(c, tags)
		return
	}
	if regexpByArchives.MatchString(path) {
		// TODO
		//f.queryByArchives(c)
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
	c.File(filepath.Join(os.Getenv("BASE"), "front/html", path))
}

func (f *Front) handle304(c *gin.Context, p *models.Post) bool {
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
	home.LatestPosts = f.server.ListPosts(&service.ListPostsRequest{
		Fields:  "id,title,type",
		Limit:   20,
		OrderBy: "date DESC",
	})
	home.LatestComments = newComments(f.server.ListComments(&service.ListCommentsRequest{
		Parent:  0,
		Limit:   10,
		OrderBy: "date DESC",
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
	f.server.IncrementPostView(id)
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

func (f *Front) tempRenderPost(c *gin.Context, p *models.Post) {
	post := newPost(p, f.server)
	post.RelatedPosts = f.server.GetRelatedPosts(post.ID)
	post.Tags = f.server.GetPostTags(post.ID)
	c.Header("Last-Modified", datetime.My2Gmt(post.Modified))
	w := c.Writer
	header := &ThemeHeaderData{
		Title: post.Title,
		Header: func() {
			f.render(w, "content_header", post)
			// TODO
			//fmt.Fprint(w, post.CustomHeader())
		},
	}
	footer := &ThemeFooterData{
		Footer: func() {
			f.render(w, "content_footer", post)
			//fmt.Fprint(w, post.CustomFooter())
		},
	}
	f.render(w, "header", header)
	f.render(w, "content", post)
	f.render(w, "footer", footer)
}

// TODO
/*
func (f *Front) queryByTags(c *gin.Context, tags string) {
	posts, err := postmgr.GetPostsByTags(gdb, tags)
	if err != nil {
		EndReq(c, err, posts)
		return
	}
	f.render(c.Writer, "tags", &QueryTags{Posts: posts, Tag: tags})
}

func (f *Front) queryByArchives(c *gin.Context) {
	tags := f.server.ListTagsWithCount(&models.ListTagsWithCountRequest{
		Limit:      50,
		MergeAlias: true,
	}).Tags
	posts, _ := postmgr.GetDateArchives(gdb)

	cats, _ := catmgr.GetTree(gdb)
	postCounts, _ := catmgr.GetCountOfCategoriesAll(gdb)

	var fn func([]*Category) (string, int64)
	fn = func(cats []*Category) (string, int64) {
		s := ""
		n := int64(0)
		for _, cat := range cats {
			postCount := postCounts[cat.ID]
			s1 := fmt.Sprintf(`<li data-cid=%d class=folder><i class="folder-name fa fa-folder-o"></i><span class="folder-name">%s(`, cat.ID, cat.Name)
			s2 := `)</span><ul>`
			s3, childCount := fn(cat.Children)
			s4 := `</ul></li>`
			c := fmt.Sprint(postCount)
			if len(cat.Children) > 0 {
				c += fmt.Sprintf("/%d", postCount+childCount)
			}
			s += s1 + c + s2 + s3 + s4
			n += postCount + childCount
		}
		return s, n
	}

	catstr, _ := fn(cats)

	header := &ThemeHeaderData{
		Title: "文章归档",
		Header: func() {
			f.render(c.Writer, "archives_header", nil)
		},
	}

	footer := &ThemeFooterData{
		Footer: func() {
			f.render(c.Writer, "archives_footer", nil)
		},
	}

	a := &Archives{
		Title: "文章归档",
		Tags:  tags,
		Dates: posts,
		Cats:  template.HTML(catstr),
	}

	f.render(c.Writer, "header", header)
	f.render(c.Writer, "archives", a)
	f.render(c.Writer, "footer", footer)
}
*/
func (f *Front) postNotFound(c *gin.Context) {
	c.Status(404)
	f.render(c.Writer, "404", nil)
}
