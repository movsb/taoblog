package main

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	regexpAdminLogin = regexp.MustCompile(`^/login$`)
	regexpAdminIndex = regexp.MustCompile(`^/index$`)
)

type LoginData struct {
	Redirect string
}

func (d *LoginData) PageType() string {
	return "login"
}

type Admin struct {
}

func NewAdmin() *Admin {
	a := &Admin{}
	return a
}

func (a *Admin) Query(c *gin.Context, path string) {
	if regexpAdminLogin.MatchString(path) {
		a.queryLogin(c)
		return
	}
	if !auth(c, false) {
		c.Redirect(302, "/admin/login?redirect="+url.QueryEscape("/admin"+path))
		return
	}
	if regexpAdminIndex.MatchString(path) {
		a.queryIndex(c)
		return
	}
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

	renderer.RenderLogin(c, &d)
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

}
