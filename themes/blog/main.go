package blog

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
	"github.com/movsb/taoblog/service/modules/rss"
	"github.com/movsb/taoblog/service/modules/sitemap"
	"github.com/movsb/taoblog/themes/blog/pkg/watcher"
	"github.com/movsb/taoblog/themes/data"
	"github.com/movsb/taoblog/themes/modules/handle304"
)

// Blog ...
type Blog struct {
	cfg     *config.Config
	base    string // base directory
	service *service.Service
	auth    *auth.Auth

	rss     *rss.RSS
	sitemap *sitemap.Sitemap

	tmplLock         sync.RWMutex
	partialTemplates *template.Template
	namedTemplates   map[string]*template.Template
}

// NewBlog ...
func NewBlog(cfg *config.Config, service *service.Service, auth *auth.Auth, base string) *Blog {
	b := &Blog{
		cfg:     cfg,
		base:    base,
		service: service,
		auth:    auth,
	}
	if cfg.Site.RSS.Enabled {
		b.rss = rss.New(cfg, service, auth)
	}
	if cfg.Site.Sitemap.Enabled {
		b.sitemap = sitemap.New(cfg, service, auth)
	}
	b.loadTemplates()
	b.watchTheme()
	return b
}

func createMenus(items []config.MenuItem) string {
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
		buf.WriteString("<ol>\n")
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
		buf.WriteString("</ol>\n")
	}
	genSubMenus(menus, items)
	return menus.String()
}

func (b *Blog) getPartial() *template.Template {
	b.tmplLock.RLock()
	defer b.tmplLock.RUnlock()
	return b.partialTemplates
}

func (b *Blog) getNamed() map[string]*template.Template {
	b.tmplLock.RLock()
	defer b.tmplLock.RUnlock()
	return b.namedTemplates
}

