package auth

import (
	"bytes"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/go-webauthn/webauthn/webauthn"
	googleidtokenverifier "github.com/movsb/google-idtoken-verifier"
	"github.com/movsb/taoblog/cmd/config"
)

type Auth struct {
	cfg      config.AuthConfig
	optioner Optioner
	devMode  bool
}

type Optioner interface {
	SetOption(name string, value any)
	GetDefaultStringOption(name string, def string) string
}

// New ...
// DevMode：开发者模式不会限制 Cookie 的 Secure 属性，此属性只允许 HTTPS 和 localhost 的 Cookie。
func New(cfg config.AuthConfig, devMode bool) *Auth {
	a := Auth{
		cfg:     cfg,
		devMode: devMode,
	}
	if len(cfg.AdminEmails) > 0 {
		admin.Email = cfg.AdminEmails[0]
	}
	admin.DisplayName = cfg.AdminName
	return &a
}

func (a *Auth) SetAdminWebAuthnCredentials(j string) {
	if err := json.Unmarshal([]byte(j), &admin.webAuthnCredentials); err != nil {
		panic("credentials:" + err.Error())
	}
}

func (a *Auth) SetService(optioner Optioner) {
	a.optioner = optioner
}

// temporary
func (a *Auth) Config() *config.AuthConfig {
	return &a.cfg
}

// 找不到返回空。
func (o *Auth) GetUserByID(id int64) *User {
	if id == admin.ID {
		return admin
	}
	return guest
}

func (o *Auth) AddWebAuthnCredential(user *User, cred *webauthn.Credential) {
	existed := slices.IndexFunc(user.webAuthnCredentials, func(c webauthn.Credential) bool {
		return bytes.Equal(c.PublicKey, cred.PublicKey)
	})
	if existed >= 0 {
		user.webAuthnCredentials[existed] = *cred
	} else {
		user.webAuthnCredentials = append(user.webAuthnCredentials, *cred)
	}
	body, err := json.Marshal(user.webAuthnCredentials)
	if err != nil {
		panic(err)
	}
	o.optioner.SetOption(`admin_webauthn_credentials`, string(body))
}

func (o *Auth) Login() string {
	return fmt.Sprintf(
		`%s:%s`,
		o.cfg.Basic.Username,
		o.sha1(o.cfg.Basic.Password),
	)
}

func (*Auth) sha1(in string) string {
	h := sha1.Sum([]byte(in))
	return fmt.Sprintf("%x", h)
}

func constantEqual(x, y string) bool {
	return subtle.ConstantTimeCompare([]byte(x), []byte(y)) == 1
}

func (o *Auth) AuthLogin(username string, password string) *User {
	if username != `` {
		if constantEqual(username, o.cfg.Basic.Username) {
			if constantEqual(password, o.cfg.Basic.Password) {
				return admin
			}
		}
	}
	return guest
}

func (o *Auth) AuthRequest(req *http.Request) *User {
	loginCookie, err := req.Cookie(CookieNameLogin)
	if err != nil {
		return guest
	}

	login := loginCookie.Value
	userAgent := req.Header.Get(`User-Agent`)
	return o.AuthCookie(login, userAgent)
}

func (o *Auth) AuthCookie(login string, userAgent string) *User {
	if userAgent == "" {
		return guest
	}

	if o.sha1(userAgent+o.Login()) == login {
		return admin
	}

	return guest
}

func (o *Auth) AuthGoogle(token string) *User {
	fullClientID := o.cfg.Google.ClientID + ".apps.googleusercontent.com"
	claims, err := googleidtokenverifier.Verify(token, fullClientID)
	if err != nil {
		return guest
	}
	if claims.Sub == o.cfg.Google.UserID {
		return admin
	}
	return guest
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
	var user struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return guest
	}
	if user.ID != 0 && user.ID == o.cfg.Github.UserID {
		return admin
	}
	return guest
}

const (
	CookieNameLogin  = `taoblog.login`
	CookieNameUserID = `taoblog.user_id`
)

// MakeCookie ...
func (a *Auth) MakeCookie(u *User, w http.ResponseWriter, r *http.Request) {
	agent := r.Header.Get("User-Agent")
	cookie := a.sha1(agent + a.Login())
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameLogin,
		Value:    cookie,
		MaxAge:   0,
		Path:     `/`,
		Domain:   ``,
		Secure:   !a.devMode,
		HttpOnly: true,
	})
	// 只用于前端展示使用，不能用作凭证。
	http.SetCookie(w, &http.Cookie{
		Name:   CookieNameUserID,
		Value:  fmt.Sprint(u.ID),
		MaxAge: 0,
		Path:   `/`,
		Domain: ``,
	})
}

// RemoveCookie ...
func (a *Auth) RemoveCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameLogin,
		Value:    ``,
		MaxAge:   -1,
		Path:     `/`,
		Domain:   ``,
		Secure:   !a.devMode,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   CookieNameUserID,
		Value:  ``,
		MaxAge: -1,
		Path:   `/`,
		Domain: ``,
	})
}
