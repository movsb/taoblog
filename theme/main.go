package theme

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/theme/data"
	"github.com/movsb/taoblog/theme/modules/canonical"
	"github.com/movsb/taoblog/theme/modules/handle304"
	"github.com/movsb/taoblog/theme/modules/rss"
	"github.com/movsb/taoblog/theme/modules/sitemap"
	"github.com/movsb/taoblog/theme/modules/watcher"
)

// Theme ...
type Theme struct {
	cfg     *config.Config
	base    string // base directory
	service *service.Service
	auth    *auth.Auth

	redir canonical.RedirectFinder

	rss     *rss.RSS
	sitemap *sitemap.Sitemap

	tmplLock         sync.RWMutex
	partialTemplates *template.Template
	namedTemplates   map[string]*template.Template
}

// NewTheme ...
func NewTheme(cfg *config.Config, service *service.Service, auth *auth.Auth, base string) *Theme {
	t := &Theme{
		cfg:     cfg,
		base:    base,
		service: service,
		auth:    auth,
		redir:   service,
	}
	if cfg.Site.RSS.Enabled {
		t.rss = rss.New(cfg, service, auth)
	}
	if cfg.Site.Sitemap.Enabled {
		t.sitemap = sitemap.New(cfg, service, auth)
	}
	t.loadTemplates()
	t.watchTheme()
	return t
}

func createMenus(items []config.MenuItem, outer bool) string {
	menus := bytes.NewBuffer(nil)
	var genSubMenus func(buf *bytes.Buffer, items []config.MenuItem)
	a := func(item config.MenuItem) string {
		s := "<a"
		if item.Blank {
			s += " target=_blank"
		}
		if item.Link != "" {
			// TODO maybe error
			s += fmt.Sprintf(` href="%s"`, html.EscapeString(item.Link))
		}
		s += fmt.Sprintf(`>%s</a>`, html.EscapeString(item.Name))
		return s
	}
	genSubMenus = func(buf *bytes.Buffer, items []config.MenuItem) {
		if len(items) <= 0 {
			return
		}
		if outer {
			buf.WriteString("<ol>\n")
		}
		for _, item := range items {
			if len(item.Items) == 0 {
				buf.WriteString(fmt.Sprintf("<li>%s</li>\n", a(item)))
			} else {
				buf.WriteString("<li>\n")
				buf.WriteString(fmt.Sprintf("%s\n", a(item)))
				genSubMenus(buf, item.Items)
				buf.WriteString("</li>\n")
			}
		}
		if outer {
			buf.WriteString("</ol>\n")
		}
	}
	genSubMenus(menus, items)
	return menus.String()
}

func (t *Theme) getPartial() *template.Template {
	t.tmplLock.RLock()
	defer t.tmplLock.RUnlock()
	return t.partialTemplates
}

func (t *Theme) getNamed() map[string]*template.Template {
	t.tmplLock.RLock()
	defer t.tmplLock.RUnlock()
	return t.namedTemplates
}

func (t *Theme) watchTheme() {
	go func() {
		root, exts := filepath.Join(t.base, "templates"), []string{`.html`}
		w := watcher.NewFolderChangedWatcher(root, exts)
		for range w.Watch() {
			log.Println(`reload templates`)
			t.loadTemplates()
		}
	}()
	go func() {
		root, exts := filepath.Join(t.base, "styles"), []string{`.scss`}
		w := watcher.NewFolderChangedWatcher(root, exts)
		for range w.Watch() {
			cmd := exec.Command(`make`, `theme`)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Println(err)
			} else {
				log.Println(`rebuild styles`)
			}
		}
	}()
}

