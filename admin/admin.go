package admin

import (
	"embed"
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

//go:embed editor.js script.js
var root embed.FS

type LoginData struct {
	Name           string
	GoogleClientID string
	GitHubClientID string
}

func (d *LoginData) HasSocialLogins() bool {
	return d.GoogleClientID != `` || d.GitHubClientID != ``
}

type Admin struct {
	prefix    string
	templates *template.Template
	auth      *auth.Auth
	webAuthn  *auth.WebAuthn
	canGoogle atomic.Bool

	svc *service.Service

	displayName string
}

func NewAdmin(svc *service.Service, auth1 *auth.Auth, prefix string, domain, displayName string, origins []string) *Admin {
	if !strings.HasSuffix(prefix, "/") {
		panic("前缀应该以 / 结束。")
	}
	a := &Admin{
		svc:         svc,
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

	m.Handle(`GET /{$}`, a.requireLogin(a.getRoot))
	m.Handle(`GET /script.js`, a.serveFile(`script.js`))
	m.Handle(`GET /editor.js`, a.serveFile(`editor.js`))

	m.HandleFunc(`GET /login`, a.getLogin)
	m.HandleFunc(`GET /logout`, a.getLogout)
	m.HandleFunc(`POST /logout`, a.postLogout)

	m.Handle(`GET /profile`, a.requireLogin(a.getProfile))
	m.Handle(`GET /editor`, a.requireLogin(a.getEditor))

	m.HandleFunc(`POST /login/basic`, a.loginByPassword)
	m.HandleFunc(`GET /login/github`, a.loginByGithub)
	m.HandleFunc(`POST /login/google`, a.loginByGoogle)

	const webAuthnPrefix = `/login/webauthn/`
	m.Handle(webAuthnPrefix, a.webAuthn.Handler(webAuthnPrefix))

	return http.StripPrefix(strings.TrimSuffix(a.prefix, "/"), m)
}

func (a *Admin) prefixed(s string) string {
	return filepath.Join(a.prefix, s)
}

func (a *Admin) serveFile(path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, root, path)
	})
}

func (a *Admin) redirectToLogin(w http.ResponseWriter, r *http.Request, to string) {
	args := url.Values{}
	args.Set(`u`, to)

	u, err := url.Parse(a.prefixed(`/login`))
	if err != nil {
		panic(err)
	}

	u.RawQuery = args.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}

func (a *Admin) requireLogin(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.auth.AuthRequest(r).IsAdmin() {
			a.redirectToLogin(w, r, r.RequestURI)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (a *Admin) getRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, a.prefixed(`/profile`), http.StatusFound)
}

func (a *Admin) loadTemplates() {
	tmpl, err := template.New("admin").ParseFiles(
		`admin/login.html`,
		`admin/profile.html`,
		`admin/editor.html`,
	) // TODO: don't use relative path.
	if err != nil {
		panic(err)
	}
	a.templates = tmpl
}

func (a *Admin) getLogin(w http.ResponseWriter, r *http.Request) {
	if a.auth.AuthRequest(r).IsAdmin() {
		to := a.prefixed(`/profile`)
		if u := r.URL.Query().Get(`u`); u != "" {
			to = u
		}
		http.Redirect(w, r, to, http.StatusFound)
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
func (a *Admin) postLogout(w http.ResponseWriter, r *http.Request) {
	a.auth.RemoveCookie(w)
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
	d := &ProfileData{
		Name: a.displayName,
		User: a.auth.AuthRequest(r),
	}
	if err := a.templates.ExecuteTemplate(w, `profile.html`, &d); err != nil {
		log.Println(`admin: failed to render:`, err)
		return
	}
}

type EditorData struct {
	User *auth.User
	Post *protocols.Post
}

func (a *Admin) getEditor(w http.ResponseWriter, r *http.Request) {
	d := &EditorData{
		User: a.auth.AuthRequest(r),
	}

	if pidStr := r.URL.Query().Get(`id`); pidStr != "" {
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			panic(err)
		}
		post := a.svc.MustGetPost(int64(pid))
		d.Post = post
	}

	if err := a.templates.ExecuteTemplate(w, `editor.html`, &d); err != nil {
		log.Println(`admin: failed to render:`, err)
		return
	}
}
