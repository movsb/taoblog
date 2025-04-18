package theme

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/theme/blog"
	"github.com/movsb/taoblog/theme/data"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
	"github.com/movsb/taoblog/theme/modules/handle304"
	"github.com/movsb/taoblog/theme/modules/sass"
	"github.com/movsb/taoblog/theme/share/variables"
	"github.com/movsb/taorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Theme ...
type Theme struct {
	ctx context.Context

	rootFS fs.FS
	tmplFS fs.FS

	cfg *config.Config

	// NOTE：这是进程内直接调用的。
	// 如果改成连接，需要考虑 metadata 转发问题。
	service  proto.TaoBlogServer
	impl     service.ToBeImplementedByRpc
	searcher proto.SearchServer
	auth     *auth.Auth

	postFileHandler theme_fs.PostFileHandler

	templates *utils.TemplateLoader

	// 主题的变化应该贡献给 304.
	// Git 在本地是 head，但是会随时修改主题，
	// 所以 git 不够用，或者说已经没作用。
	themeChangedAt time.Time

	specialMux *http.ServeMux

	incViewDebouncer *_IncViewDebouncer
}

func New(ctx context.Context, devMode bool, cfg *config.Config, service proto.TaoBlogServer, impl service.ToBeImplementedByRpc, searcher proto.SearchServer, auth *auth.Auth, postFileHandler theme_fs.PostFileHandler) *Theme {
	var rootFS, tmplFS fs.FS

	if devMode {
		dir := blog.SourceRelativeDir
		rootFS = utils.NewOSDirFS(dir.Join(`statics`))
		tmplFS = utils.NewOSDirFS(dir.Join(`templates`))
		sass.WatchAsync(dir.Join(`styles`), `style.scss`, `../statics/style.css`)
	} else {
		// TODO 硬编码成 blog 了。
		rootFS = utils.Must1(fs.Sub(blog.Root, `statics`))
		tmplFS = utils.Must1(fs.Sub(blog.Root, `templates`))
	}

	t := &Theme{
		ctx: ctx,

		rootFS: rootFS,
		tmplFS: tmplFS,

		cfg:      cfg,
		service:  service,
		impl:     impl,
		searcher: searcher,
		auth:     auth,

		postFileHandler: postFileHandler,
		themeChangedAt:  time.Now(),

		specialMux: http.NewServeMux(),
	}

	t.incViewDebouncer = NewIncViewDebouncer(ctx, impl.IncrementViewCount)

	m := t.specialMux

	m.HandleFunc(`GET /search`, t.querySearch)

	// TODO:  严重：切换登录后会显示上一个用户的全部文章
	// 一个解决办法是在 etag 里面加上用户编号。
	// 以及必须验证缓存。
	m.Handle(`GET /posts`, t.lastPostTime304HandlerFunc(t.queryPosts))
	m.Handle(`GET /tweets`, t.lastPostTime304HandlerFunc(t.queryTweets))

	t.loadTemplates()

	variables.SetConfig(&cfg.Theme.Variables)

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
	t.templates = utils.NewTemplateLoader(t.tmplFS, t.funcs(), func() {
		t.themeChangedAt = time.Now()
	})
}

func (t *Theme) executeTemplate(name string, w io.Writer, d *data.Data) {
	t2 := t.templates.GetNamed(name)
	if t2 == nil {
		panic(`未找到模板：` + name)
	}
	d.SetWriterAndTemplate(w, t2)
	if err := t2.Execute(w, d); err != nil {
		log.Println("\033[31m", err, "\033[m")
	}
}

func (t *Theme) Exception(w http.ResponseWriter, req *http.Request, e any) bool {
	if err, ok := e.(error); ok {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.PermissionDenied:
				w.WriteHeader(http.StatusForbidden)
				t.executeTemplate(`error.html`, w, &data.Data{
					Context: req.Context(),
					Data: &data.ErrorData{
						Message: "你无权查看此内容：" + st.Message(),
					},
				})
				return true
			case codes.NotFound:
				w.WriteHeader(http.StatusNotFound)
				t.executeTemplate(`error.html`, w, &data.Data{
					Context: req.Context(),
					Data: &data.ErrorData{
						Message: `你查看的内容不存在。`,
					},
				})
				return true
			}
		}
		if taorm.IsNotFoundError(err) {
			w.WriteHeader(http.StatusNotFound)
			t.executeTemplate(`error.html`, w, &data.Data{
				Context: req.Context(),
				Data: &data.ErrorData{
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
	d := data.NewDataForHome(req.Context(), t.service, t.impl)
	t.executeTemplate(`home.html`, w, d)
	return nil
}

func (t *Theme) querySearch(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForSearch(r.Context(), t.service, t.searcher, r)
	t.executeTemplate(`search.html`, w, d)
}

func (t *Theme) ChangedAt() time.Time {
	return t.themeChangedAt
}

////////////////////////////////////////////////////////////////////////////////

func (t *Theme) lastPostTime304HandlerFunc(h http.HandlerFunc) http.Handler {
	return t.lastPostTime304Handler(h)
}

// NOTE: user.id: 用于区别同一页面对不同用户的缓存行为。
func (t *Theme) lastPostTime304Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := utils.Must1(t.service.GetInfo(r.Context(), &proto.GetInfoRequest{}))
		ac := auth.Context(r.Context())
		h3 := handle304.New(nil,
			handle304.WithNotModified(time.Unix(int64(info.LastPostedAt), 0)),
			handle304.WithEntityTag(version.GitCommit, t.ChangedAt, info.LastPostedAt, ac.User.ID),
		)
		if h3.Match(w, r) {
			return
		}
		h3.Respond(w)
		handle304.MustRevalidate(w)
		h.ServeHTTP(w, r)
	})
}

