package main

import (
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

var (
	regexpByID       = regexp.MustCompile(`/(\d+)/$`)
	regexpBySlug     = regexp.MustCompile(`^/(.+)/([^/]+)\.html$`)
	regexpByArchives = regexp.MustCompile(`^/archives$`)
	regexpByPage     = regexp.MustCompile(`^((/[0-9a-zA-Z\-_]+)*)/([0-9a-zA-Z\-_]+)$`)
	regexpHome       = regexp.MustCompile(`^/$`)
)

type Home struct {
	Title          string
	PostCount      int64
	PageCount      int64
	CommentCount   int64
	LatestPosts    []*PostForLatest
	LatestComments []*Comment
}

func (h *Home) PageType() string {
	return "home"
}

type Blog struct {
}

func NewBlog() *Blog {
	b := &Blog{}
	return b
}

func (b *Blog) Query(c *gin.Context, path string) {
	if regexpHome.MatchString(path) {
		b.queryHome(c)
		return
	}
	if regexpByID.MatchString(path) {
		matches := regexpByID.FindStringSubmatch(path)
		id, _ := strconv.ParseInt(matches[1], 10, 64)
		b.queryByID(c, id)
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
	if regexpByArchives.MatchString(path) {
		b.queryByArchives(c)
		return
	}
	c.File(filepath.Join(config.base, path))
}

func (b *Blog) queryHome(c *gin.Context) {
	home := &Home{}
	home.Title = "首页"
	home.PostCount = optmgr.GetDefInt(gdb, "post_count", 0)
	home.PageCount = optmgr.GetDefInt(gdb, "page_count", 0)
	home.CommentCount = optmgr.GetDefInt(gdb, "comment_count", 0)
	home.LatestPosts, _ = postmgr.GetLatest(gdb, 20)
	home.LatestComments, _ = cmtmgr.GetRecentComments(gdb, 10)
	renderer.RenderHome(c, home)
}

func (b *Blog) queryByID(c *gin.Context, id int64) {
	post, err := postmgr.GetPostByID(gdb, id, "")
	if err != nil {
		EndReq(c, err, post)
		return
	}
	renderer.RenderPost(c, post)
}

func (b *Blog) queryBySlug(c *gin.Context, tree string, slug string) {
	post, err := postmgr.GetPostBySlug(gdb, tree, slug, "", false)
	if err != nil {
		EndReq(c, err, post)
		return
	}
	renderer.RenderPost(c, post)
}

func (b *Blog) queryByPage(c *gin.Context, parents string, slug string) {
	post, err := postmgr.GetPostBySlug(gdb, parents, slug, "", true)
	if err != nil {
		EndReq(c, err, post)
		return
	}
	renderer.RenderPost(c, post)
}

func (b *Blog) queryByArchives(c *gin.Context) {

}
