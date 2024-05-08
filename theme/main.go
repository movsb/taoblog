package theme

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/theme/blog"
	"github.com/movsb/taoblog/theme/data"
	"github.com/movsb/taoblog/theme/modules/canonical"
	"github.com/movsb/taoblog/theme/modules/handle304"
	"github.com/movsb/taoblog/theme/modules/rss"
	"github.com/movsb/taoblog/theme/modules/sitemap"
	"github.com/movsb/taorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Theme ...
type Theme struct {
	rootFS fs.FS
	tmplFS fs.FS

	cfg     *config.Config
	service *service.Service
	auth    *auth.Auth

	redir canonical.RedirectFinder

	rss     *rss.RSS
	sitemap *sitemap.Sitemap

	templates *utils.TemplateLoader

	specialMux *http.ServeMux
}

func New(devMode bool, cfg *config.Config, service *service.Service, auth *auth.Auth) *Theme {
	var rootFS, tmplFS, stylesFS fs.FS

	if devMode {
		rootFS = os.DirFS(`theme/blog/statics`)
		tmplFS = utils.NewLocal(`theme/blog/templates`)
		stylesFS = utils.NewLocal(`theme/blog/styles`)
	} else {
		// TODO 硬编码成 blog 了。
		rootFS = utils.Must(fs.Sub(blog.Root, `statics`))
		tmplFS = utils.Must(fs.Sub(blog.Root, `templates`))
		stylesFS = utils.Must(fs.Sub(blog.Root, `styles`))
	}

	t := &Theme{
		rootFS: rootFS,
		tmplFS: tmplFS,

		cfg:     cfg,
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
	t.watchStyles(stylesFS)

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

func (t *Theme) loadTemplates() {
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
			if t := t.templates.GetPartial(name); t != nil {
				return t.Execute(data.Writer, data)
			}
			// TODO 找不到应该报错。
			return nil
		},
		"partial": func(name string, data *data.Data, data2 any) error {
			if t := t.templates.GetPartial(name); t != nil {
				return t.Execute(data.Writer, data2)
			}
			// TODO 找不到应该报错。
			return nil
		},
	}

	t.templates = utils.NewTemplateLoader(t.tmplFS, funcs)
}

func (t *Theme) watchStyles(stylesFS fs.FS) {
	if changed, ok := stylesFS.(utils.FsWithChangeNotify); ok {
		log.Println(`Listening for style changes`)
		go func() {
			for event := range changed.Changed() {
				switch event.Op {
				case fsnotify.Create, fsnotify.Remove, fsnotify.Write:
					// log.Println(`Will run make`)
					cmd := exec.Command(`make`, `theme`)
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					if err := cmd.Run(); err != nil {
						log.Println(err)
					} else {
						log.Println(`Rebuilt styles`)
					}
				}
			}
		}()
	}
}

func (t *Theme) executeTemplate(name string, w io.Writer, d *data.Data) {
	t2 := t.templates.GetNamed(name)
	if t2 == nil {
		panic(`未找到模板：` + name)
	}
	if d == nil {
		d = &data.Data{}
	}
	d.Template = t2
	d.Writer = w
	if err := t2.Execute(w, d); err != nil {
		log.Println(err)
	}
}

func (t *Theme) Exception(w http.ResponseWriter, req *http.Request, e interface{}) bool {
	if err, ok := e.(error); ok {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.PermissionDenied:
				w.WriteHeader(403)
				t.executeTemplate(`403.html`, w, nil)
				return true
			case codes.NotFound:
				w.WriteHeader(http.StatusNotFound)
				t.executeTemplate(`404.html`, w, nil)
				return true
			}
		}
		if taorm.IsNotFoundError(err) {
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
			t.executeTemplate(`404.html`, w, nil)
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
	d := data.NewDataForHome(req.Context(), t.cfg, t.service)
	t.executeTemplate(`home.html`, w, d)
	return nil
}

func (t *Theme) querySearch(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForSearch(t.cfg, t.auth.AuthRequest(r), t.service, r)
	t.executeTemplate(`search.html`, w, d)
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
	d := data.NewDataForPosts(r.Context(), t.cfg, t.service, r)
	t.executeTemplate(`posts.html`, w, d)
}

func (t *Theme) queryTweets(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForTweets(r.Context(), t.service.Config(), t.service)
	t.executeTemplate(`tweets.html`, w, d)
}

func (t *Theme) queryTags(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForTags(t.cfg, t.auth.AuthRequest(r), t.service)
	t.executeTemplate(`tags.html`, w, d)
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

	rsp, err := t.service.GetPostComments(req.Context(), &protocols.GetPostCommentsRequest{Id: p.Id})
	if err != nil {
		panic(err)
	}

	d := data.NewDataForPost(t.cfg, t.auth.AuthRequest(req), t.service, p, rsp.Comments)

	var name string
	if p.Type == `tweet` {
		name = `tweet.html`
	} else {
		name = `post.html`
	}
	t.executeTemplate(name, w, d)
}

func (t *Theme) QueryByTags(w http.ResponseWriter, req *http.Request, tags []string) {
	d := data.NewDataForTag(t.cfg, t.auth.AuthRequest(req), t.service, tags)
	t.executeTemplate(`tag.html`, w, d)
}

func (t *Theme) QueryFile(w http.ResponseWriter, req *http.Request, postID int64, file string) {
	fs, err := t.service.FileSystemForPost(req.Context(), postID)
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
	// TODO: 改成 ServeFileFS
	http.ServeContent(w, req, stat.Name(), stat.ModTime(), fp)
}

func (t *Theme) userMustCanSeePost(req *http.Request, post *protocols.Post) {
	user := t.auth.AuthRequest(req)
	if user.IsGuest() && post.Status != "public" {
		panic(status.Error(codes.PermissionDenied, "你无权限查看此文章。"))
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
	// 正式环境也不要缓存太久，因为博客在经常更新。
	w.Header().Add(`Cache-Control`, cacheControl)
	// TODO embed 没有 last modified
	http.ServeFileFS(w, req, t.rootFS, file)
}
