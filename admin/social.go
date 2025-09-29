package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/auth/cookies"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/pquerna/otp/totp"
)

func (a *Admin) loginByPassword(w http.ResponseWriter, r *http.Request) {
	var t struct {
		Username string `json:"username"`
		Password string `json:"password"`
		OTP      string `json:"otp"`
	}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&t); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `invalid body: %v`, err)
		return
	}

	user := a.auth.AuthLogin(t.Username, t.Password)
	if user.IsGuest() {
		http.Error(w, `用户不存在/密码不正确。`, http.StatusUnauthorized)
		return
	}

	if user.OtpSecret != `` {
		if t.OTP == `` {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]any{
				`requires_otp`: true,
			})
			return
		}
		if !totp.Validate(t.OTP, user.OtpSecret) {
			http.Error(w, `动态密码不正确。`, http.StatusUnauthorized)
			return
		}
	}

	cookies.MakeCookie(w, r, int(user.ID), user.Password, user.Nickname)
	w.WriteHeader(http.StatusOK)

	// 如果没有设置 OTP，强制提醒设置。
	if user.OtpSecret == `` {
		json.NewEncoder(w).Encode(map[string]any{
			`otp_not_set`: true,
		})
	}
}

type _ClientLoginData struct {
	Authorized bool
}

// 同时处理登录跳转、未授权页面、已授权页面。
func (a *Admin) loginByClient(w http.ResponseWriter, r *http.Request) {
	random := r.URL.Query().Get(`random`)
	if random == `` {
		http.Error(w, `无效登录信息。`, 400)
		return
	}

	ac := auth.Context(r.Context())

	// 授权进行页面。
	if !ac.User.IsGuest() && r.Method == http.MethodGet {
		a.executeTemplate(w, `client.html`, _ClientLoginData{
			Authorized: false,
		})
		return
	}

	// 授权完成页面。
	if !ac.User.IsGuest() && r.Method == http.MethodPost {
		a.auth.Passkeys.SetClientLoginToken(random, cookies.TokenValue(int(ac.User.ID), ac.User.Password))
		a.executeTemplate(w, `client.html`, _ClientLoginData{
			Authorized: true,
		})
		return
	}

	// 登录重定向页面。
	u := utils.Must1(url.Parse(a.prefixed(`/login/client`)))
	q := u.Query()
	q.Set(`random`, random)
	q.Set(`authorized`, `0`)
	u.RawQuery = q.Encode()

	a.redirectToLogin(w, r, u.String())
}

func (a *Admin) loginByGithub(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	user := a.auth.AuthGitHub(code)
	if user.IsAdmin() {
		cookies.MakeCookie(w, r, int(user.ID), user.Password, user.Nickname)
		http.Redirect(w, r, `/`, http.StatusFound)
	} else {
		http.Redirect(w, r, a.prefixed(`/login`), http.StatusFound)
	}
}

func (a *Admin) loginByGoogle(w http.ResponseWriter, r *http.Request) {
	var t struct {
		Token string `json:"token"`
	}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&t); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `invalid body: %v`, err)
		return
	}
	user := a.auth.AuthGoogle(t.Token)
	if user.IsAdmin() {
		cookies.MakeCookie(w, r, int(user.ID), user.Password, user.Nickname)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}
