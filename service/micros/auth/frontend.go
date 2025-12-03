package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/cookies"
	"github.com/movsb/taoblog/service/micros/auth/user"
	auth_webauthn "github.com/movsb/taoblog/service/micros/auth/webauthn"
	"github.com/movsb/taorm"
)

type Auth struct {
	db *taorm.DB

	userManager *UserManager

	getHome, getName func() string

	// 站点名字和主页地址可能变动，变动后重建。
	webAuthn        atomic.Pointer[webauthn.WebAuthn]
	webAuthnHandler atomic.Value // http.Handler
}

func NewAuth(db *taorm.DB, getHome, getName func() string, userManager *UserManager) *Auth {
	a := Auth{
		db:          db,
		userManager: userManager,

		getHome: getHome,
		getName: getName,
	}

	a.createWebAuthn()

	// TODO 需要订阅配置变更并重建。
	go func() {
		lastName, lastHome := getName(), getHome()
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			name, home := getName(), getHome()
			if name != lastName || home != lastHome {
				lastName, lastHome = name, home
				a.createWebAuthn()
			}
		}
	}()

	return &a
}

func (o *Auth) GetWA() *webauthn.WebAuthn {
	return o.webAuthn.Load()
}

func (o *Auth) GetWebAuthnHandler() http.Handler {
	return o.webAuthnHandler.Load().(http.Handler)
}

func (o *Auth) createWebAuthn() {
	homeURL := utils.Must1(url.Parse(o.getHome()))

	wa := utils.Must1(webauthn.New(&webauthn.Config{
		RPID:          homeURL.Hostname(),
		RPDisplayName: o.getName(),
		RPOrigins:     []string{homeURL.String()},
	}))

	wah := auth_webauthn.NewWebAuthn(o.userManager, wa)

	o.webAuthn.Store(wa)
	o.webAuthnHandler.Store(wah.Handler())
}

// 通过纯用户名/密码登录。
//
// 增加用户系统后，用户名是用户编号。
func (o *Auth) AuthLogin(username string, password string) *user.User {
	id, err := strconv.Atoi(username)
	if err == nil {
		u, err := o.userManager.GetUserByPassword(context.Background(), id, password)
		if err == nil {
			return u
		}
	}
	return user.Guest
}

func (o *Auth) GetUserByID(id int) (*user.User, error) {
	return o.userManager.GetUserByID(context.Background(), id)
}
func (o *Auth) GetUserByToken(id int, token string) (*user.User, error) {
	return o.userManager.GetUserByToken(context.Background(), id, token)
}

func (o *Auth) AuthToken(id int, token string) *user.User {
	u, err := o.userManager.GetUserByToken(context.Background(), id, token)
	if err == nil {
		return u
	}
	return user.Guest
}

func (o *Auth) AuthRequest(w http.ResponseWriter, req *http.Request) *user.User {
	loginCookie, err := req.Cookie(cookies.CookieNameLogin)
	if err != nil {
		if a := req.Header.Get(`Authorization`); a != "" {
			id, token, _ := cookies.ParseAuthorization(a)
			return o.AuthToken(id, token)
		}
		return user.Guest
	}

	login := loginCookie.Value
	userAgent := req.Header.Get(`User-Agent`)
	user, refresh := o.AuthCookie(login, userAgent)

	// 只在 home 的时候检测并刷新 cookies，避免太频繁。
	if refresh && req.URL.Path == `/` {
		cookies.MakeCookie(w, req, int(user.ID), user.Password, user.Nickname)
	}

	return user
}

func (o *Auth) AuthCookie(login string, userAgent string) (*user.User, bool) {
	var uu *user.User

	valid, refresh := cookies.ValidateCookieValue(login, userAgent, func(userID int) (password string) {
		u, err := o.userManager.GetUserByID(context.TODO(), userID)
		if err != nil {
			return ``
		}
		uu = u
		return u.Password
	})
	if !valid {
		return user.Guest, false
	}

	return uu, refresh
}

func (o *Auth) AuthGoogle(token string) *user.User {
	return user.Guest
	/*
		fullClientID := o.cfg.Google.ClientID + ".apps.googleusercontent.com"
		claims, err := googleidtokenverifier.Verify(token, fullClientID)
		if err != nil || claims.Sub == "" {
			return guest
		}

		var user models.User
		if err := o.db.Where(`google_user_id=?`, claims.Sub).Find(&user); err != nil {
			log.Println(`谷歌登录错误：`, err)
			return guest
		}

		return &User{User: &user}
	*/
}

func (o *Auth) AuthGitHub(code string) *user.User {
	return user.Guest
	/*
		accessTokenURL, _ := url.Parse("https://github.com/login/oauth/access_token")
		values := url.Values{}
		values.Set("client_id", o.cfg.Github.ClientID)
		values.Set("client_secret", o.cfg.Github.ClientSecret)
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
		var gUser struct {
			ID int64 `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&gUser); err != nil || gUser.ID == 0 {
			return guest
		}

		var user models.User
		if err := o.db.Where(`github_user_id=?`, gUser.ID).Find(&user); err != nil {
			log.Println(`鸡盒登录错误：`, err)
			return guest
		}

		return &User{User: &user}
	*/
}

func (a *Auth) GenCookieForPasskeys(u *user.User, agent string) []*proto.FinishPasskeysLoginResponse_Cookie {
	cookie := cookies.CookieValue(agent, int(u.ID), u.Password)
	return []*proto.FinishPasskeysLoginResponse_Cookie{
		{
			Name:     cookies.CookieNameLogin,
			Value:    cookie,
			HttpOnly: true,
		},
		{
			Name:     cookies.CookieNameUserID,
			Value:    fmt.Sprint(u.ID),
			HttpOnly: false,
		},
		{
			Name:     cookies.CookieNameNickname,
			Value:    url.PathEscape(u.Nickname),
			HttpOnly: false,
		},
	}
}
