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
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	proto "github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/theme/blog"
	"github.com/movsb/taoblog/theme/data"
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

	cfg *config.Config

	// NOTE：这是进程内直接调用的。
	// 如果改成连接，需要考虑 metadata 转发问题。
	service  proto.TaoBlogServer
	impl     service.ToBeImplementedByRpc
	searcher proto.SearchServer
	auth     *auth.Auth

	rss     *rss.RSS
	sitemap *sitemap.Sitemap

	templates *utils.TemplateLoader

	// 主题的变化应该贡献给 304.
	// Git 在本地是 head，但是会随时修改主题，
	// 所以 git 不够用，或者说已经没作用。
	themeChangedAt time.Time

	specialMux *http.ServeMux
}

func New(devMode bool, cfg *config.Config, service proto.TaoBlogServer, impl service.ToBeImplementedByRpc, searcher proto.SearchServer, auth *auth.Auth) *Theme {
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

		cfg:      cfg,
		service:  service,
		impl:     impl,
		searcher: searcher,
		auth:     auth,

		themeChangedAt: time.Now(),

		specialMux: http.NewServeMux(),
	}

	m := t.specialMux

	if r := cfg.Site.RSS; r.Enabled {
		t.rss = rss.New(service, rss.WithArticleCount(r.ArticleCount))
		m.Handle(`GET /rss`, t.LastPostTime304Handler(t.rss))
	}
	if cfg.Site.Sitemap.Enabled {
		t.sitemap = sitemap.New(service, impl)
		m.Handle(`GET /sitemap.xml`, t.LastPostTime304Handler(t.sitemap))
	}

	m.HandleFunc(`GET /search`, t.querySearch)
	m.Handle(`GET /posts`, t.LastPostTime304HandlerFunc(t.queryPosts))
	m.Handle(`GET /tweets`, t.LastPostTime304HandlerFunc(t.queryTweets))
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

	customTheme := t.cfg.Site.Theme.Stylesheets.Render()

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
		"apply_site_theme_customs": func() template.HTML {
			return template.HTML(customTheme)
		},
	}

	t.templates = utils.NewTemplateLoader(t.tmplFS, funcs, func() {
		t.themeChangedAt = time.Now()
	})
}

