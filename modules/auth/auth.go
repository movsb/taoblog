package auth

import (
	"bytes"
	"context"
	"crypto/subtle"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/movsb/taoblog/modules/auth/cookies"
	"github.com/movsb/taoblog/modules/utils/db"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm"
	"github.com/phuslu/lru"
	"google.golang.org/grpc/metadata"
)

type Auth struct {
	db *taorm.DB

	userCache *lru.TTLCache[int, *models.User]

	passkeys *Passkeys
}

// DevMode：开发者模式不会限制 Cookie 的 Secure 属性，此属性只允许 HTTPS 和 localhost 的 Cookie。
func New(db *taorm.DB, home *url.URL, siteName string) *Auth {
	a := Auth{
		db: db,

		userCache: lru.NewTTLCache[int, *models.User](16),
	}

	config := &webauthn.Config{
		RPID:          home.Hostname(),
		RPDisplayName: siteName,
		RPOrigins:     []string{home.String()},
	}
	wa, err := webauthn.New(config)
	if err != nil {
		panic(err)
	}
	p := NewPasskeys(home, db, wa,
		a.GenCookieForPasskeys,
		a.DropUserCache,
	)
	a.passkeys = p

	return &a
}

func (o *Auth) Passkeys() *Passkeys {
	return o.passkeys
}

func (o *Auth) getDB(ctx context.Context) *taorm.DB {
	return db.FromContextDefault(ctx, o.db)
}

// TODO: 改成接口。
// NOTE: 错误的时候也会缓存，nil 值，以避免不必要的查询。
func (o *Auth) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	if id == int64(SystemID) {
		return system.User, nil
	}

	user, err, _ := o.userCache.GetOrLoad(ctx, int(id), func(ctx context.Context, i int) (*models.User, time.Duration, error) {
		var user models.User
		err := o.getDB(ctx).Where(`id=?`, id).Find(&user)
		if err != nil {
			if !taorm.IsNotFoundError(err) {
				return nil, 0, fmt.Errorf(`GetUserByID: %w`, err)
			}
			// NOTE: 错误的时候也会存，nil 值，以避免不必要的查询。
			// 但是最终会返回错误；CreateUser 的时候清除。
			return nil, time.Minute * 10, nil
		}
		return &user, time.Minute * 10, nil
	})

	if user == nil {
		err = sql.ErrNoRows
	}

	return user, err
}

func (o *Auth) SetUserOTPSecret(user *User, secret string) {
	o.getDB(context.Background()).Model(user.User).Where(`id=?`, user.ID).MustUpdateMap(taorm.M{
		`otp_secret`: secret,
	})
	o.DropUserCache(int(user.ID))
}

func (o *Auth) AddWebAuthnCredential(user *User, cred *webauthn.Credential) {
	existed := slices.IndexFunc(user.Credentials, func(c webauthn.Credential) bool {
		return bytes.Equal(c.PublicKey, cred.PublicKey) ||
			// 不允许为同一认证器添加多个凭证。
			// TODO 认证器为了隐私会使 AAGUID 全部为零，这里的判断无效。
			bytes.Equal(c.Authenticator.AAGUID, cred.Authenticator.AAGUID)
	})
	if existed >= 0 {
		user.Credentials[existed] = *cred
	} else {
		user.Credentials = append(user.Credentials, *cred)
	}
	o.getDB(context.Background()).Model(user.User).Where(`id=?`, user.ID).MustUpdateMap(taorm.M{
		`credentials`: user.Credentials,
	})
	o.DropUserCache(int(user.ID))
}

func (o *Auth) DropUserCache(id int) {
	o.userCache.Delete(id)
}

// https://x.com/sebastienlorber/status/1932367017065025675
func constantEqual(x, y string) bool {
	return subtle.ConstantTimeCompare([]byte(x), []byte(y)) == 1
}

// 增加用户系统后，用户名是用户编号。
// 通过纯用户名/密码登录。
func (o *Auth) AuthLogin(username string, password string) *User {
	id, err := strconv.Atoi(username)
	if err == nil {
		u, err := o.userByPasswordOrToken(id, password, ``)
		if err == nil {
			return &User{u}
		}
	}
	return guest
}

func (o *Auth) AuthToken(id int, token string) *User {
	u, err := o.userByPasswordOrToken(id, ``, token)
	if err == nil {
		return &User{u}
	}
	return guest
}

func (o *Auth) authRequest(w http.ResponseWriter, req *http.Request) *User {
	loginCookie, err := req.Cookie(cookies.CookieNameLogin)
	if err != nil {
		if a := req.Header.Get(`Authorization`); a != "" {
			id, token, _ := cookies.ParseAuthorization(a)
			return o.AuthToken(id, token)
		}
		return guest
	}

	login := loginCookie.Value
	userAgent := req.Header.Get(`User-Agent`)
	user, refresh := o.authCookie(login, userAgent)

	// 只在 home 的时候检测并刷新 cookies，避免太频繁。
	if refresh && req.URL.Path == `/` {
		cookies.MakeCookie(w, req, int(user.ID), user.Password, user.Nickname)
	}

	return user
}

func (o *Auth) authCookie(login string, userAgent string) (*User, bool) {
	var user *models.User

	valid, refresh := cookies.ValidateCookieValue(login, userAgent, func(userID int) (password string) {
		u, err := o.GetUserByID(context.TODO(), int64(userID))
		if err != nil {
			return ``
		}
		user = u
		return u.Password
	})
	if !valid {
		return guest, false
	}

	return &User{user}, refresh
}

func (o *Auth) AuthGoogle(token string) *User {
	return guest
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

func (o *Auth) AuthGitHub(code string) *User {
	return guest
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

func (a *Auth) GenCookieForPasskeys(u *User, agent string) []*proto.FinishPasskeysLoginResponse_Cookie {
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

// 仅用于测试的帐号。
// 可同时用于 HTTP 和 GRPC 请求。
func TestingUserContextForClient(user *User) context.Context {
	const userAgent = `go_test`
	md := metadata.Pairs()
	md.Append(GatewayCookie, cookies.CookieValue(userAgent, int(user.ID), user.Password))
	md.Append(GatewayUserAgent, userAgent)
	md.Append(`Authorization`, user.AuthorizationValue())
	return metadata.NewOutgoingContext(context.TODO(), md)
}

func TestingUserContextForServer(user *User) context.Context {
	return context.WithValue(context.Background(), ctxAuthKey{}, &AuthContext{
		User:       user,
		UserAgent:  `go_test`,
		RemoteAddr: localhost,
	})
}
