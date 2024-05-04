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
	"strconv"
	"sync"
	"time"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/version"
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

	specialMux *http.ServeMux
}

func New(cfg *config.Config, service *service.Service, auth *auth.Auth, base string) *Theme {
	t := &Theme{
		cfg:     cfg,
		base:    base,
		service: service,
		auth:    auth,
		redir:   service,

		specialMux: http.NewServeMux(),
	}

	m := t.specialMux

	if r := cfg.Site.RSS; r.Enabled {
		t.rss = rss.New(service, auth, rss.WithArticleCount(r.ArticleCount))
		m.Handle(`GET /rss`, t.LastPostTime304Handler(t.rss))
	}
	if cfg.Site.Sitemap.Enabled {
		t.sitemap = sitemap.New(service, auth)
		m.Handle(`GET /sitemap.xml`, t.LastPostTime304Handler(t.sitemap))
	}

	m.HandleFunc(`GET /search`, t.querySearch)
	m.Handle(`GET /posts`, t.LastPostTime304HandlerFunc(t.queryPosts))
	m.HandleFunc(`GET /tweets`, t.queryTweets)
	m.Handle(`GET /tags`, t.LastPostTime304HandlerFunc(t.queryTags))

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
		"partial": func(name string, data *data.Data, data2 any) error {
			if t := t.getPartial().Lookup(name); t != nil {
				return t.Execute(data.Writer, data2)
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
	// 兼容非常早期的 p 查询参数。随时可以移除。
	if id, err := strconv.Atoi(query.Get("p")); err == nil && id > 0 {
		http.Redirect(w, req, fmt.Sprintf(`/%d/`, id), http.StatusPermanentRedirect)
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

func (t *Theme) Post304Handler(w http.ResponseWriter, r *http.Request, p *protocols.Post) bool {
	h3 := handle304.New(
		handle304.WithNotModified(time.Unix(int64(p.Modified), 0)),
		handle304.WithEntityTag(version.GitCommit, p.Modified, p.LastCommentedAt),
	)
	if h3.Match(w, r) {
		return true
	}
	h3.Response(w)
	return false
}

func (t *Theme) LastPostTime304HandlerFunc(h http.HandlerFunc) http.Handler {
	return t.LastPostTime304Handler(h)
}

func (t *Theme) LastPostTime304Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		last := t.service.LastArticleUpdateTime()
		h3 := handle304.New(
			handle304.WithNotModified(last),
			handle304.WithEntityTag(version.GitCommit, last),
		)
		if h3.Match(w, r) {
			return
		}
		h3.Response(w)
		h.ServeHTTP(w, r)
	})
}

func (t *Theme) queryPosts(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForPosts(t.cfg, t.auth.AuthRequest(r), t.service, r)
	tmpl := t.getNamed()[`posts.html`]
	d.Template = tmpl
	d.Writer = w

	if err := tmpl.Execute(w, d); err != nil {
		panic(err)
	}
}

func (t *Theme) queryTweets(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForTweets(t.service.Config(), t.auth.AuthRequest(r), t.service)
	tmpl := t.getNamed()[`tweets.html`]
	d.Template = tmpl
	d.Writer = w

	if err := tmpl.Execute(w, d); err != nil {
		panic(err)
	}
}

func (t *Theme) queryTags(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForTags(t.cfg, t.auth.AuthRequest(r), t.service)
	tmpl := t.getNamed()[`tags.html`]
	d.Template = tmpl
	d.Writer = w

	if err := tmpl.Execute(w, d); err != nil {
		panic(err)
	}
}

func (t *Theme) QueryByID(w http.ResponseWriter, req *http.Request, id int64) error {
	post := t.service.GetPostByID(id)
	t.userMustCanSeePost(req, post)

	if post.Type == `page` {
		link := t.service.GetLink(id)
		// 因为只处理了一层页面路径，所以要判断一下。
		if link != t.service.GetPlainLink(id) {
			http.Redirect(w, req, link, http.StatusPermanentRedirect)
			return nil
		}
		return nil
	}

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
	if t.Post304Handler(w, req, p) {
		return
	}

	var d *data.Data
	d = data.NewDataForPost(t.cfg, t.auth.AuthRequest(req), t.service, p, t.service.ListPostAllComments(req, p.Id))

	var tmpl *template.Template
	if p.Type == `tweet` {
		tmpl = t.getNamed()[`tweet.html`]
	} else {
		tmpl = t.getNamed()[`post.html`]
	}
	d.Template = tmpl
	d.Writer = w

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
	fs, err := t.service.FileSystemForPost(postID)
	if err != nil {
		panic(err)
	}
	fp, err := fs.OpenFile(file)
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
	if h, p := t.specialMux.Handler(req); p != "" {
		h.ServeHTTP(w, req)
		return true
	}
	return false
}

var cacheControl = `max-age=21600, must-revalidate`

func (t *Theme) QueryStatic(w http.ResponseWriter, req *http.Request, file string) {
	path := filepath.Join(t.base, `statics`, file)

	// 开发模式下不要缓存资源文件，因为经常更新，否则需要强制刷新，太蛋疼了，会加强很多字体文件，很慢。
	if t.service.DevMode() {
		fp, err := os.Open(filepath.Clean(path))
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer fp.Close()
		http.ServeContent(w, req, path, time.Time{}, fp)
		return
	}

	// 正式环境也不要缓存太久，因为博客在经常更新。
	w.Header().Add(`Cache-Control`, cacheControl)
	http.ServeFile(w, req, path)
}
