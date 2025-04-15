package admin

import (
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	urlpkg "net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/gateway"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
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
	svc     proto.TaoBlogServer
	gateway *gateway.Gateway

	customTheme *config.ThemeConfig

	templates *utils.TemplateLoader

	displayName string
}

func NewAdmin(devMode bool, gateway *gateway.Gateway, svc proto.TaoBlogServer, auth1 *auth.Auth, prefix string, domain, displayName string, origins []string, options ...Option) *Admin {
	if !strings.HasSuffix(prefix, "/") {
		panic("前缀应该以 / 结束。")
	}

	var rootFS fs.FS
	var tmplFS fs.FS

	if devMode {
		dir := dir.SourceRelativeDir()
		rootFS = os.DirFS(dir.Join(`statics`))
		tmplFS = utils.NewOSDirFS(dir.Join(`templates`))
	} else {
		rootFS = utils.Must1(fs.Sub(root, `statics`))
		tmplFS = utils.Must1(fs.Sub(root, `templates`))
	}

	a := &Admin{
		rootFS:      rootFS,
		tmplFS:      tmplFS,
		svc:         svc,
		gateway:     gateway,
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

// 下面的网址在中国已经能访问，不能再用它来判断是否可以访问 Google 主站。
//
//	https://www.gstatic.com/generate_204
func (a *Admin) detectNetwork() {
	resp, err := http.Get(`https://www.google.com/favicon.ico`)
	if err == nil {
		resp.Body.Close()
	}
	// 无需判断状态码，只需保证能访问（证书正确）即可。
	yes := err == nil
	log.Println(`google accessible: `, yes)
	a.canGoogle.Store(yes)
}

// fs 可以传 nil，表示使用 embed 文件系统。
func (a *Admin) Handler() http.Handler {
	m := http.NewServeMux()

	// 奇怪，这里不能写 GET /，会冲突。
	m.HandleFunc(`/`, func(w http.ResponseWriter, r *http.Request) {
		utils.ServeFSWithAutoModTime(w, r, a.rootFS, r.URL.Path)
	})
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

func (a *Admin) prefixed(s string) string {
	return filepath.Join(a.prefix, s)
}

func (a *Admin) redirectToLogin(w http.ResponseWriter, r *http.Request, to string) {
	args := urlpkg.Values{}
	args.Set(`u`, to)

	u, err := urlpkg.Parse(a.prefixed(`/login`))
	if err != nil {
		panic(err)
	}

	u.RawQuery = args.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}

func (a *Admin) requireLogin(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.auth.AuthRequest(r).IsGuest() {
			h.ServeHTTP(w, r)
			return
		}
		a.redirectToLogin(w, r, r.RequestURI)
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
	if !a.auth.AuthRequest(r).IsGuest() {
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
	a    *Admin
	Name string
	User *auth.User
}

// 输出的是 ID，不是 PublicKey。目前只作展示使用。
func (d *ProfileData) PublicKeys() []string {
	ss := make([]string, 0, len(d.User.Credentials))
	for _, c := range d.User.Credentials {
		ss = append(ss, base64.RawURLEncoding.EncodeToString(c.ID))
	}
	return ss
}

func (d *ProfileData) AvatarURL() string {
	return d.a.gateway.AvatarURL(int(d.User.ID))
}

func (a *Admin) getProfile(w http.ResponseWriter, r *http.Request) {
	d := &ProfileData{
		a:    a,
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
	if isNew := r.URL.Query().Get(`new`) == `1`; isNew {
		post := utils.Must1(a.svc.CreatePost(r.Context(),
			&proto.Post{
				Type:       `post`,
				SourceType: `markdown`,
				Source:     fmt.Sprintf("# %s\n\n", models.Untitled),
			},
		))
		args := urlpkg.Values{}
		args.Set(`id`, fmt.Sprint(post.Id))
		url := a.prefixed(`editor`) + `?` + args.Encode()
		http.Redirect(w, r, url, http.StatusFound)
		return
	}
	if pid, _ := strconv.Atoi(r.URL.Query().Get(`id`)); pid > 0 {
		rsp, err := a.svc.GetPost(r.Context(), &proto.GetPostRequest{
			Id: int32(pid),
			GetPostOptions: &proto.GetPostOptions{
				ContentOptions: co.For(co.Editor),
				WithUserPerms:  true,
			},
		})
		if err != nil {
			panic(err)
		}
		d := EditorData{
			User: a.auth.AuthRequest(r),
			Post: rsp,
		}
		a.executeTemplate(w, `editor.html`, &d)
		return
	}
	utils.HTTPError(w, http.StatusBadRequest)
}