func (t *Theme) queryPosts(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForPosts(r.Context(), t.service, t.impl, r)
	t.executeTemplate(`posts.html`, w, d)
}

func (t *Theme) queryTweets(w http.ResponseWriter, r *http.Request) {
	d := data.NewDataForTweets(r.Context(), t.service)
	t.executeTemplate(`tweets.html`, w, d)
}

////////////////////////////////////////////////////////////////////////////////

func (t *Theme) post304Handler(w http.ResponseWriter, r *http.Request, p *proto.Post) (handle304.BundleHandler, bool) {
	ac := auth.Context(r.Context())
	h3 := handle304.New(nil,
		handle304.WithNotModified(time.Unix(int64(p.Modified), 0)),
		handle304.WithEntityTag(version.GitCommit, t.ChangedAt, p.Modified, p.LastCommentedAt, ac.User.ID),
	)
	if h3.Match(w, r) {
		return h3, true
	}
	return h3, false
}

func (t *Theme) QueryByID(w http.ResponseWriter, r *http.Request, id int64) {
	p, err := t.service.GetPost(r.Context(),
		&proto.GetPostRequest{
			Id: int32(id),
			GetPostOptions: &proto.GetPostOptions{
				WithRelates:    true,
				WithLink:       proto.LinkKind_LinkKindRooted,
				ContentOptions: co.For(co.QueryByID),
				WithComments:   true,
				WithToc:        1,
			},
		},
	)
	if err != nil {
		panic(err)
	}

	h3, handled := t.post304Handler(w, r, p)
	if handled {
		return
	}

	if p.Type == `page` {
		link := t.impl.GetLink(id)
		// 因为只处理了一层页面路径，所以要判断一下。
		if link != t.impl.GetPlainLink(id) {
			u := *r.URL
			u.Path = link
			http.Redirect(w, r, u.String(), http.StatusPermanentRedirect)
			return
		}
		return
	}

	t.incViewDebouncer.Increase(int(p.Id))
	h3.Respond(w)
	t.tempRenderPost(w, r, p)
}

func (t *Theme) QueryByPage(w http.ResponseWriter, r *http.Request, path string) (int64, error) {
	p, err := t.service.GetPost(r.Context(),
		&proto.GetPostRequest{
			Page: path,
			GetPostOptions: &proto.GetPostOptions{
				WithRelates:    false, // 页面总是不是显示相关文章。
				WithLink:       proto.LinkKind_LinkKindRooted,
				ContentOptions: co.For(co.QueryByPage),
				WithComments:   true,
				WithToc:        1,
			},
		},
	)
	if err != nil {
		panic(err)
	}

	h3, handled := t.post304Handler(w, r, p)
	if handled {
		return p.Id, nil
	}

	t.incViewDebouncer.Increase(int(p.Id))
	h3.Respond(w)
	t.tempRenderPost(w, r, p)

	return p.Id, nil
}

// TODO 304 不要放这里处理。
func (t *Theme) tempRenderPost(w http.ResponseWriter, req *http.Request, p *proto.Post) {
	d := data.NewDataForPost(req.Context(), t.service, p)

	var name string
	if p.Type == `tweet` {
		name = `tweet.html`
	} else {
		name = `post.html`
	}
	t.executeTemplate(name, w, d)
}

func (t *Theme) QueryByTags(w http.ResponseWriter, req *http.Request, tags []string) {
	d := data.NewDataForTag(req.Context(), t.impl, tags)
	t.executeTemplate(`tag.html`, w, d)
}

// 注意：file 不以 / 开头。
// TODO 添加权限测试用例。
func (t *Theme) QueryFile(w http.ResponseWriter, req *http.Request, postID int64, file string) {
	// 权限检查
	utils.Must1(t.service.GetPost(req.Context(), &proto.GetPostRequest{Id: int32(postID)}))

	// 所有人禁止访问特殊文件：以 . 或者 _ 开头的文件或目录。
	// TODO：以及 config.yaml | README.md
	switch file[0] {
	case '.', '_':
		panic(status.Error(codes.PermissionDenied, `尝试访问不允许访问的文件。`))
	}
	switch path.Base(file)[0] {
	case '.', '_':
		panic(status.Error(codes.PermissionDenied, `尝试访问不允许访问的文件。`))
	}
	// 为了不区分大小写，所以没有用 switch。
	if strings.EqualFold(file, `config.yml`) || strings.EqualFold(file, `config.yaml`) || strings.EqualFold(file, `README.md`) {
		panic(status.Error(codes.PermissionDenied, `尝试访问不允许访问的文件。`))
	}

	t.postFileHandler.HandlePostFile(w, req, int(postID), file)
}

func (t *Theme) QuerySpecial(w http.ResponseWriter, req *http.Request, file string) bool {
	if h, p := t.specialMux.Handler(req); p != "" {
		h.ServeHTTP(w, req)
		return true
	}
	return false
}

// TODO 没有处理错误（比如文件不存在）。
func (t *Theme) QueryStatic(w http.ResponseWriter, req *http.Request, file string) {
	handle304.MustRevalidate(w)
	utils.ServeFSWithAutoModTime(w, req, t.rootFS, file)
}