func (t *Theme) loadTemplates() {
	t.tmplLock.Lock()
	defer t.tmplLock.Unlock()

	defer func() {
		if err := recover(); err != nil {
			t.partialTemplates = template.Must(template.New(`empty`).Parse(``))
			t.namedTemplates = nil
			log.Println(err)
		}
	}()

	menustr := createMenus(t.cfg.Menus, false)
	funcs := template.FuncMap{
		// https://githut.com/golang/go/issues/14256
		"raw": func(s string) template.HTML {
			return template.HTML(s)
		},
		"get_config": func(name string) string {
			switch name {
			case `blog_name`:
				return t.service.Name()
			case `blog_desc`:
				return t.service.Description()
			default:
				panic(`cannot get this option`)
			}
		},
		"menus": func() template.HTML {
			return template.HTML(menustr)
		},
		"render": func(name string, data *data.Data) error {
			if t := data.Template.Lookup(name); t != nil {
				return t.Execute(data.Writer, data)
			}
			if t := t.getPartial().Lookup(name); t != nil {
				return t.Execute(data.Writer, data)
			}
			return nil
		},
	}

	t.partialTemplates = template.New(`partial`).Funcs(funcs)
	t.namedTemplates = make(map[string]*template.Template)

	templateFiles, err := filepath.Glob(filepath.Join(t.base, `templates`, `*.html`))
	if err != nil {
		panic(err)
	}

	for _, path := range templateFiles {
		name := filepath.Base(path)
		if name[0] == '_' {
			template.Must(t.partialTemplates.ParseFiles(path))
		} else {
			tmpl := template.Must(template.New(name).Funcs(funcs).ParseFiles(path))
			t.namedTemplates[tmpl.Name()] = tmpl
		}
	}
}

func (t *Theme) Exception(w http.ResponseWriter, req *http.Request, e interface{}) bool {
	switch te := e.(type) {
	case *service.PostNotFoundError, *service.TagNotFoundError, *service.CategoryNotFoundError:
		if t.redir != nil {
			target, err := t.redir.FindRedirect(req.URL.Path)
			if err != nil {
				log.Println(`FindRedirect failed. `, err)
				// fallthrough
			}
			if target != `` {
				http.Redirect(w, req, target, http.StatusPermanentRedirect)
				return true
			}
		}
		w.WriteHeader(404)
		t.getNamed()[`404.html`].Execute(w, nil)
		return true
	case string: // hack hack
		switch te {
		case "403":
			w.WriteHeader(403)
			t.getNamed()[`403.html`].Execute(w, nil)
			return true
		}
	}
	return false
}

func (t *Theme) ProcessHomeQueries(w http.ResponseWriter, req *http.Request, query url.Values) bool {
	if p := query.Get("p"); p != "" {
		w.Header().Set(`Location`, fmt.Sprintf("/%s/", p))
		w.WriteHeader(301)
		return true
	}
	return false
}

func (t *Theme) QueryHome(w http.ResponseWriter, req *http.Request) error {
	d := data.NewDataForHome(t.cfg, t.auth.AuthRequest(req), t.service)
	tmpl := t.getNamed()[`home.html`]
	d.Template = tmpl
	d.Writer = w
	if err := tmpl.Execute(w, d); err != nil {
		log.Println(err)
	}
	return nil
}

func (t *Theme) querySearch(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForSearch(t.cfg, t.auth.AuthRequest(r), t.service, r)
	tmpl := t.getNamed()[`search.html`]
	d.Template = tmpl
	d.Writer = w

	if err := tmpl.Execute(w, d); err != nil {
		panic(err)
	}
}

func (t *Theme) queryPosts(w http.ResponseWriter, r *http.Request) {
	if handle304.ArticleRequest(w, r, t.service.LastArticleUpdateTime()) {
		return
	}

	d := data.NewDataForPosts(t.cfg, t.auth.AuthRequest(r), t.service, r)
	tmpl := t.getNamed()[`posts.html`]
	d.Template = tmpl
	d.Writer = w

	handle304.ArticleResponse(w, t.service.LastArticleUpdateTime())

	if err := tmpl.Execute(w, d); err != nil {
		panic(err)
	}
}

