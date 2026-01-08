package theme

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"reflect"
	"sync"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/theme/data"
)

type _OpenGraphData struct {
	Title string
	Image template.URL
	URL   template.URL
}

var tmplOpenGraph = sync.OnceValue(func() *template.Template {
	return template.Must(template.New(`open graph`).Parse(`
<meta property="og:type" content="article">
<meta property="og:title" content="{{.Title}}">
<meta property="og:image" content="{{.Image}}">
<meta property="twitter:card" content="summary_large_image">
<meta property="twitter:title" content="{{.Title}}">
<meta property="twitter:image" content="{{.Image}}">
`))
})

func (t *Theme) funcs() map[string]any {
	menustr := createMenus(t.cfg.Menus, false)
	customTheme := t.cfg.Theme.Stylesheets.Render()

	return map[string]any{
		// https://githut.com/golang/go/issues/14256
		"raw": func(s string) template.HTML {
			return template.HTML(s)
		},
		"menus": func() template.HTML {
			return template.HTML(menustr)
		},
		"render": func(name string, d *data.Data) error {
			return d.Execute(name, t.templates.GetPartial(name))
		},
		"partial": func(name string, data *data.Data, data2 any) error {
			if t := t.templates.GetPartial(name); t != nil {
				return data.ExecutePartial(t, data2)
			}
			// TODO 找不到应该报错。
			return nil
		},
		"apply_site_theme_customs": func() template.HTML {
			return template.HTML(customTheme)
		},
		// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/time
		"friendlyDateTime": func(s int32) template.HTML {
			tz := t.cfg.Site.GetTimezoneLocation()
			now := time.Now().In(tz)
			t := time.Unix(int64(s), 0).In(tz)
			r := t.Format(time.RFC3339)
			f := utils.RelativeDateFrom(t, now)
			return template.HTML(fmt.Sprintf(
				`<time class="date" datetime="%s" title="%s" data-unix="%d">%s</time>`,
				r, r, s, f,
			))
		},
		"authorName": func(p *data.Post) string {
			u, err := t.authFrontend.GetUserByID(int(p.UserId))
			if err != nil {
				panic(err)
			}
			return u.Nickname
		},
		// 服务器开始运行的时间。
		`siteSince`: func() string {
			t := time.Unix(int64(t.cfg.Site.Since), 0)
			return t.Format(`2006年01月02日`)
		},
		// 服务器已经运行了多少天。
		`siteDays`: func() int {
			t := time.Unix(int64(t.cfg.Site.Since), 0)
			return int(time.Since(t).Hours()/24) + 1
		},
		// 站点名。
		`siteName`: func() string {
			return t.cfg.Site.GetName()
		},
		`tmplOpenGraph`: func(p *data.Post) template.HTML {
			home := utils.Must1(url.Parse(t.cfg.Site.GetHome()))

			data := _OpenGraphData{
				Title: p.Title,
				Image: template.URL(home.JoinPath(`/v0/posts`, fmt.Sprint(p.ID), `open_graph.png`).String()),
			}

			buf := bytes.NewBuffer(nil)
			tmplOpenGraph().Execute(buf, data)
			return template.HTML(buf.String())
		},
		`sitePageTitle`: func(d *data.Data) string {
			name := t.cfg.Site.GetName()
			if d.Title() == `` {
				return name
			}

			// 非公开内容不返回 html title，以防止出现在浏览器地址栏的历史记录中（边输入边出现）。
			if p, ok := d.Data.(*data.PostData); ok && p.Post.Status != models.PostStatusPublic {
				return ``
			}

			return fmt.Sprintf(`%s - %s`, d.Title(), name)
		},
		`editLinkHTML`: func(d *data.Data) template.HTML {
			if p, ok := d.Data.(*data.PostData); ok {
				if p.Post.UserId == int32(d.User.ID) {
					return template.HTML(fmt.Sprintf(`<span class="edit-button"><a href="/admin/editor?id=%d">编辑</a></span>`, p.Post.ID))
				}
			}
			return ``
		},
		// TODO 这个函数好像已经没有存在的意义？
		`strip`: func(obj any) (any, error) {
			// user := auth.Context(d.Context).User
			switch typed := obj.(type) {
			case *data.Post:
				return &proto.Post{
					Id:       typed.Id,
					Date:     typed.Date,
					Modified: typed.Modified,
					UserId:   typed.UserId,
				}, nil
			}
			return "", fmt.Errorf(`不知道如何列集：%v`, reflect.TypeOf(obj).String())
		},
		`comments`: func(d *data.Data) template.HTML {
			switch typed := d.Data.(type) {
			case *data.PostData:
				if len(typed.Comments) <= 0 {
					return `<script>TaoBlog.comments=[];__initComments();</script>`
				}
				buf := bytes.NewBuffer(nil)
				// PB 的 json 序列化是不支持数组的。
				buf.WriteString("<script>\nTaoBlog.comments = [\n")
				// NOTE: 这个  marshaller 会把 < > 给转义了，其实没必要。
				encoder := jsonpb.Marshaler{OrigName: true}
				for _, c := range typed.Comments {
					encoder.Marshal(buf, c)
					buf.WriteString(",\n")
				}
				buf.WriteString("];\n")
				buf.WriteString("__initComments();\n</script>")
				return template.HTML(buf.String())
			}
			return ``
		},
	}
}
