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

type LoginData struct {
	Redirect string
}

func (d *LoginData) PageType() string {
	return "admin_login"
}

type AdminIndexData struct {
}

func (d *AdminIndexData) PageType() string {
	return "admin_index"
}

type AdminTagManageData struct {
	Tags []*TagWithCount
}

func (d *AdminTagManageData) PageType() string {
	return "admin_tag_manage"
}

type AdminPostManageData struct {
}

func (d *AdminPostManageData) PageType() string {
	return "admin_post_manage"
}

type AdminCategoryManageData struct {
	CategoryJSON string
}

func (d *AdminCategoryManageData) PageType() string {
	return "admin_category_manage"
}

type AdminPostEditData struct {
	*Post
	New bool
}

func (d *AdminPostEditData) PageType() string {
	return "admin_post_edit"
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

	renderer.Render(c, "admin_login", &d)
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
	renderer.Render(c, "admin_index", d)
}

func (a *Admin) queryTagManage(c *gin.Context) {
	tags, _ := tagmgr.ListTagsWithCount(gdb, 0, false)
	d := &AdminTagManageData{
		Tags: tags,
	}
	renderer.Render(c, "tag_manage", d)
}

func (a *Admin) queryPostManage(c *gin.Context) {
	d := &AdminPostManageData{}
	renderer.Render(c, "admin_post_manage", d)
}

func (a *Admin) queryCategoryManage(c *gin.Context) {
	d := &AdminCategoryManageData{}
	renderer.Render(c, "admin_category_manage", d)
}

func (a *Admin) queryPostEdit(c *gin.Context) {
	p := &Post{}
	d := AdminPostEditData{
		Post: p,
	}
	renderer.Render(c, "admin_header", &d)
	renderer.Render(c, "admin_post_edit", &d)
	renderer.Render(c, "admin_footer", &d)
}