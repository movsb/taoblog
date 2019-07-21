package auth

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	googleidtokenverifier "github.com/movsb/google-idtoken-verifier"
)

type ctxAuthKey struct{}

type AuthContext struct {
	User *User
}

type User struct {
	ID int64
}

var guest = &User{
	ID: 0,
}

var admin = &User{
	ID: 1,
}

func (u *User) IsGuest() bool {
	return u.ID == 0
}

func (u *User) IsAdmin() bool {
	return u.ID != 0
}

func (u *User) Context(parent context.Context) context.Context {
	return context.WithValue(parent, ctxAuthKey{}, AuthContext{u})
}

func HashPassword(password string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(password)))
}

type Auth struct {
	key string
	SavedAuth
}

type SavedAuth struct {
	Username           string `json:"username,omitempty"`
	Password           string `json:"password,omitempty"`
	GoogleClientID     string `json:"google_client_id,omitempty"`
	AdminGoogleID      string `json:"admin_google_id,omitempty"`
	GitHubClientID     string `json:"github_client_id,omitempty"`
	GitHubClientSecret string `json:"github_client_secret,omitempty"`
	AdminGitHubID      int64  `json:"admin_github_id,omitempty"`
}

func (a SavedAuth) Encode() string {
	bys, _ := json.Marshal(&a)
	return string(bys)
}

func (o *Auth) SetKey(key string) {
	o.key = key
}

func (o *Auth) SetLogin(login string) {
	if err := json.Unmarshal([]byte(login), &o.SavedAuth); err != nil {
		panic(err)
	}
}

func (o *Auth) SetGitHub(clientID string, clientSecret string, adminID string) {
	o.GitHubClientID = clientID
	o.GitHubClientSecret = clientSecret
	fmt.Sscanf(adminID, "%d", &o.AdminGitHubID)
}

func (o *Auth) Login() string {
	return o.Username + "," + o.Password
}

func (*Auth) sha1(in string) string {
	h := sha1.Sum([]byte(in))
	return fmt.Sprintf("%x", h)
}

func (o *Auth) AuthLogin(username string, password string) bool {
	if username == o.Username {
		if o.sha1(password) == o.Password {
			return true
		}
	}
	return false
}

func (o *Auth) AuthCookie(c *gin.Context) *User {
	login, err := c.Cookie("login")
	if err != nil {
		return guest
	}

	agent := c.GetHeader("User-Agent")
	if agent == "" {
		return guest
	}

	if o.sha1(agent+o.Login()) == login {
		return admin
	}

	return guest
}

func (o *Auth) AuthHeader(c *gin.Context) *User {
	key := c.GetHeader("Authorization")
	if key == o.key {
		return admin
	}
	return guest
}

func (o *Auth) AuthGoogle(token string) *User {
	fullClientID := o.GoogleClientID + ".apps.googleusercontent.com"
	claims, err := googleidtokenverifier.Verify(token, fullClientID)
	if err != nil {
		return guest
	}
	if claims.Sub == o.AdminGoogleID {
		return admin
	}
	return guest
}

func (o *Auth) AuthGitHub(code string) *User {
	accessTokenURL, _ := url.Parse("https://github.com/login/oauth/access_token")
	values := url.Values{}
	values.Set("client_id", o.GitHubClientID)
	values.Set("client_secret", o.GitHubClientSecret)
	values.Set("code", code)
	accessTokenURL.RawQuery = values.Encode()
	req, err := http.NewRequest("POST", accessTokenURL.String(), nil)
	if err != nil {
		return guest
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return guest
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return guest
	}
	var accessTokenStruct struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&accessTokenStruct); err != nil {
		return guest
	}
	req, err = http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return guest
	}
	req.Header.Set("Authorization", "token "+accessTokenStruct.AccessToken)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return guest
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return guest
	}
	var user struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return guest
	}
	if user.ID != 0 && user.ID == o.AdminGitHubID {
		return admin
	}
	return guest
}

func (a *Auth) AuthContext(ctx context.Context) *User {
	if value, ok := ctx.Value(ctxAuthKey{}).(AuthContext); ok {
		return value.User
	}
	return guest
}

func (o *Auth) MakeCookie(c *gin.Context) {
	agent := c.GetHeader("User-Agent")
	cookie := o.sha1(agent + o.Login())
	c.SetCookie("login", cookie, 0, "/", "", true, true)
}

func (o *Auth) DeleteCookie(c *gin.Context) {
	c.SetCookie("login", "", -1, "/", "", true, true)
}

func (o *Auth) Middle(c *gin.Context) {
	cookieUser := o.AuthCookie(c)
	if !cookieUser.IsGuest() {
		c.Next()
		return
	}
	headerUser := o.AuthHeader(c)
	if !headerUser.IsGuest() {
		c.Next()
		return
	}
	c.AbortWithStatus(401)
}
