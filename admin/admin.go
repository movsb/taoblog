package admin

import (
	"embed"
	"encoding/base64"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/theme/modules/handle304"
)

//go:embed statics templates
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
	rootFS fs.FS
	tmplFS fs.FS

	prefix    string
	auth      *auth.Auth
	webAuthn  *auth.WebAuthn
	canGoogle atomic.Bool

	// NOTE：这是进程内直接调用的。
	// 如果改成连接，需要考虑 metadata 转发问题。
	svc proto.TaoBlogServer

	customTheme *config.ThemeConfig

	templates *utils.TemplateLoader

	displayName string
}

func NewAdmin(devMode bool, svc proto.TaoBlogServer, auth1 *auth.Auth, prefix string, domain, displayName string, origins []string, options ...Option) *Admin {
	if !strings.HasSuffix(prefix, "/") {
		panic("前缀应该以 / 结束。")
	}

	var rootFS fs.FS
	var tmplFS fs.FS

	if devMode {
		dir := dir.SourceRelativeDir()
		rootFS = os.DirFS(dir.Join(`statics`))
		tmplFS = utils.NewLocal(dir.Join(`templates`))
	} else {
		rootFS = utils.Must(fs.Sub(root, `statics`))
		tmplFS = utils.Must(fs.Sub(root, `templates`))
	}

	a := &Admin{
		rootFS:      rootFS,
		tmplFS:      tmplFS,
		svc:         svc,
		prefix:      prefix,
		auth:        auth1,
		displayName: displayName,
		webAuthn:    auth.NewWebAuthn(auth1, domain, displayName, origins),
	}

	for _, opt := range options {
		opt(a)
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

// fs 可以传 nil，表示使用 embed 文件系统。
func (a *Admin) Handler() http.Handler {
	m := http.NewServeMux()

	// 奇怪，这里不能写 GET /，会冲突。
	m.Handle(`/`, a.cachedHandler(http.FileServerFS(a.rootFS)))
	m.Handle(`GET /{$}`, a.requireLogin(a.getRoot))

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

func (a *Admin) cachedHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if service.DevMode() {
			handle304.MustRevalidate(w)
		} else {
			handle304.CacheShortly(w)
		}
		h.ServeHTTP(w, r)
	})
}

func (a *Admin) prefixed(s string) string {
	return filepath.Join(a.prefix, s)
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
	var customTheme string
	if a.customTheme != nil {
		customTheme = a.customTheme.Stylesheets.Render()
	}
	funcs := template.FuncMap{
		"apply_site_theme_customs": func() template.HTML {
			return template.HTML(customTheme)
		},
	}
	a.templates = utils.NewTemplateLoader(a.tmplFS, funcs, nil)
}

func (a *Admin) executeTemplate(w io.Writer, name string, data any) {
	t2 := a.templates.GetNamed(name)
	if t2 == nil {
		panic(`未找到模板：` + name)
	}
	if err := t2.Execute(w, data); err != nil {
		log.Println(err)
	}
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

	a.executeTemplate(w, `login.html`, &d)
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
	a.executeTemplate(w, `profile.html`, &d)
}

type EditorData struct {
	User *auth.User
	Post *proto.Post
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
		rsp, err := a.svc.GetPost(r.Context(), &proto.GetPostRequest{
			Id:             int32(pid),
			ContentOptions: co.For(co.Editor),
		})
		if err != nil {
			panic(err)
		}
		d.Post = rsp
	}
	a.executeTemplate(w, `editor.html`, &d)
}
