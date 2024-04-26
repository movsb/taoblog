package admin

import (
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/movsb/taoblog/modules/auth"
)

type LoginData struct {
	Name           string
	GoogleClientID string
	GitHubClientID string
}

func (d *LoginData) HasSocialLogins() bool {
	return d.GoogleClientID != `` || d.GitHubClientID != ``
}

type Admin struct {
	prefix    string // not including last /
	templates *template.Template
	auth      *auth.Auth
	webAuthn  *auth.WebAuthn
	canGoogle atomic.Bool

	displayName string
}

func NewAdmin(auth1 *auth.Auth, prefix string, domain, displayName string, origins []string) *Admin {
	if !strings.HasSuffix(prefix, "/") {
		panic("前缀应该以 / 结束。")
	}
	a := &Admin{
		prefix:      prefix,
		auth:        auth1,
		displayName: displayName,
		webAuthn:    auth.NewWebAuthn(auth1, domain, displayName, origins),
	}
	a.loadTemplates()
	go a.detectNetwork()
	return a
}

func (a *Admin) detectNetwork() {
	resp, err := http.Get(`https://www.gstatic.com/generate_204`)
	if err == nil {
		resp.Body.Close()
	}
	yes := err == nil && resp.StatusCode == 204
	log.Println(`google accessible: `, yes)
	a.canGoogle.Store(yes)
}

func (a *Admin) Handler() http.Handler {
	m := http.NewServeMux()

	m.HandleFunc(`GET /{$}`, a.getRoot)
	m.HandleFunc(`GET /script.js`, a.getScript)

	m.HandleFunc(`GET /login`, a.getLogin)
	m.HandleFunc(`GET /logout`, a.getLogout)

	m.HandleFunc(`GET /profile`, a.getProfile)

	m.HandleFunc(`POST /login/basic`, a.loginByPassword)
	m.HandleFunc(`GET /login/github`, a.loginByGithub)
	m.HandleFunc(`POST /login/google`, a.loginByGoogle)

	const webAuthnPrefix = `/login/webauthn/`
	m.Handle(webAuthnPrefix, a.webAuthn.Handler(webAuthnPrefix))

	prefix := strings.TrimSuffix(a.prefix, "/")
	if prefix == "" {
		return m
	}
	return http.StripPrefix(prefix, m)
}

func (a *Admin) prefixed(s string) string {
	return filepath.Join(a.prefix, s)
}

func (a *Admin) getScript(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "admin/script.js")
}

func (a *Admin) getRoot(w http.ResponseWriter, r *http.Request) {
	if a.authorized(r) {
		http.Redirect(w, r, a.prefixed(`/profile`), http.StatusFound)
		return
	}
	http.Redirect(w, r, a.prefixed(`/login`), http.StatusFound)
}

func (a *Admin) authorized(r *http.Request) bool {
	user := a.auth.AuthRequest(r)
	return user.IsAdmin()
}

func (a *Admin) loadTemplates() {
	tmpl, err := template.New("admin").ParseFiles(
		`admin/login.html`,
		`admin/profile.html`,
	) // TODO: don't use relative path.
	if err != nil {
		panic(err)
	}
	a.templates = tmpl
}

func (a *Admin) getLogin(w http.ResponseWriter, r *http.Request) {
	if a.authorized(r) {
		http.Redirect(w, r, a.prefixed(`/profile`), http.StatusFound)
		return
	}

	d := LoginData{
		Name: a.displayName,
	}
	if a.canGoogle.Load() {
		d.GoogleClientID = a.auth.Config().Google.ClientID
	}
	d.GitHubClientID = a.auth.Config().Github.ClientID

	if err := a.templates.ExecuteTemplate(w, `login.html`, &d); err != nil {
		log.Println(`admin: failed to render:`, err)
		return
	}
}

func (a *Admin) getLogout(w http.ResponseWriter, r *http.Request) {
	a.auth.RemoveCookie(w)
	http.Redirect(w, r, a.prefixed(`/login`), http.StatusFound)
}

type ProfileData struct {
	Name string
	User *auth.User
}

func (d *ProfileData) PublicKeys() []string {
	ss := make([]string, 0, len(d.User.WebAuthnCredentials()))
	for _, c := range d.User.WebAuthnCredentials() {
		ss = append(ss, base64.RawURLEncoding.EncodeToString(c.ID))
	}
	return ss
}

func (a *Admin) getProfile(w http.ResponseWriter, r *http.Request) {
	if !a.authorized(r) {
		http.Redirect(w, r, a.prefixed(`/login`), http.StatusFound)
		return
	}
	d := &ProfileData{
		Name: a.displayName,
		User: a.auth.AuthRequest(r),
	}
	if err := a.templates.ExecuteTemplate(w, `profile.html`, &d); err != nil {
		log.Println(`admin: failed to render:`, err)
		return
	}
}