func (b *Blog) watchTheme() {
	go func() {
		root, exts := filepath.Join(b.base, "templates"), []string{`.html`}
		w := watcher.NewFolderChangedWatcher(root, exts)
		for range w.Watch() {
			log.Println(`reload templates`)
			b.loadTemplates()
		}
	}()
	go func() {
		root, exts := filepath.Join(b.base, "statics", "sass"), []string{`.scss`}
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

func (b *Blog) loadTemplates() {
	b.tmplLock.Lock()
	defer b.tmplLock.Unlock()

	defer func() {
		if err := recover(); err != nil {
			b.partialTemplates = template.Must(template.New(`empty`).Parse(``))
			b.namedTemplates = nil
			log.Println(err)
		}
	}()

	menustr := createMenus(b.cfg.Menus)
	funcs := template.FuncMap{
		// https://github.com/golang/go/issues/14256
		"raw": func(s string) template.HTML {
			return template.HTML(s)
		},
		"get_config": func(name string) string {
			switch name {
			case `blog_name`:
				return b.service.Name()
			case `blog_desc`:
				return b.service.Description()
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
			if t := b.getPartial().Lookup(name); t != nil {
				return t.Execute(data.Writer, data)
			}
			return nil
		},
	}

	b.partialTemplates = template.New(`partial`).Funcs(funcs)
	b.namedTemplates = make(map[string]*template.Template)

	templateFiles, err := filepath.Glob(filepath.Join(b.base, `templates`, `*.html`))
	if err != nil {
		panic(err)
	}

	for _, path := range templateFiles {
		name := filepath.Base(path)
		if name[0] == '_' {
			template.Must(b.partialTemplates.ParseFiles(path))
		} else {
			tmpl := template.Must(template.New(name).Funcs(funcs).ParseFiles(path))
			b.namedTemplates[tmpl.Name()] = tmpl
		}
	}
}

func (b *Blog) Exception(w http.ResponseWriter, req *http.Request, e interface{}) bool {
	switch te := e.(type) {
	case *service.PostNotFoundError, *service.TagNotFoundError, *service.CategoryNotFoundError:
		w.WriteHeader(404)
		b.getNamed()[`404.html`].Execute(w, nil)
		return true
	case string: // hack hack
		switch te {
		case "403":
			w.WriteHeader(403)
			b.getNamed()[`403.html`].Execute(w, nil)
			return true
		}
	}
	return false
}

func (b *Blog) ProcessHomeQueries(w http.ResponseWriter, req *http.Request, query url.Values) bool {
	if p := query.Get("p"); p != "" {
		w.Header().Set(`Location`, fmt.Sprintf("/%s/", p))
		w.WriteHeader(301)
		return true
	}
	return false
}

func (b *Blog) QueryHome(w http.ResponseWriter, req *http.Request) error {
	d := data.NewDataForHome(b.cfg, b.auth.AuthCookie2(req), b.service)
	t := b.getNamed()[`home.html`]
	d.Template = t
	d.Writer = w
	if err := t.Execute(w, d); err != nil {
		log.Println(err)
	}
	return nil
}

func (b *Blog) querySearch(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForSearch(b.cfg, b.auth.AuthCookie2(r), b.service, r)
	t := b.getNamed()[`search.html`]
	d.Template = t
	d.Writer = w

	if err := t.Execute(w, d); err != nil {
		panic(err)
	}
}

func (b *Blog) queryPosts(w http.ResponseWriter, r *http.Request) {
	if handle304.ArticleRequest(w, r, b.service.LastArticleUpdateTime()) {
		return
	}

	d := data.NewDataForPosts(b.cfg, b.auth.AuthCookie2(r), b.service, r)
	t := b.getNamed()[`posts.html`]
	d.Template = t
	d.Writer = w

	handle304.ArticleResponse(w, b.service.LastArticleUpdateTime())

	if err := t.Execute(w, d); err != nil {
		panic(err)
	}
}

func (b *Blog) queryTags(w http.ResponseWriter, r *http.Request) {
	if handle304.ArticleRequest(w, r, b.service.LastArticleUpdateTime()) {
		return
	}

	d := data.NewDataForTags(b.cfg, b.auth.AuthCookie2(r), b.service)
	t := b.getNamed()[`tags.html`]
	d.Template = t
	d.Writer = w

	handle304.ArticleResponse(w, b.service.LastArticleUpdateTime())

	if err := t.Execute(w, d); err != nil {
		panic(err)
	}

}

func (b *Blog) QueryByID(w http.ResponseWriter, req *http.Request, id int64) error {
	post := b.service.GetPostByID(id)
	b.userMustCanSeePost(req, post)
	b.incView(post.Id)
	b.tempRenderPost(w, req, post)
	return nil
}

func (b *Blog) incView(id int64) {
	b.service.IncrementPostPageView(id)
}

func (b *Blog) QueryBySlug(w http.ResponseWriter, req *http.Request, tree string, slug string) (int64, error) {
	post := b.service.GetPostBySlug(tree, slug)
	b.userMustCanSeePost(req, post)
	b.incView(post.Id)
	b.tempRenderPost(w, req, post)
	return post.Id, nil
}

func (b *Blog) QueryByPage(w http.ResponseWriter, req *http.Request, parents string, slug string) (int64, error) {
	post := b.service.GetPostByPage(parents, slug)
	b.userMustCanSeePost(req, post)
	b.incView(post.Id)
	b.tempRenderPost(w, req, post)
	return post.Id, nil
}

func (b *Blog) tempRenderPost(w http.ResponseWriter, req *http.Request, p *protocols.Post) {
	if handle304.ArticleRequest(w, req, time.Unix(int64(p.Modified), 0)) {
		return
	}

	d := data.NewDataForPost(b.cfg, b.auth.AuthCookie2(req), b.service, p)
	t := b.getNamed()[`post.html`]
	d.Template = t
	d.Writer = w

	handle304.ArticleResponse(w, time.Unix(int64(p.Modified), 0))

	if err := t.Execute(w, d); err != nil {
		log.Println(err)
	}
}

func (b *Blog) QueryByTags(w http.ResponseWriter, req *http.Request, tags []string) {
	d := data.NewDataForTag(b.cfg, b.auth.AuthCookie2(req), b.service, tags)
	t := b.getNamed()[`tag.html`]
	d.Template = t
	d.Writer = w
	if err := t.Execute(w, d); err != nil {
		panic(err)
	}
}

func (b *Blog) QueryFile(w http.ResponseWriter, req *http.Request, postID int64, file string) {
	file = filepath.Clean(file)
	fp, err := b.service.Store().Open(postID, file)
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

func (b *Blog) userMustCanSeePost(req *http.Request, post *protocols.Post) {
	user := b.auth.AuthCookie2(req)
	if user.IsGuest() && post.Status != "public" {
		panic("403")
	}
}

func (b *Blog) QuerySpecial(w http.ResponseWriter, req *http.Request, file string) bool {
	if req.Method == http.MethodGet {
		switch file {
		case `/rss`:
			if b.cfg.Site.RSS.Enabled {
				b.rss.ServeHTTP(w, req)
				return true
			}
		case `/sitemap.xml`:
			if b.cfg.Site.Sitemap.Enabled {
				b.sitemap.ServeHTTP(w, req)
				return true
			}
		case `/search`:
			b.querySearch(w, req)
			return true
		case `/posts`:
			b.queryPosts(w, req)
			return true
		case `/tags`:
			b.queryTags(w, req)
			return true
		}
	}
	return false
}

// QueryStatic ...
func (b *Blog) QueryStatic(w http.ResponseWriter, req *http.Request, file string) {
	path := filepath.Join(b.base, `statics`, file)
	http.ServeFile(w, req, path)
}
