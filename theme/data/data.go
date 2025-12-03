package data

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/auth/user"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/xeonx/timeago"
)

// Data holds all data for rendering the site.
type Data struct {
	svc      proto.TaoBlogServer
	writer   io.Writer
	template *template.Template

	Context context.Context
	User    *user.User

	// Metadata
	Meta MetaData

	// 可能是以下之下：
	//
	//  - *HomeData
	//  - *PostData
	//  - *SearchData
	//  - *PostsData
	//  - *TagData
	//  - *TweetsData
	//  - *ErrorData
	Data any

	Partials []any
}

func (d *Data) Execute(name string, alt *template.Template) error {
	tt := d.template.Lookup(name)
	if tt == nil {
		tt = alt
	}
	if tt != nil {
		return tt.Execute(d.writer, d)
	}
	return nil
}

func (d *Data) SetWriterAndTemplate(w io.Writer, t *template.Template) {
	d.writer = w
	d.template = t
}

func (d *Data) ShowHeader() bool {
	switch d.Data.(type) {
	case *PostData, *TweetsData:
		return false
	}
	return true
}

// 暂时全局关闭，排版不太优雅，侧栏数据少，需要目录的文章少。
func (d *Data) ShowAsideRight() bool {
	return false
	n := 0
	switch typed := d.Data.(type) {
	case *PostData:
		if typed.TOC() != `` {
			n++
		}
	}
	return n > 0
}

// 给 header 用于指示页面类型。
func (d *Data) Kind() string {
	switch d.Data.(type) {
	case *PostData:
		return `post`
	case *HomeData:
		return `home`
	}
	return `unknown`
}

type _Info struct {
	proto *proto.GetInfoResponse
}

func (info _Info) ExpiryStatus() string {
	var a []string
	if d := info.proto.CertDaysLeft; d > 0 {
		a = append(a, fmt.Sprintf(`证书有效期剩余 %d 天`, d))
	}
	if d := info.proto.DomainDaysLeft; d > 0 {
		a = append(a, fmt.Sprintf(`域名有效期剩余 %d 天`, d))
	}
	if len(a) == 0 {
		return ``
	}
	return `运维状态：` + strings.Join(a, `，`) + `。`
}

func (info _Info) BackupStatus() string {
	friendly := func(t int32) string {
		tm := time.Unix(int64(t), 0)
		return timeago.Chinese.Format(tm)
	}

	var a []string
	if d := info.proto.LastBackupAt; d > 0 {
		a = append(a, fmt.Sprintf(`上次备份于 %s`, friendly(d)))
	}
	if d := info.proto.LastSyncAt; d > 0 {
		a = append(a, fmt.Sprintf(`上次同步于 %s`, friendly(d)))
	}
	if len(a) == 0 {
		return ``
	}
	return `备份状态：` + strings.Join(a, `，`) + `。`
}

func (info _Info) StorageStatus() string {
	var a []string
	if d := info.proto.GetStorage().GetPosts(); d > 0 {
		a = append(a, fmt.Sprintf(`文章数据库：%s`, utils.ByteCountIEC(d)))
	}
	if d := info.proto.GetStorage().GetFiles(); d > 0 {
		a = append(a, fmt.Sprintf(`文件数据库：%s`, utils.ByteCountIEC(d)))
	}
	if len(a) == 0 {
		return ``
	}
	return `存储状态：` + strings.Join(a, `，`) + `。`
}

func (d *Data) Info() *_Info {
	if d.Context == nil {
		d.Context = user.GuestForLocal(context.Background())
	}
	return &_Info{utils.Must1(d.svc.GetInfo(d.Context, &proto.GetInfoRequest{}))}
}

func (d *Data) ExecutePartial(t *template.Template, partial any) error {
	d.Partials = append(d.Partials, partial)
	defer func() {
		d.Partials = d.Partials[:len(d.Partials)-1]
	}()
	return t.Execute(d.writer, d)
}

func (d *Data) Partial() (any, error) {
	if len(d.Partials) > 0 {
		return d.Partials[len(d.Partials)-1], nil
	}
	return nil, fmt.Errorf(`没有部分模板的数据可用。`)
}

func (d *Data) Title() string {
	return d.Meta.Title
}

func (d *Data) TweetName() string {
	return TweetName
}

func (d *Data) BodyClass() string {
	c := []string{}
	switch typed := d.Data.(type) {
	case *PostData:
		if typed.Post.Wide() {
			c = append(c, `wide`)
		}
		if typed.Post.Type == `tweet` {
			c = append(c, `tweet`)
		}
	case *TweetsData:
		c = append(c, `tweets`)
	}
	return strings.Join(c, ` `)
}

// MetaData ...
type MetaData struct {
	Title string // 实际上应该为站点标题，但是好像成了文章标题？
}

type ErrorData struct {
	Message string
}
