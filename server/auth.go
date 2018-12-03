package main

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

type GenericAuth struct {
	savedUser     string
	savedPassword string
	key           string
}

func (o *GenericAuth) SetKey(key string) {
	o.key = key
}

func (o *GenericAuth) SetLogin(login string) {
	ss := strings.SplitN(login, ",", 2)
	o.savedUser = ss[0]
	o.savedPassword = ss[1]
}

func (o *GenericAuth) Login() string {
	return o.savedUser + "," + o.savedPassword
}

func (*GenericAuth) sha1(in string) string {
	h := sha1.Sum([]byte(in))
	return fmt.Sprintf("%x", h)
}

func (o *GenericAuth) Auth(username string, password string) bool {
	if username == o.savedUser {
		if o.sha1(password) == o.savedPassword {
			return true
		}
	}
	return false
}

func (o *GenericAuth) AuthCookie(c *gin.Context) bool {
	login, err := c.Cookie("login")
	if err != nil {
		return false
	}

	agent := c.GetHeader("User-Agent")
	if agent == "" {
		return false
	}

	return o.sha1(agent+o.Login()) == login
}

func (o *GenericAuth) AuthHeader(c *gin.Context) bool {
	key := c.GetHeader("Authorization")
	return key == o.key
}

func (o *GenericAuth) MakeCookie(c *gin.Context) {
	agent := c.GetHeader("User-Agent")
	cookie := o.sha1(agent + o.Login())
	c.SetCookie("login", cookie, 0, "/", "", true, true)
}
