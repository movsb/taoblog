package admin

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"image/png"
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
	"github.com/movsb/taoblog/modules/auth/cookies"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/theme/modules/handle304"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

//go:embed statics templates dynamic
var _embed embed.FS
var _root = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `admin`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithScripts(module, `dynamic/script.js`)
	})
}

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
		rootFS = utils.Must1(fs.Sub(_embed, `statics`))
		tmplFS = utils.Must1(fs.Sub(_embed, `templates`))
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
//
// 换：<https://google.com/generate_204>
// https://x.com/kholinchan/status/1638515221643026432
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
		handle304.MustRevalidate(w)
		utils.ServeFSWithAutoModTime(w, r, a.rootFS, r.URL.Path)
	})
	m.Handle(`GET /{$}`, a.requireLogin(a.getRoot))

	m.HandleFunc(`GET /login`, a.getLogin)
	m.HandleFunc(`GET /logout`, a.getLogout)
	m.HandleFunc(`POST /logout`, a.postLogout)

	m.Handle(`GET /profile`, a.requireLogin(a.getProfile))
	m.Handle(`GET /editor`, a.requireLogin(a.getEditor))
	m.Handle(`GET /drafts`, a.requireLogin(a.getDrafts))
	m.Handle(`GET /reorder`, a.requireLogin(a.getReorder))
	m.Handle(`GET /otp`, a.requireLogin(a.getOTP))
	m.Handle(`POST /otp`, a.requireLogin(a.postOTP))
	m.Handle(`GET /notify`, a.requireLogin(a.getNotify))
	m.Handle(`POST /notify`, a.requireLogin(a.postNotify))
	m.Handle(`GET /category`, a.requireLogin(a.getCategory))

	m.HandleFunc(`POST /login/basic`, a.loginByPassword)
	m.HandleFunc(`GET /login/github`, a.loginByGithub)
	m.HandleFunc(`POST /login/google`, a.loginByGoogle)
	m.HandleFunc(`/login/client`, a.loginByClient)

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
		if !auth.Context(r.Context()).User.IsGuest() {
			w.Header().Add(`Cache-Control`, `no-store`)
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
	if !auth.Context(r.Context()).User.IsGuest() {
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
		// d.GoogleClientID = a.auth.Config().Google.ClientID
	}
	// d.GitHubClientID = a.auth.Config().Github.ClientID

	a.executeTemplate(w, `login.html`, &d)
}

func (a *Admin) getLogout(w http.ResponseWriter, r *http.Request) {
	cookies.RemoveCookie(w)
	http.Redirect(w, r, a.prefixed(`/login`), http.StatusFound)
}
func (a *Admin) postLogout(w http.ResponseWriter, r *http.Request) {
	cookies.RemoveCookie(w)
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
		User: auth.Context(r.Context()).User,
	}
	a.executeTemplate(w, `profile.html`, &d)
}

type ReorderData struct {
	Posts []*proto.Post
}

func (a *Admin) getReorder(w http.ResponseWriter, r *http.Request) {
	d := &ReorderData{
		Posts: utils.Must1(a.svc.GetTopPosts(r.Context(), &proto.GetTopPostsRequest{})).Posts,
	}
	a.executeTemplate(w, `reorder.html`, &d)
}

type OTPData struct {
	User *auth.User

	Prompt bool

	Set   bool
	Image template.URL // base64 data url
	URL   template.URL // otp url

	Validate bool
	Error    string
}

func (a *Admin) getOTP(w http.ResponseWriter, r *http.Request) {
	ac := auth.MustNotBeGuest(r.Context())

	isPrompt := r.URL.Query().Get(`prompt`) == `1`
	if isPrompt {
		d := OTPData{
			Prompt: true,
		}
		a.executeTemplate(w, `otp.html`, &d)
		return
	}

	isSet := r.URL.Query().Get(`set`) == `1`
	if isSet {
		// 防止重复生成覆盖。
		if ac.User.OtpSecret != `` {
			utils.HTTPError(w, http.StatusBadRequest)
			return
		}

		key := utils.Must1(totp.Generate(totp.GenerateOpts{
			Issuer:      a.displayName,
			AccountName: fmt.Sprint(ac.User.ID),
		}))

		image := func() string {
			buf := bytes.NewBuffer(nil)
			png.Encode(buf, utils.Must1(key.Image(500, 500)))
			return utils.CreateDataURL(buf.Bytes()).String()
		}()

		d := OTPData{
			User:  ac.User,
			Set:   true,
			URL:   template.URL(key.URL()),
			Image: template.URL(image),
		}

		a.executeTemplate(w, `otp.html`, &d)
		return
	}

	utils.HTTPError(w, http.StatusNotFound)
}

