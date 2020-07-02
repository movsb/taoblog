package admin

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
)

var (
	regexpAdminLogin          = regexp.MustCompile(`^/login$`)
	regexpAdminLogout         = regexp.MustCompile(`^/logout$`)
	regexpAdminIndex          = regexp.MustCompile(`^/index$`)
	regexpAdminPostEdit       = regexp.MustCompile(`^/post-edit$`)
	regexpAdminTagManage      = regexp.MustCompile(`^/tag-manage$`)
	regexpAdminPostManage     = regexp.MustCompile(`^/post-manage$`)
	regexpAdminCategoryManage = regexp.MustCompile(`^/category-manage$`)
)

// AdminHeaderData ...
type AdminHeaderData struct {
	Title  string
	Header func()
}

// HeaderHook ...
func (d *AdminHeaderData) HeaderHook() string {
	if d.Header != nil {
		d.Header()
	}
	return ""
}

// AdminFooterData ...
type AdminFooterData struct {
	Footer func()
}

// FooterHook ...
func (d *AdminFooterData) FooterHook() string {
	if d.Footer != nil {
		d.Footer()
	}
	return ""
}

type LoginData struct {
	Redirect       string
	GoogleClientID string
	GitHubClientID string
}

type AdminIndexData struct {
}

type AdminTagManageData struct {
	Tags []*models.TagWithCount
}

type AdminPostManageData struct {
}

type AdminCategoryManageData struct {
	CategoryJSON string
}

type AdminPostEditData struct {
	*protocols.Post
	New bool
}

func (d *AdminPostEditData) Link() string {
	if d.New {
		return fmt.Sprint(d.Id)
	}
	return fmt.Sprintf("/%d/", d.Id)
}

func (d *AdminPostEditData) TagStr() string {
	if d.New {
		return ""
	}
	return strings.Join(d.Tags, ",")
}

type Admin struct {
	service   *service.Service
	templates *template.Template
	mux       *http.ServeMux
	auth      *auth.Auth
	canGoogle bool
}

func NewAdmin(service *service.Service, auth *auth.Auth, mux *http.ServeMux) *Admin {
	a := &Admin{
		service: service,
		mux:     mux,
		auth:    auth,
	}
	a.loadTemplates()
	mux.Handle(`/admin/`, a)
	go a.detectNetwork()
	return a
}

func (a *Admin) detectNetwork() {
	resp, err := http.Get("https://www.google.com")
	if err != nil {
		return
	}
	resp.Body.Close()
	a.canGoogle = true
}

func (a *Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		path := strings.TrimPrefix(r.URL.Path, `/admin`)
		switch path {
		case "", "/":
			w.Header().Set(`Location`, `/admin/login`)
			w.WriteHeader(302)
			return
		}
		a.Query(w, r, path)
	} else if r.Method == http.MethodPost {
		path := strings.TrimPrefix(r.URL.Path, `/admin`)
		a.Post(w, r, path)
	}
}

func (a *Admin) _auth(r *http.Request) bool {
	user := a.auth.AuthCookie2(r)
	return !user.IsGuest()
}

func (a *Admin) noCache(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "no-cache")
}

func (a *Admin) render(w io.Writer, name string, data interface{}) {
	if err := a.templates.ExecuteTemplate(w, name, data); err != nil {
		panic(err)
	}
}

func (a *Admin) loadTemplates() {
	funcs := template.FuncMap{
		"get_config": func(name string) string {
			return a.service.GetDefaultStringOption(name, "")
		},
	}

	var tmpl *template.Template
	tmpl = template.New("admin").Funcs(funcs)
	path := filepath.Join("admin", "templates", "*.html")
	tmpl, err := a.templates.ParseGlob(path)
	if err != nil {
		panic(err)
	}
	a.templates = tmpl
}

