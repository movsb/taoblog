package data

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"html/template"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/movsb/taoblog/modules/globals"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
)

// PostData ...
type PostData struct {
	Post     *Post
	Comments []*proto.Comment
}

type Comment struct {
	*proto.Comment
}

func (c *Comment) Text() string {
	if c.Source != `` {
		return c.Source
	}
	return c.Content
}

type LatestCommentsByPost struct {
	Post     *Post
	Comments []*Comment
}

func (d *PostData) CommentsAsJsonArray() template.JS {
	if d.Comments == nil {
		d.Comments = make([]*proto.Comment, 0)
	}

	buf := bytes.NewBuffer(nil)

	// 简单格式化一下。
	// [
	//      {...},
	//      {...},
	//      {...},
	// ]
	buf.WriteString("[\n")
	// TODO 这个  marshaller 会把 < > 给转义了，其实没必要。
	encoder := jsonpb.Marshaler{
		OrigName: true,
	}
	for _, c := range d.Comments {
		encoder.Marshal(buf, c)
		buf.WriteString(",\n")
	}
	buf.WriteString("]")
	return template.JS(buf.String())
}

// NewDataForPost ...
func NewDataForPost(ctx context.Context, service proto.TaoBlogServer, post *proto.Post) *Data {
	d := &Data{
		Context: ctx,
		Meta: MetaData{
			Title: post.Title,
		},
	}
	p := &PostData{
		Post: newPost(post),
	}
	d.Post = p
	p.Comments = p.Post.CommentList
	return d
}

// Post ...
type Post struct {
	*proto.Post
	ID      int64
	Content template.HTML
}

func newPost(post *proto.Post) *Post {
	p := &Post{
		Post:    post,
		ID:      post.Id,
		Content: template.HTML(post.Content),
	}
	if p.Metas == nil {
		p.Metas = &proto.Metas{}
	}
	return p
}

// 返回文章的公开状态字符串。
func (p *Post) StatusString() string {
	switch p.Post.Status {
	case ``:
		panic(`post.Status empty`)
	case `public`:
		return ``
	case `draft`:
		return `[私密] `
	default:
		panic(`unknown post status`)
	}
}

func (p *Post) IsPrivate() bool {
	return p.Post.Status == models.PostStatusPrivate
}

func (p *Post) CommentString() string {
	if p.Comments == 0 {
		return `没有评论`
	}
	return fmt.Sprintf(`%d 条评论`, p.Comments)
}

func (p *Post) ShortDateString() string {
	t := time.Unix(int64(p.Date), 0).In(globals.LoadTimezoneOrDefault(p.DateTimezone, time.Local))
	y, m, d := t.Date()
	return fmt.Sprintf(`%d年%02d月%02d日`, y, m, d)
}

func (p *Post) DateString() string {
	t := time.Unix(int64(p.Date), 0).In(globals.LoadTimezoneOrDefault(p.DateTimezone, time.Local))
	return t.Format(time.RFC3339)
}

func (p *Post) ModifiedString() string {
	t := time.Unix(int64(p.Modified), 0).In(globals.LoadTimezoneOrDefault(p.ModifiedTimezone, time.Local))
	return t.Format(time.RFC3339)
}

// TagsString ...
func (p *Post) TagsString() template.HTML {
	var ts []string
	for _, t := range p.Tags {
		et := html.EscapeString(t)
		ts = append(ts, fmt.Sprintf(`<a href="/tags/%[1]s">%[1]s</a>`, et))
	}
	return template.HTML(strings.Join(ts, " · "))
}

func (p *Post) Outdated() bool {
	return p.Metas.Outdated
}

// 是否开启宽屏？
func (p *Post) Wide() bool {
	return p.Metas.Wide
}

var geoTmpl = template.Must(template.New(`geo`).Parse(`
<geo-link longitude="{{.Longitude}}" latitude="{{.Latitude}}">{{.Name}}</geo-link>
`))

func (d *Post) GeoElement() template.HTML {
	g := d.GetMetas().GetGeo()
	if g != nil && g.Longitude > 0 && g.Latitude > 0 && g.Name != "" {
		var buf bytes.Buffer
		geoTmpl.Execute(&buf, g)
		return template.HTML(buf.String())
	}
	return ``
}

func (d *Post) OriginClass() string {
	if o := d.Origin(); o != nil {
		return fmt.Sprintf(`origin-%s`, strings.ToLower(o.PlatformString()))
	}
	return ""
}

type Origin proto.Metas_Origin

func (o *Origin) PlatformString() string {
	switch o.Platform {
	case proto.Metas_Origin_Twitter:
		return `Twitter`
	}
	return ""
}
func (o *Origin) URL() string {
	switch o.Platform {
	case proto.Metas_Origin_Twitter:
		id := o.Slugs[0]
		return fmt.Sprintf(`https://twitter.com/twitter/status/%v`, id)
	}
	return ""
}

func (d *Post) Origin() *Origin {
	if d.Metas != nil && d.Metas.Origin != nil {
		o := (*Origin)(d.Metas.Origin)
		if o.PlatformString() != "" {
			return o
		}
	}
	return nil
}
