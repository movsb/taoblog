package main

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	regexpAdminLogin          = regexp.MustCompile(`^/login$`)
	regexpAdminLogout         = regexp.MustCompile(`^/logout$`)
	regexpAdminIndex          = regexp.MustCompile(`^/index$`)
	regexpAdminPostEdit       = regexp.MustCompile(`^/post-edit$`)
	regexpAdminTagManage      = regexp.MustCompile(`^/tag-manage$`)
	regexpAdminPostManage     = regexp.MustCompile(`^/post-manage$`)
	regexpAdminCategoryManage = regexp.MustCompile(`^/category-manage$`)
)

// AdminHeaderData ...
type AdminHeaderData struct {
	Title  string
	Header func()
}

// HeaderHook ...
func (d *AdminHeaderData) HeaderHook() string {
	if d.Header != nil {
		d.Header()
	}
	return ""
}

// AdminFooterData ...
type AdminFooterData struct {
	Footer func()
}

// FooterHook ...
func (d *AdminFooterData) FooterHook() string {
	if d.Footer != nil {
		d.Footer()
	}
	return ""
}

type LoginData struct {
	Redirect string
}

type AdminIndexData struct {
}

type AdminTagManageData struct {
	Tags []*TagWithCount
}

type AdminPostManageData struct {
}

type AdminCategoryManageData struct {
	CategoryJSON string
}

type AdminPostEditData struct {
	*Post
	New bool
}

func (d *AdminPostEditData) Link() string {
	if d.New {
		return fmt.Sprint(d.ID)
	}
	return fmt.Sprintf("https://%s/%d/", optmgr.GetDef(gdb, "home", ""), d.ID)
}

func (d *AdminPostEditData) TagStr() string {
	if d.New {
		return ""
	}
	return strings.Join(d.Tags, ",")
}

type Admin struct {
}

func NewAdmin() *Admin {
	a := &Admin{}
	return a
}

func (a *Admin) noCache(c *gin.Context) {
	c.Header("Cache-Control", "no-cache")
}

func (a *Admin) Query(c *gin.Context, path string) {
	if regexpAdminLogin.MatchString(path) {
		a.queryLogin(c)
		return
	}
	if regexpAdminLogout.MatchString(path) {
		a.queryLogout(c)
		return
	}
	if !auth(c, false) {
		c.Redirect(302, "/admin/login?redirect="+url.QueryEscape("/admin"+path))
		return
	}
	a.noCache(c)
	if regexpAdminIndex.MatchString(path) {
		a.queryIndex(c)
		return
	}
	if regexpAdminPostEdit.MatchString(path) {
		a.queryPostEdit(c)
		return
	}
	if regexpAdminTagManage.MatchString(path) {
		a.queryTagManage(c)
		return
	}
	if regexpAdminPostManage.MatchString(path) {
		a.queryPostManage(c)
		return
	}
	if regexpAdminCategoryManage.MatchString(path) {
		a.queryCategoryManage(c)
		return
	}
	c.File(filepath.Join(config.base, "admin", path))
}

func (a *Admin) Post(c *gin.Context, path string) {
	if regexpAdminLogin.MatchString(path) {
		a.postLogin(c)
		return
	}
	if !auth(c, false) {
		c.String(403, "")
		return
	}
}

func (a *Admin) queryLogin(c *gin.Context) {
	if auth(c, false) {
		c.Redirect(302, "/admin/index")
		return
	}
	redirect := c.Query("redirect")
	if redirect == "" || !strings.HasPrefix(redirect, "/") {
		redirect = "/admin/index"
	}

	d := LoginData{
		Redirect: redirect,
	}

	adminRender.Render(c.Writer, "login", &d)
}

func (a *Admin) queryLogout(c *gin.Context) {
	auther.DeleteCookie(c)
	c.Redirect(302, "/admin/login")
}

func (a *Admin) postLogin(c *gin.Context) {
	user := c.PostForm("user")
	password := c.PostForm("passwd")
	redirect := c.PostForm("redirect")
	if auther.Auth(user, password) {
		auther.MakeCookie(c)
		c.Redirect(302, redirect)
	} else {
		c.Redirect(302, c.Request.URL.String())
	}
}

func (a *Admin) queryIndex(c *gin.Context) {
	d := &AdminIndexData{}
	header := &AdminHeaderData{
		Title: "首页",
		Header: func() {
			adminRender.Render(c.Writer, "index_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			adminRender.Render(c.Writer, "index_footer", nil)
		},
	}
	adminRender.Render(c.Writer, "header", header)
	adminRender.Render(c.Writer, "index", d)
	adminRender.Render(c.Writer, "footer", footer)
}

func (a *Admin) queryTagManage(c *gin.Context) {
	tags, _ := tagmgr.ListTagsWithCount(gdb, 0, false)
	d := &AdminTagManageData{
		Tags: tags,
	}
	header := &AdminHeaderData{
		Title: "标签管理",
		Header: func() {
			adminRender.Render(c.Writer, "tag_manage_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			adminRender.Render(c.Writer, "tag_manage_footer", nil)
		},
	}
	adminRender.Render(c.Writer, "header", header)
	adminRender.Render(c.Writer, "tag_manage", d)
	adminRender.Render(c.Writer, "footer", footer)
}

func (a *Admin) queryPostManage(c *gin.Context) {
	d := &AdminPostManageData{}
	header := &AdminHeaderData{
		Title: "文章管理",
		Header: func() {
			adminRender.Render(c.Writer, "post_manage_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			adminRender.Render(c.Writer, "post_manage_footer", nil)
		},
	}
	adminRender.Render(c.Writer, "header", header)
	adminRender.Render(c.Writer, "post_manage", d)
	adminRender.Render(c.Writer, "footer", footer)
}

func (a *Admin) queryCategoryManage(c *gin.Context) {
	d := &AdminCategoryManageData{}
	header := &AdminHeaderData{
		Title: "分类管理",
		Header: func() {
			adminRender.Render(c.Writer, "category_manage_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			adminRender.Render(c.Writer, "category_manage_footer", nil)
		},
	}
	adminRender.Render(c.Writer, "header", header)
	adminRender.Render(c.Writer, "category_manage", d)
	adminRender.Render(c.Writer, "footer", footer)
}

func (a *Admin) queryPostEdit(c *gin.Context) {
	p := &Post{}
	d := AdminPostEditData{
		New:  true,
		Post: p,
	}
	header := &AdminHeaderData{
		Title: "文章编辑",
		Header: func() {
			adminRender.Render(c.Writer, "post_edit_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			adminRender.Render(c.Writer, "post_edit_footer", nil)
		},
	}
	adminRender.Render(c.Writer, "header", header)
	adminRender.Render(c.Writer, "post_edit", &d)
	adminRender.Render(c.Writer, "footer", footer)
}
