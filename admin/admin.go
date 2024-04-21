package admin

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/movsb/taoblog/modules/auth"
)

type LoginData struct {
	GoogleClientID string
	GitHubClientID string
}

func (d *LoginData) HasSocialLogins() bool {
	return d.GoogleClientID != `` || d.GitHubClientID != ``
}

type Admin struct {
	prefix    string // not including last /
	templates *template.Template
	mux       *http.ServeMux
	auth      *auth.Auth
	canGoogle atomic.Bool
}

func NewAdmin(auth *auth.Auth, prefix string) *Admin {
	if !strings.HasSuffix(prefix, "/") {
		panic("前缀应该以 / 结束。")
	}
	a := &Admin{
		prefix: prefix,
		mux:    http.NewServeMux(),
		auth:   auth,
	}
	a.route()
	a.loadTemplates()
	go a.detectNetwork()
	return a
}

func (a *Admin) route() {
	m := a.mux

	m.HandleFunc(`GET /{$}`, a.getRoot)

	m.HandleFunc(`GET /login`, a.getLogin)
	m.HandleFunc(`GET /logout`, a.getLogout)

	m.HandleFunc(`POST /login/basic`, a.loginByPassword)
	m.HandleFunc(`GET /login/github`, a.loginByGithub)
	m.HandleFunc(`POST /login/google`, a.loginByGoogle)
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
	prefix := strings.TrimSuffix(a.prefix, "/")
	if prefix == "" {
		return a.mux
	}
	return http.StripPrefix(prefix, a.mux)
}

func (a *Admin) prefixed(s string) string {
	return filepath.Join(a.prefix, s)
}

func (a *Admin) getRoot(w http.ResponseWriter, r *http.Request) {
	if a.authorized(r) {
		http.Redirect(w, r, `/`, http.StatusFound)
		return
	}
	http.Redirect(w, r, a.prefixed(`/login`), http.StatusFound)
}

func (a *Admin) authorized(r *http.Request) bool {
	user := a.auth.AuthRequest(r)
	return user.IsAdmin()
}

func (a *Admin) loadTemplates() {
	tmpl, err := template.New("admin").ParseFiles(`admin/login.html`) // TODO: don't use relative path.
	if err != nil {
		panic(err)
	}
	a.templates = tmpl
}

func (a *Admin) getLogin(w http.ResponseWriter, r *http.Request) {
	if a.authorized(r) {
		http.Redirect(w, r, a.prefixed(`/`), http.StatusFound)
		return
	}

	d := LoginData{}
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