func (t *Theme) watchStyles(stylesFS fs.FS) {
	if changed, ok := stylesFS.(utils.FsWithChangeNotify); ok {
		log.Println(`Listening for style changes`)

		bundle := func() {
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

		bundle()

		go func() {
			debouncer := utils.NewDebouncer(time.Second, bundle)
			for event := range changed.Changed() {
				switch event.Op {
				case fsnotify.Create, fsnotify.Remove, fsnotify.Write:
					debouncer.Enter()
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

func (t *Theme) Exception(w http.ResponseWriter, req *http.Request, e any) bool {
	if err, ok := e.(error); ok {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.PermissionDenied:
				w.WriteHeader(http.StatusForbidden)
				t.executeTemplate(`error.html`, w, &data.Data{
					Error: &data.ErrorData{
						Message: `你无权查看此内容。`,
					},
				})
				return true
			case codes.NotFound:
				w.WriteHeader(http.StatusNotFound)
				t.executeTemplate(`error.html`, w, &data.Data{
					Error: &data.ErrorData{
						Message: `你查看的内容不存在。`,
					},
				})
				return true
			}
		}
		if taorm.IsNotFoundError(err) {
			w.WriteHeader(http.StatusNotFound)
			t.executeTemplate(`error.html`, w, &data.Data{
				Error: &data.ErrorData{
					Message: `你查看的内容不存在。`,
				},
			})
			return true
		}
	}
	return false
}

func (t *Theme) ProcessHomeQueries(w http.ResponseWriter, req *http.Request, query url.Values) bool {
	return false
}

func (t *Theme) QueryHome(w http.ResponseWriter, req *http.Request) error {
	d := data.NewDataForHome(req.Context(), t.cfg, t.service, t.impl)
	t.executeTemplate(`home.html`, w, d)
	return nil
}

func (t *Theme) querySearch(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForSearch(t.cfg, t.auth.AuthRequest(r), t.service, t.searcher, r)
	t.executeTemplate(`search.html`, w, d)
}

func (t *Theme) Post304Handler(w http.ResponseWriter, r *http.Request, p *proto.Post) bool {
	h3 := handle304.New(
		handle304.WithNotModified(time.Unix(int64(p.Modified), 0)),
		handle304.WithEntityTag(version.GitCommit, t.impl.ThemeChangedAt, t.ChangedAt, p.Modified, p.LastCommentedAt),
	)
	if h3.Match(w, r) {
		return true
	}
	h3.Response(w)
	handle304.MustRevalidate(w)
	return false
}

func (t *Theme) LastPostTime304HandlerFunc(h http.HandlerFunc) http.Handler {
	return t.LastPostTime304Handler(h)
}

func (t *Theme) ChangedAt() time.Time {
	return t.themeChangedAt
}

func (t *Theme) LastPostTime304Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := utils.Must(t.service.GetInfo(r.Context(), &proto.GetInfoRequest{}))
		h3 := handle304.New(
			handle304.WithNotModified(time.Unix(int64(info.LastPostedAt), 0)),
			handle304.WithEntityTag(version.GitCommit, t.impl.ThemeChangedAt, t.ChangedAt, info.LastPostedAt),
		)
		if h3.Match(w, r) {
			return
		}
		h3.Response(w)
		handle304.MustRevalidate(w)
		h.ServeHTTP(w, r)
	})
}

func (t *Theme) queryPosts(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForPosts(r.Context(), t.cfg, t.service, t.impl, r)
	t.executeTemplate(`posts.html`, w, d)
}

func (t *Theme) queryTweets(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForTweets(r.Context(), t.impl.Config(), t.service)
	t.executeTemplate(`tweets.html`, w, d)
}

func (t *Theme) queryTags(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForTags(t.cfg, t.auth.AuthRequest(r), t.service, t.impl)
	t.executeTemplate(`tags.html`, w, d)
}

func (t *Theme) QueryByID(w http.ResponseWriter, req *http.Request, id int64) {
	post, err := t.service.GetPost(req.Context(),
		&proto.GetPostRequest{
			Id:          int32(id),
			WithRelates: true,
			WithLink:    proto.LinkKind_LinkKindRooted,
			ContentOptions: &proto.PostContentOptions{
				WithContent:       true,
				RenderCodeBlocks:  true,
				UseAbsolutePaths:  false,
				OpenLinksInNewTab: proto.PostContentOptions_OpenLinkInNewTabKindAll,
			},
		},
	)
	if err != nil {
		panic(err)
	}

	if post.Type == `page` {
		link := t.impl.GetLink(id)
		// 因为只处理了一层页面路径，所以要判断一下。
		if link != t.impl.GetPlainLink(id) {
			u := *req.URL
			u.Path = link
			http.Redirect(w, req, u.String(), http.StatusPermanentRedirect)
			return
		}
		return
	}

	t.incView(post.Id)
	t.tempRenderPost(w, req, post)
}

func (t *Theme) incView(id int64) {
	t.impl.IncrementPostPageView(id)
}

func (t *Theme) QueryByPage(w http.ResponseWriter, req *http.Request, path string) (int64, error) {
	post, err := t.service.GetPost(req.Context(),
		&proto.GetPostRequest{
			Page:        path,
			WithRelates: false, // 页面总是不是显示相关文章。
			WithLink:    proto.LinkKind_LinkKindRooted,
			ContentOptions: &proto.PostContentOptions{
				WithContent:       true,
				RenderCodeBlocks:  true,
				UseAbsolutePaths:  false,
				OpenLinksInNewTab: proto.PostContentOptions_OpenLinkInNewTabKindAll,
			},
		},
	)
	if err != nil {
		return 0, err
	}

	t.incView(post.Id)
	t.tempRenderPost(w, req, post)
	return post.Id, nil
}

func (t *Theme) tempRenderPost(w http.ResponseWriter, req *http.Request, p *proto.Post) {
	if t.Post304Handler(w, req, p) {
		return
	}

	rsp, err := t.service.GetPostComments(req.Context(), &proto.GetPostCommentsRequest{Id: p.Id})
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
	d := data.NewDataForTag(req.Context(), t.cfg, t.service, tags)
	t.executeTemplate(`tag.html`, w, d)
}

// TODO 没限制不可访问文章的附件是否不可访问。
// 毕竟，文章不可访问后，文件列表暂时拿不到。
func (t *Theme) QueryFile(w http.ResponseWriter, req *http.Request, postID int64, file string) {
	fs, err := t.impl.FileSystemForPost(req.Context(), postID)
	if err != nil {
		panic(err)
	}
	http.ServeFileFS(w, req, fs, file)
}

func (t *Theme) QuerySpecial(w http.ResponseWriter, req *http.Request, file string) bool {
	if h, p := t.specialMux.Handler(req); p != "" {
		h.ServeHTTP(w, req)
		return true
	}
	return false
}

// TODO 支持本地静态文件以临时存放临时文件。
// TODO 没有处理错误（比较文件不存在）。
func (t *Theme) QueryStatic(w http.ResponseWriter, req *http.Request, file string) {
	if service.DevMode() {
		handle304.MustRevalidate(w)
	} else {
		handle304.CacheShortly(w)
	}
	http.ServeFileFS(w, req, t.rootFS, file)
}
