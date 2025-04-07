package theme

import (
	"fmt"
	"html/template"
	"time"

	"github.com/movsb/taoblog/theme/data"
	"github.com/xeonx/timeago"
)

func (t *Theme) funcs() map[string]any {
	menustr := createMenus(t.cfg.Menus, false)
	customTheme := t.cfg.Theme.Stylesheets.Render()
	fixedZone := time.FixedZone(`+8`, 8*60*60)

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
			t := time.Unix(int64(s), 0).In(fixedZone)
			r := t.Format(time.RFC3339)
			f := timeago.Chinese.Format(t)
			return template.HTML(fmt.Sprintf(
				`<time class="date" datetime="%s" title="%s" data-unix="%d">%s</time>`,
				r, r, s, f,
			))
		},
		"authorName": func(p *data.Post) string {
			u, err := t.auth.GetUserByID(int64(p.UserId))
			if err != nil {
				panic(err)
			}
			return u.Nickname
		},
		// 服务器开始运行的时间。
		`siteSince`: func() string {
			return t.cfg.Site.Since.String()
		},
		// 服务器已经运行了多少天。
		`siteDays`: func() int {
			return t.cfg.Site.Since.Days()
		},
		// 站点名。
		`siteName`: func() string {
			return t.cfg.Site.Name
		},
		`sitePageTitle`: func(s string) string {
			name := t.cfg.Site.Name
			if s != `` {
				return fmt.Sprintf(`%s - %s`, s, name)
			}
			return name
		},
		`editLinkHTML`: func(d *data.Data) template.HTML {
			if p, ok := d.Data.(*data.PostData); ok {
				if p.Post.UserId == int32(d.User.ID) {
					return template.HTML(fmt.Sprintf(`<span class="edit-button"><a href="/admin/editor?id=%d">编辑</a></span>`, p.Post.ID))
				}
			}
			return ``
		},
	}
}
