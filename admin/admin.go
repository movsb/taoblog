package admin

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service"
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
	service   *service.Service
	templates *template.Template
	mux       *utils.ServeMuxWithMethod
	auth      *auth.Auth
	canGoogle bool // not thread safe and has race, but it doesn't matter.
}

func NewAdmin(service *service.Service, auth *auth.Auth, prefix string) *Admin {
	a := &Admin{
		prefix:  prefix,
		service: service,
		mux:     utils.NewServeMuxWithMethod(),
		auth:    auth,
	}
	a.parsePrefix()
	a.route()
	a.loadTemplates()
	go a.detectNetwork()
	return a
}

func (a *Admin) Prefix() string {
	return a.prefix + `/`
}

// only Path is used.
func (a *Admin) parsePrefix() {
	u, err := url.Parse(a.prefix)
	if err != nil {
		panic(err)
	}
	a.prefix = filepath.Join(`/`, u.Path)
}

func (a *Admin) route() {
	a.mux.HandleFunc(http.MethodGet, `/`, a.getRoot)

	a.mux.HandleFunc(http.MethodGet, `/login`, a.getLogin)
	a.mux.HandleFunc(http.MethodGet, `/logout`, a.getLogout)

	a.mux.HandleFunc(http.MethodPost, `/login/basic`, a.loginByPassword)
	a.mux.HandleFunc(http.MethodGet, `/login/github`, a.loginByGithub)
	a.mux.HandleFunc(http.MethodPost, `/login/google`, a.loginByGoogle)
}

func (a *Admin) detectNetwork() {
	resp, err := http.Get("https://www.google.com")
	if err != nil {
		log.Println(`google accessible: `, false)
		return
	}
	resp.Body.Close()
	a.canGoogle = true
	log.Println(`google accessible: `, true)
}

func (a *Admin) Handler() http.Handler {
	return http.StripPrefix(a.prefix, a.mux)
}

func (a *Admin) prefixed(s string) string {
	return filepath.Join(a.prefix, s)
}

func (a *Admin) getRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != `/` {
		http.NotFound(w, r)
		return
	}
	if a.authorized(r) {
		w.Header().Set(`Location`, `/`)
		w.WriteHeader(302)
		w.Write([]byte(`Logged in, redirecting to home...`))
		return
	}
	w.Header().Set(`Location`, a.prefixed(`/login`))
	w.WriteHeader(302)
}

func (a *Admin) authorized(r *http.Request) bool {
	user := a.auth.AuthCookie2(r)
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
	if a.canGoogle {
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
