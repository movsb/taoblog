package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	regexpHome       = regexp.MustCompile(`^/$`)
	regexpByID       = regexp.MustCompile(`/(\d+)/$`)
	regexpBySlug     = regexp.MustCompile(`^/(.+)/([^/]+)\.html$`)
	regexpByTags     = regexp.MustCompile(`^/tags/(.*)$`)
	regexpByArchives = regexp.MustCompile(`^/archives$`)
	regexpByPage     = regexp.MustCompile(`^((/[0-9a-zA-Z\-_]+)*)/([0-9a-zA-Z\-_]+)$`)
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
	LatestPosts    []*PostForLatest
	LatestComments []*Comment
}

type Archives struct {
	Tags  []*TagWithCount
	Dates []*PostForDate
	Cats  template.HTML
	Title string
}

type QueryTags struct {
	Tag   string
	Posts []*PostForArchiveQuery
}

type Blog struct {
}

func NewBlog() *Blog {
	b := &Blog{}
	return b
}

func (b *Blog) Query(c *gin.Context, path string) {
	if regexpHome.MatchString(path) {
		if b.processHomeQueries(c) {
			return
		}
		b.queryHome(c)
		return
	}
	if regexpByID.MatchString(path) {
		matches := regexpByID.FindStringSubmatch(path)
		id, _ := strconv.ParseInt(matches[1], 10, 64)
		b.queryByID(c, id)
		return
	}
	if regexpByTags.MatchString(path) {
		matches := regexpByTags.FindStringSubmatch(path)
		tags := matches[1]
		b.queryByTags(c, tags)
		return
	}
	if regexpByArchives.MatchString(path) {
		b.queryByArchives(c)
		return
	}
	if regexpBySlug.MatchString(path) {
		matches := regexpBySlug.FindStringSubmatch(path)
		tree := matches[1]
		slug := matches[2]
		b.queryBySlug(c, tree, slug)
		return
	}
	if regexpByPage.MatchString(path) {
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
	c.File(filepath.Join(config.base, path))
}

func (b *Blog) processHomeQueries(c *gin.Context) bool {
	if p, ok := c.GetQuery("p"); ok {
		c.Redirect(301, fmt.Sprintf("/%s/", p))
		return true
	}
	return false
}

func (b *Blog) queryHome(c *gin.Context) {
	header := &ThemeHeaderData{
		Title: "",
		Header: func() {
			themeRender.Render(c.Writer, "home_header", nil)
		},
	}

	footer := &ThemeFooterData{
		Footer: func() {
			themeRender.Render(c.Writer, "home_footer", nil)
		},
	}

	home := &Home{}
	home.PostCount = optmgr.GetDefInt(gdb, "post_count", 0)
	home.PageCount = optmgr.GetDefInt(gdb, "page_count", 0)
	home.CommentCount = optmgr.GetDefInt(gdb, "comment_count", 0)
	home.LatestPosts, _ = postmgr.GetLatest(gdb, 20)
	home.LatestComments, _ = cmtmgr.GetRecentComments(gdb, 10)

	themeRender.Render(c.Writer, "header", header)
	themeRender.Render(c.Writer, "home", home)
	themeRender.Render(c.Writer, "footer", footer)
}

func (b *Blog) queryByID(c *gin.Context, id int64) {
	post, err := postmgr.GetPostByID(gdb, id, "")
	if err != nil {
		EndReq(c, err, post)
		return
	}
	postmgr.IncrementPageView(gdb, post.ID)
	tempRenderPost(c.Writer, post)
}

func (b *Blog) queryBySlug(c *gin.Context, tree string, slug string) {
	post, err := postmgr.GetPostBySlug(gdb, tree, slug, "", false)
	if err != nil {
		EndReq(c, err, post)
		return
	}
	postmgr.IncrementPageView(gdb, post.ID)
	tempRenderPost(c.Writer, post)
}

func (b *Blog) queryByPage(c *gin.Context, parents string, slug string) {
	post, err := postmgr.GetPostBySlug(gdb, parents, slug, "", true)
	if err != nil {
		EndReq(c, err, post)
		return
	}
	postmgr.IncrementPageView(gdb, post.ID)
	tempRenderPost(c.Writer, post)
}

func tempRenderPost(w io.Writer, post *Post) {
	header := &ThemeHeaderData{
		Title: post.Title,
		Header: func() {
			themeRender.Render(w, "content_header", post)
		},
	}
	footer := &ThemeFooterData{
		Footer: func() {
			themeRender.Render(w, "content_footer", post)
		},
	}
	themeRender.Render(w, "header", header)
	themeRender.Render(w, "content", post)
	themeRender.Render(w, "footer", footer)
}

func (b *Blog) queryByTags(c *gin.Context, tags string) {
	posts, err := postmgr.GetPostsByTags(gdb, tags)
	if err != nil {
		EndReq(c, err, posts)
		return
	}
	themeRender.Render(c.Writer, "tags", &QueryTags{Posts: posts, Tag: tags})
}

func (b *Blog) queryByArchives(c *gin.Context) {
	tags, _ := tagmgr.ListTagsWithCount(gdb, 50, true)
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
			themeRender.Render(c.Writer, "archives_header", nil)
		},
	}

	footer := &ThemeFooterData{
		Footer: func() {
			themeRender.Render(c.Writer, "archives_footer", nil)
		},
	}

	a := &Archives{
		Title: "文章归档",
		Tags:  tags,
		Dates: posts,
		Cats:  template.HTML(catstr),
	}

	themeRender.Render(c.Writer, "header", header)
	themeRender.Render(c.Writer, "archives", a)
	themeRender.Render(c.Writer, "footer", footer)
}
