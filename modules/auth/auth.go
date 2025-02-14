package auth

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/go-webauthn/webauthn/webauthn"
	googleidtokenverifier "github.com/movsb/google-idtoken-verifier"
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm"
	"google.golang.org/grpc/metadata"
)

type Auth struct {
	db *taorm.DB

	cfg config.AuthConfig
}

// DevMode：开发者模式不会限制 Cookie 的 Secure 属性，此属性只允许 HTTPS 和 localhost 的 Cookie。
func New(cfg config.AuthConfig, db *taorm.DB) *Auth {
	a := Auth{
		db:  db,
		cfg: cfg,
	}
	return &a
}

// temporary
func (a *Auth) Config() *config.AuthConfig {
	return &a.cfg
}

// NOTE：系统管理员因为不因为登录所以不允许查找。
// TODO: 改成接口。
func (o *Auth) GetUserByID(id int64) (*models.User, error) {
	var user models.User
	if err := o.db.Where(`id=?`, id).Find(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (o *Auth) AddWebAuthnCredential(user *User, cred *webauthn.Credential) {
	existed := slices.IndexFunc(user.Credentials, func(c webauthn.Credential) bool {
		return bytes.Equal(c.PublicKey, cred.PublicKey) ||
			// 不允许为同一认证器添加多个凭证。
			bytes.Equal(c.Authenticator.AAGUID, cred.Authenticator.AAGUID)
	})
	if existed >= 0 {
		user.Credentials[existed] = *cred
	} else {
		user.Credentials = append(user.Credentials, *cred)
	}
	o.db.Model(user.User).Where(`id=?`, user.ID).MustUpdateMap(taorm.M{
		`credentials`: user.Credentials,
	})
}

func login(username, password string) string {
	return fmt.Sprintf(
		`%s:%s`,
		username,
		shasum(password),
	)
}

func shasum(in string) string {
	h := sha1.Sum([]byte(in))
	return fmt.Sprintf("%x", h)
}

func constantEqual(x, y string) bool {
	return subtle.ConstantTimeCompare([]byte(x), []byte(y)) == 1
}

// 增加用户系统后，用户名可以是用户编号。
// TODO: 防止每次验证都检查数据库，应该限流。
func (o *Auth) AuthLogin(username string, password string) *User {
	id, err := strconv.Atoi(username)
	if err == nil {
		u, err := o.userByKey(id, password)
		if err == nil {
			return &User{u}
		}
	}
	return guest
}

func (o *Auth) AuthRequest(req *http.Request) *User {
	loginCookie, err := req.Cookie(CookieNameLogin)
	if err != nil {
		if a := req.Header.Get(`Authorization`); a != "" {
			if id, token, ok := ParseAuthorization(a); ok {
				return o.AuthLogin(fmt.Sprint(id), token)
			}
		}
		return guest
	}

	login := loginCookie.Value
	userAgent := req.Header.Get(`User-Agent`)
	return o.authCookie(login, userAgent)
}

func (o *Auth) authCookie(login string, userAgent string) *User {
	if userAgent == "" {
		return guest
	}

	splits := strings.Split(login, `:`)
	if len(splits) != 2 {
		return guest
	}
	id, err := strconv.Atoi(splits[0])
	if err != nil {
		return guest
	}
	u, err := o.GetUserByID(int64(id))
	if err != nil {
		return guest
	}

	if login == cookieValue(userAgent, splits[0], u.Password) {
		return &User{u}
	}

	return guest
}

func (o *Auth) AuthGoogle(token string) *User {
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
}

func (o *Auth) AuthGitHub(code string) *User {
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
}

const (
	CookieNameLogin  = `taoblog.login`
	CookieNameUserID = `taoblog.user_id`
)

func cookieValue(userAgent, username, password string) string {
	return username + ":" + shasum(userAgent+login(username, password))
}

// MakeCookie ...
func (a *Auth) MakeCookie(u *User, w http.ResponseWriter, r *http.Request) {
	agent := r.Header.Get("User-Agent")
	cookie := cookieValue(agent, fmt.Sprint(u.ID), u.Password)
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameLogin,
		Value:    cookie,
		MaxAge:   0,
		Path:     `/`,
		Domain:   ``,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	// 只用于前端展示使用，不能用作凭证。
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameUserID,
		Value:    fmt.Sprint(u.ID),
		MaxAge:   0,
		Path:     `/`,
		Domain:   ``,
		Secure:   true,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *Auth) GenCookieForPasskeys(u *User, agent string) []*proto.FinishPasskeysLoginResponse_Cookie {
	cookie := cookieValue(agent, fmt.Sprint(u.ID), u.Password)
	return []*proto.FinishPasskeysLoginResponse_Cookie{
		{
			Name:     CookieNameLogin,
			Value:    cookie,
			HttpOnly: true,
		},
		{
			Name:     CookieNameUserID,
			Value:    fmt.Sprint(u.ID),
			HttpOnly: false,
		},
	}
}

// RemoveCookie ...
func (a *Auth) RemoveCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameLogin,
		Value:    ``,
		MaxAge:   -1,
		Path:     `/`,
		Domain:   ``,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameUserID,
		Value:    ``,
		MaxAge:   -1,
		Path:     `/`,
		Domain:   ``,
		Secure:   true,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
}

// 仅用于测试的帐号。
// 可同时用于 HTTP 和 GRPC 请求。
func TestingUserContext(user *User, userAgent string) context.Context {
	md := metadata.Pairs()
	md.Append(GatewayCookie, cookieValue(userAgent, fmt.Sprint(user.ID), user.Password))
	md.Append(GatewayUserAgent, userAgent)
	md.Append(`Authorization`, fmt.Sprintf(`token %d:%s`, user.ID, user.Password))
	return metadata.NewOutgoingContext(context.TODO(), md)
}
