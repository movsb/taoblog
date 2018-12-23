package admin

import (
	"fmt"
	"html/template"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/movsb/taoblog/protocols"
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
	Tags []*protocols.TagWithCount
}

type AdminPostManageData struct {
}

type AdminCategoryManageData struct {
	CategoryJSON string
}

type AdminPostEditData struct {
	*protocols.Post
	New bool
}

func (d *AdminPostEditData) Link() string {
	if d.New {
		return fmt.Sprint(d.ID)
	}
	// TODO with home
	return fmt.Sprintf("/%d/", d.ID)
}

func (d *AdminPostEditData) TagStr() string {
	if d.New {
		return ""
	}
	return strings.Join(d.Tags, ",")
}

type Admin struct {
	server    protocols.IServer
	templates *template.Template
	router    *gin.RouterGroup
}

func NewAdmin(server protocols.IServer, router *gin.RouterGroup) *Admin {
	a := &Admin{
		server: server,
		router: router,
	}
	a.loadTemplates()
	a.route()
	return a
}

func (a *Admin) route() {
	g := a.router.Group("/admin")
	g.GET("/*path", func(c *gin.Context) {
		path := c.Param("path")
		switch path {
		case "", "/":
			c.Redirect(302, "/admin/login")
			return
		}
		a.Query(c, path)
	})
	g.POST("/*path", func(c *gin.Context) {
		path := c.Param("path")
		a.Post(c, path)
	})
}

func (a *Admin) auth(c *gin.Context) bool {
	return a.server.Auth(&protocols.AuthRequest{c}).Success
}

func (a *Admin) noCache(c *gin.Context) {
	c.Header("Cache-Control", "no-cache")
}

func (a *Admin) render(w io.Writer, name string, data interface{}) {
	if err := a.templates.ExecuteTemplate(w, name, data); err != nil {
		panic(err)
	}
}

func (a *Admin) loadTemplates() {
	funcs := template.FuncMap{
		"get_config": func(name string) string {
			return a.server.GetOption(&protocols.GetOptionRequest{
				Name:    name,
				Default: true,
				Value:   "",
			}).Value
		},
	}

	var tmpl *template.Template
	tmpl = template.New("admin").Funcs(funcs)
	path := filepath.Join(os.Getenv("BASE"), "admin", "templates", "*.html")
	tmpl, err := a.templates.ParseGlob(path)
	if err != nil {
		panic(err)
	}
	a.templates = tmpl
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
	if !a.auth(c) {
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
	c.File(filepath.Join(os.Getenv("BASE"), "admin/templates", path))
}

func (a *Admin) Post(c *gin.Context, path string) {
	if regexpAdminLogin.MatchString(path) {
		a.postLogin(c)
		return
	}
	if !a.auth(c) {
		c.String(403, "")
		return
	}
}

func (a *Admin) queryLogin(c *gin.Context) {
	if a.auth(c) {
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

	a.render(c.Writer, "login", &d)
}

func (a *Admin) queryLogout(c *gin.Context) {
	c.SetCookie("login", "", -1, "/", "", true, true)
	c.Redirect(302, "/admin/login")
}

func (a *Admin) postLogin(c *gin.Context) {
	user := c.PostForm("user")
	password := c.PostForm("passwd")
	redirect := c.PostForm("redirect")
	auth := a.server.AuthLogin(&protocols.AuthLoginRequest{
		UserAgent: c.GetHeader("User-Agent"),
		Username:  user,
		Password:  password,
	})
	if auth.Success {
		c.SetCookie("login", auth.Cookie, 0, "/", "", true, true)
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
			a.render(c.Writer, "index_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			a.render(c.Writer, "index_footer", nil)
		},
	}
	a.render(c.Writer, "header", header)
	a.render(c.Writer, "index", d)
	a.render(c.Writer, "footer", footer)
}

func (a *Admin) queryTagManage(c *gin.Context) {
	d := &AdminTagManageData{
		Tags: a.server.ListTagsWithCount(&protocols.ListTagsWithCountRequest{
			Limit:      0,
			MergeAlias: false,
		}).Tags,
	}
	header := &AdminHeaderData{
		Title: "标签管理",
		Header: func() {
			a.render(c.Writer, "tag_manage_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			a.render(c.Writer, "tag_manage_footer", nil)
		},
	}
	a.render(c.Writer, "header", header)
	a.render(c.Writer, "tag_manage", d)
	a.render(c.Writer, "footer", footer)
}

func (a *Admin) queryPostManage(c *gin.Context) {
	d := &AdminPostManageData{}
	header := &AdminHeaderData{
		Title: "文章管理",
		Header: func() {
			a.render(c.Writer, "post_manage_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			a.render(c.Writer, "post_manage_footer", nil)
		},
	}
	a.render(c.Writer, "header", header)
	a.render(c.Writer, "post_manage", d)
	a.render(c.Writer, "footer", footer)
}

func (a *Admin) queryCategoryManage(c *gin.Context) {
	d := &AdminCategoryManageData{}
	header := &AdminHeaderData{
		Title: "分类管理",
		Header: func() {
			a.render(c.Writer, "category_manage_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			a.render(c.Writer, "category_manage_footer", nil)
		},
	}
	a.render(c.Writer, "header", header)
	a.render(c.Writer, "category_manage", d)
	a.render(c.Writer, "footer", footer)
}

func (a *Admin) queryPostEdit(c *gin.Context) {
	p := &protocols.Post{}
	d := AdminPostEditData{
		New:  true,
		Post: p,
	}
	header := &AdminHeaderData{
		Title: "文章编辑",
		Header: func() {
			a.render(c.Writer, "post_edit_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			a.render(c.Writer, "post_edit_footer", nil)
		},
	}
	a.render(c.Writer, "header", header)
	a.render(c.Writer, "post_edit", &d)
	a.render(c.Writer, "footer", footer)
}