func (a *Admin) Query(w http.ResponseWriter, r *http.Request, path string) {
	if regexpAdminLogin.MatchString(path) {
		a.queryLogin(w, r)
		return
	}
	if regexpAdminLogout.MatchString(path) {
		a.queryLogout(w, r)
		return
	}
	if !a._auth(r) {
		w.Header().Set(`Location`, "/admin/login?redirect="+url.QueryEscape("/admin"+path))
		return
	}
	a.noCache(w, r)
	if regexpAdminIndex.MatchString(path) {
		a.queryIndex(w, r)
		return
	}
	if regexpAdminPostEdit.MatchString(path) {
		a.queryPostEdit(w, r)
		return
	}
	if regexpAdminTagManage.MatchString(path) {
		a.queryTagManage(w, r)
		return
	}
	if regexpAdminPostManage.MatchString(path) {
		a.queryPostManage(w, r)
		return
	}
	if regexpAdminCategoryManage.MatchString(path) {
		a.queryCategoryManage(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join("admin/statics", path))
}

func (a *Admin) Post(w http.ResponseWriter, r *http.Request, path string) {
	if regexpAdminLogin.MatchString(path) {
		a.postLogin(w, r)
		return
	}
	if !a._auth(r) {
		w.WriteHeader(403)
		return
	}
}

func (a *Admin) queryLogin(w http.ResponseWriter, r *http.Request) {
	// TODO this is a hack
	if r.URL.Query().Get("type") == "github" {
		a.postLogin(w, r)
		return
	}

	if a._auth(r) {
		w.Header().Set(`Location`, `/admin/index`)
		w.WriteHeader(302)
		return
	}
	redirect := r.URL.Query().Get("redirect")
	if redirect == "" || !strings.HasPrefix(redirect, "/") {
		redirect = "/admin/index"
	}

	d := LoginData{
		Redirect: redirect,
	}
	if a.canGoogle {
		d.GoogleClientID = a.auth.Config().Google.ClientID
	}
	d.GitHubClientID = a.auth.Config().Github.ClientID

	a.render(w, "login", &d)
}

func (a *Admin) queryLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     `login`,
		Value:    ``,
		MaxAge:   -1,
		Path:     `/`,
		Domain:   ``,
		Secure:   true,
		HttpOnly: true,
	})
	w.Header().Set(`Location`, `/admin/login`)
	w.WriteHeader(302)
}

func (a *Admin) postLogin(w http.ResponseWriter, r *http.Request) {
	redirect := ""
	success := false

	typ := r.URL.Query().Get(`type`)
	if typ == `` {
		typ = `baic`
	}

	switch typ {
	case "basic":
		username := r.FormValue("user")
		password := r.FormValue("passwd")
		success = a.auth.AuthLogin(username, password)
		redirect = r.FormValue("redirect")
	case "google":
		token := r.FormValue("token")
		success = !a.auth.AuthGoogle(token).IsGuest()
		redirect = r.FormValue("redirect")
	case "github":
		code := r.URL.Query().Get("code")
		success = !a.auth.AuthGitHub(code).IsGuest()
	}

	if success {
		a.auth.MakeCookie(w, r)
		if typ == "google" {
			d, _ := json.Marshal(map[string]interface{}{
				"redirect": redirect,
			})
			w.Write(d)
		} else {
			w.Header().Set(`Location`, redirect)
			w.WriteHeader(302)
		}
	} else {
		w.Header().Set(`Location`, `/admin/login`)
		w.WriteHeader(302)
	}
}

func (a *Admin) queryIndex(w http.ResponseWriter, r *http.Request) {
	d := &AdminIndexData{}
	header := &AdminHeaderData{
		Title: "首页",
		Header: func() {
			a.render(w, "index_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			a.render(w, "index_footer", nil)
		},
	}
	a.render(w, "header", header)
	a.render(w, "index", d)
	a.render(w, "footer", footer)
}

func (a *Admin) queryTagManage(w http.ResponseWriter, r *http.Request) {
	d := &AdminTagManageData{
		//Tags: a.server.ListTagsWithCount(0, false),
	}
	header := &AdminHeaderData{
		Title: "标签管理",
		Header: func() {
			a.render(w, "tag_manage_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			a.render(w, "tag_manage_footer", nil)
		},
	}
	a.render(w, "header", header)
	a.render(w, "tag_manage", d)
	a.render(w, "footer", footer)
}

func (a *Admin) queryPostManage(w http.ResponseWriter, r *http.Request) {
	d := &AdminPostManageData{}
	header := &AdminHeaderData{
		Title: "文章管理",
		Header: func() {
			a.render(w, "post_manage_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			a.render(w, "post_manage_footer", nil)
		},
	}
	a.render(w, "header", header)
	a.render(w, "post_manage", d)
	a.render(w, "footer", footer)
}

func (a *Admin) queryCategoryManage(w http.ResponseWriter, r *http.Request) {
	d := &AdminCategoryManageData{}
	header := &AdminHeaderData{
		Title: "分类管理",
		Header: func() {
			a.render(w, "category_manage_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			a.render(w, "category_manage_footer", nil)
		},
	}
	a.render(w, "header", header)
	a.render(w, "category_manage", d)
	a.render(w, "footer", footer)
}

func (a *Admin) queryPostEdit(w http.ResponseWriter, r *http.Request) {
	p := &protocols.Post{}
	d := AdminPostEditData{
		New:  true,
		Post: p,
	}
	header := &AdminHeaderData{
		Title: "文章编辑",
		Header: func() {
			a.render(w, "post_edit_header", nil)
		},
	}
	footer := &AdminFooterData{
		Footer: func() {
			a.render(w, "post_edit_footer", nil)
		},
	}
	a.render(w, "header", header)
	a.render(w, "post_edit", &d)
	a.render(w, "footer", footer)
}
