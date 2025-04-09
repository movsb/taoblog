package data

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"reflect"
	"strings"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols/go/proto"
)

// Data holds all data for rendering the site.
type Data struct {
	svc      proto.TaoBlogServer
	writer   io.Writer
	template *template.Template

	Context context.Context
	User    *auth.User

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
	case *PostData:
		return false
	}
	return true
}

func (d *Data) Info() (*proto.GetInfoResponse, error) {
	if d.Context == nil {
		d.Context = auth.GuestForLocal(context.TODO())
	}
	return d.svc.GetInfo(d.Context, &proto.GetInfoRequest{})
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

// TODO 这个函数好像已经没有存在的意义？
func (d *Data) Strip(obj any) (any, error) {
	user := auth.Context(d.Context).User
	switch typed := obj.(type) {
	case *Post:
		if user.ID == int64(typed.UserId) {
			return typed.Post, nil
		}
		return &proto.Post{
			Id:       typed.Id,
			Date:     typed.Date,
			Modified: typed.Modified,
			UserId:   typed.UserId,
		}, nil
	}
	return "", fmt.Errorf(`不知道如何列集：%v`, reflect.TypeOf(obj).String())
}
