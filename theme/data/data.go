package data

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"reflect"
	"strings"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols/go/proto"
)

// Data holds all data for rendering the site.
type Data struct {
	Context context.Context
	svc     proto.TaoBlogServer

	// all configuration.
	// It's not safe to export all configs outside.
	Config *config.Config

	// The response writer.
	Writer io.Writer

	// The template
	Template *template.Template

	// Metadata
	Meta *MetaData

	// If it is home page.
	Home *HomeData

	// If it is a post (or page).
	Post *PostData

	// If it is the Search.
	Search *SearchData

	// If it is the Posts.
	Posts *PostsData

	// If it is the tag.
	Tag *TagData

	// 碎碎念、叽叽喳喳🦜
	Tweets *TweetsData

	Error *ErrorData

	Partials []any
}

func (d *Data) Info() (*proto.GetInfoResponse, error) {
	if d.Context == nil {
		d.Context = auth.GuestContext(context.TODO())
	}
	return d.svc.GetInfo(d.Context, &proto.GetInfoRequest{})
}

func (d *Data) ExecutePartial(t *template.Template, partial any) error {
	d.Partials = append(d.Partials, partial)
	defer func() {
		d.Partials = d.Partials[:len(d.Partials)-1]
	}()
	return t.Execute(d.Writer, d)
}

func (d *Data) Partial() (any, error) {
	if len(d.Partials) > 0 {
		return d.Partials[len(d.Partials)-1], nil
	}
	return nil, fmt.Errorf(`没有部分模板的数据可用。`)
}

func (d *Data) Title() string {
	if d.Meta != nil && d.Meta.Title != "" {
		return fmt.Sprintf(`%s - %s`, d.Meta.Title, d.Config.Site.Name)
	}
	return d.Config.Site.Name
}

func (d *Data) SiteName() string {
	return d.Config.Site.Name
}

func (d *Data) TweetName() string {
	return fmt.Sprintf(`%s`, TweetName)
}

func (d *Data) BodyClass() string {
	c := []string{}
	if d.Post != nil {
		if d.Post.Post.Wide() {
			c = append(c, `wide`)
		}
		if d.Post.Post.Type == `tweet` {
			c = append(c, `tweet`)
		}
	}
	if d.Tweets != nil {
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

func (d *Data) Strip(obj any) (any, error) {
	user := auth.Context(d.Context).User
	isAdmin := user.IsAdmin()
	switch typed := obj.(type) {
	case *Post:
		if isAdmin {
			return typed.Post, nil
		}
		return &proto.Post{
			Id:       typed.Id,
			Date:     typed.Date,
			Modified: typed.Modified,
		}, nil
	}
	return "", fmt.Errorf(`不知道如何列集：%v`, reflect.TypeOf(obj).String())
}