func (a *Admin) postOTP(w http.ResponseWriter, r *http.Request) {
	ac := auth.MustNotBeGuest(r.Context())

	isValidate := r.URL.Query().Get(`validate`) == `1`
	if isValidate {
		// 防止重复生成覆盖。
		if ac.User.OtpSecret != `` {
			utils.HTTPError(w, http.StatusForbidden)
			return
		}

		var (
			// TODO: 不要使用前端上传的，内部缓存作 session。
			url      = r.PostFormValue(`url`)
			password = r.PostFormValue(`password`)
		)
		key := utils.Must1(otp.NewKeyFromURL(url))
		if !totp.Validate(password, key.Secret()) {
			w.WriteHeader(http.StatusForbidden)
			d := OTPData{
				Validate: true,
				Error:    `错误：输入的动态密码无法完成验证。`,
			}
			a.executeTemplate(w, `otp.html`, &d)
			return
		}

		a.auth.SetUserOTPSecret(ac.User, key.Secret())

		d := OTPData{
			Validate: true,
			Error:    ``,
		}
		a.executeTemplate(w, `otp.html`, &d)
		return
	}

	utils.HTTPError(w, http.StatusNotFound)
}

type NotifyData struct {
	Email     string
	BarkToken string
}

func (a *Admin) getNotify(w http.ResponseWriter, r *http.Request) {
	ac := auth.MustNotBeGuest(r.Context())

	d := NotifyData{
		Email:     ac.User.Email,
		BarkToken: ac.User.BarkToken,
	}

	a.executeTemplate(w, `notify.html`, &d)
}

func (a *Admin) postNotify(w http.ResponseWriter, r *http.Request) {
	ac := auth.MustNotBeGuest(r.Context())

	var (
		email     = r.PostFormValue(`email`)
		barkToken = r.PostFormValue(`bark_token`)
	)

	utils.Must1(a.auth.Passkeys.UpdateUser(
		r.Context(),
		&proto.UpdateUserRequest{
			User: &proto.User{
				Id:        ac.User.ID,
				Email:     email,
				BarkToken: barkToken,
			},
			UpdateEmail:     true,
			UpdateBarkToken: true,
		}),
	)

	http.Redirect(w, r, a.prefixed(`/profile`), http.StatusFound)
}

type CategoryData struct {
	Settings *proto.Settings
	Cats     []*proto.Category
}

func (a *Admin) getCategory(w http.ResponseWriter, r *http.Request) {
	ac := auth.MustNotBeGuest(r.Context())
	_ = ac

	d := CategoryData{
		Cats:     utils.Must1(a.svc.ListCategories(r.Context(), &proto.ListCategoriesRequest{})).Categories,
		Settings: utils.Must1(a.svc.GetUserSettings(r.Context(), &proto.GetUserSettingsRequest{})),
	}

	a.executeTemplate(w, `category.html`, &d)
}

type EditorData struct {
	User *auth.User
	Post *proto.Post
	Cats []*proto.Category
}

func (a *Admin) getEditor(w http.ResponseWriter, r *http.Request) {
	ac := auth.Context(r.Context())

	if isNew := r.URL.Query().Get(`new`) == `1`; isNew {
		ty := `markdown`
		if st := r.URL.Query().Get(`type`); st != `` {
			ty = st
		}

		post := utils.Must1(a.svc.CreateUntitledPost(r.Context(), &proto.CreateUntitledPostRequest{
			Type: ty,
		}))
		args := urlpkg.Values{}
		args.Set(`id`, fmt.Sprint(post.Post.Id))
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
			utils.HTTPError(w, 404)
			return
		}
		if rsp.UserId != int32(ac.User.ID) {
			utils.HTTPError(w, 403)
			return
		}
		catsRsp, err := a.svc.ListCategories(r.Context(), &proto.ListCategoriesRequest{})
		if err != nil {
			utils.HTTPError(w, 400)
		}
		d := EditorData{
			User: ac.User,
			Post: rsp,
			Cats: catsRsp.Categories,
		}
		a.executeTemplate(w, `editor.html`, &d)
		return
	}

	utils.HTTPError(w, http.StatusBadRequest)
}

type DraftsData struct {
	// 按更新时间降序排列。
	Posts []*proto.Post
}

func (a *Admin) getDrafts(w http.ResponseWriter, r *http.Request) {
	auth.MustNotBeGuest(r.Context())

	d := DraftsData{
		Posts: utils.Must1(a.svc.ListPosts(r.Context(), &proto.ListPostsRequest{
			OrderBy:   `modified desc`,
			Ownership: proto.Ownership_OwnershipDrafts,
		})).Posts,
	}

	a.executeTemplate(w, `drafts.html`, &d)
}
