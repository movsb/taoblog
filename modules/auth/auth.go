package auth

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	googleidtokenverifier "github.com/movsb/google-idtoken-verifier"
	"github.com/movsb/taoblog/cmd/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxAuthKey struct{}

type AuthContext struct {
	User *User
}

// User entity.
type User struct {
	ID int64
}

var (
	guest = &User{ID: 0}
	admin = &User{ID: 1}
)

// IsGuest ...
func (u *User) IsGuest() bool {
	return u.ID == 0
}

// IsAdmin ...
func (u *User) IsAdmin() bool {
	return u.ID != 0
}

// MustBeAdmin will panic if not admin.
func (u *User) MustBeAdmin() {
	if !u.IsAdmin() {
		panic(status.Error(codes.PermissionDenied, `not enough permission`))
	}
}

// Context creates a new context containing the user.
func (u *User) Context(parent context.Context) context.Context {
	return context.WithValue(parent, ctxAuthKey{}, AuthContext{u})
}

type Auth struct {
	cfg config.AuthConfig
}

// New ...
func New(cfg config.AuthConfig) *Auth {
	a := Auth{}
	a.cfg = cfg
	return &a
}

// temporary
func (a *Auth) Config() config.AuthConfig {
	return a.cfg
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

func (o *Auth) AuthLogin(username string, password string) bool {
	if username != `` {
		if username == o.cfg.Basic.Username {
			if password == o.cfg.Basic.Password {
				return true
			}
		}
	}
	return false
}

func (o *Auth) AuthRequest(req *http.Request) *User {
	loginCookie, err := req.Cookie(`login`)
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

func (o *Auth) AuthGRPC(ctx context.Context) *User {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if tokens, ok := md["token"]; ok && len(tokens) > 0 {
			if tokens[0] == o.cfg.Key {
				return admin
			}
		}
	}
	return o.User(ctx)
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

// User ...
func (a *Auth) User(ctx context.Context) *User {
	if value, ok := ctx.Value(ctxAuthKey{}).(AuthContext); ok {
		return value.User
	}
	return guest
}

// MakeCookie ...
func (a *Auth) MakeCookie(w http.ResponseWriter, r *http.Request) {
	agent := r.Header.Get("User-Agent")
	cookie := a.sha1(agent + a.Login())
	http.SetCookie(w, &http.Cookie{
		Name:     `login`,
		Value:    cookie,
		MaxAge:   0,
		Path:     `/`,
		Domain:   ``,
		Secure:   true,
		HttpOnly: true,
	})
}

// RemoveCookie ...
func (a *Auth) RemoveCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     `login`,
		Value:    ``,
		MaxAge:   -1,
		Path:     `/`,
		Domain:   ``,
		Secure:   true,
		HttpOnly: true,
	})
}