func (t *Theme) queryTags(w http.ResponseWriter, r *http.Request) {
	if handle304.ArticleRequest(w, r, t.service.LastArticleUpdateTime()) {
		return
	}

	d := data.NewDataForTags(t.cfg, t.auth.AuthRequest(r), t.service)
	tmpl := t.getNamed()[`tags.html`]
	d.Template = tmpl
	d.Writer = w

	handle304.ArticleResponse(w, t.service.LastArticleUpdateTime())

	if err := tmpl.Execute(w, d); err != nil {
		panic(err)
	}

}

func (t *Theme) QueryByID(w http.ResponseWriter, req *http.Request, id int64) error {
	post := t.service.GetPostByID(id)
	t.userMustCanSeePost(req, post)
	t.incView(post.Id)
	t.tempRenderPost(w, req, post)
	return nil
}

func (t *Theme) incView(id int64) {
	t.service.IncrementPostPageView(id)
}

func (t *Theme) QueryBySlug(w http.ResponseWriter, req *http.Request, tree string, slug string) (int64, error) {
	post := t.service.GetPostBySlug(tree, slug)
	t.userMustCanSeePost(req, post)
	t.incView(post.Id)
	t.tempRenderPost(w, req, post)
	return post.Id, nil
}

func (t *Theme) QueryByPage(w http.ResponseWriter, req *http.Request, parents string, slug string) (int64, error) {
	post := t.service.GetPostByPage(parents, slug)
	t.userMustCanSeePost(req, post)
	t.incView(post.Id)
	t.tempRenderPost(w, req, post)
	return post.Id, nil
}

func (t *Theme) tempRenderPost(w http.ResponseWriter, req *http.Request, p *protocols.Post) {
	if handle304.ArticleRequest(w, req, time.Unix(int64(p.Modified), 0)) {
		return
	}

	d := data.NewDataForPost(t.cfg, t.auth.AuthRequest(req), t.service, p)
	tmpl := t.getNamed()[`post.html`]
	d.Template = tmpl
	d.Writer = w

	handle304.ArticleResponse(w, time.Unix(int64(p.Modified), 0))

	if err := tmpl.Execute(w, d); err != nil {
		log.Println(err)
	}
}

func (t *Theme) QueryByTags(w http.ResponseWriter, req *http.Request, tags []string) {
	d := data.NewDataForTag(t.cfg, t.auth.AuthRequest(req), t.service, tags)
	tmpl := t.getNamed()[`tag.html`]
	d.Template = tmpl
	d.Writer = w
	if err := tmpl.Execute(w, d); err != nil {
		panic(err)
	}
}

func (t *Theme) QueryFile(w http.ResponseWriter, req *http.Request, postID int64, file string) {
	file = filepath.Clean(file)
	fp, err := t.service.Store().Open(postID, file)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, req)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer fp.Close()
	stat, err := fp.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, req, stat.Name(), stat.ModTime(), fp)
}

func (t *Theme) userMustCanSeePost(req *http.Request, post *protocols.Post) {
	user := t.auth.AuthRequest(req)
	if user.IsGuest() && post.Status != "public" {
		panic("403")
	}
}

func (t *Theme) QuerySpecial(w http.ResponseWriter, req *http.Request, file string) bool {
	if req.Method == http.MethodGet {
		switch file {
		case `/rss`:
			if t.cfg.Site.RSS.Enabled {
				t.rss.ServeHTTP(w, req)
				return true
			}
		case `/sitemap.xml`:
			if t.cfg.Site.Sitemap.Enabled {
				t.sitemap.ServeHTTP(w, req)
				return true
			}
		case `/search`:
			t.querySearch(w, req)
			return true
		case `/posts`:
			t.queryPosts(w, req)
			return true
		case `/tags`:
			t.queryTags(w, req)
			return true
		}
	}
	return false
}

// QueryStatic ...
func (t *Theme) QueryStatic(w http.ResponseWriter, req *http.Request, file string) {
	path := filepath.Join(t.base, `statics`, file)
	http.ServeFile(w, req, path)
}
