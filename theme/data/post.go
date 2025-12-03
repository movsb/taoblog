package data

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/movsb/taoblog/modules/globals"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
	"github.com/movsb/taoblog/service/models"

	wgs2gcj "github.com/googollee/eviltransform/go"
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

func (d *PostData) TOC() template.HTML {
	// GetPost 的时候已经根据喜好决定是否输出目录了。
	return template.HTML(d.Post.Toc)
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
	// NOTE: 这个  marshaller 会把 < > 给转义了，其实没必要。
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
		User:    user.Context(ctx).User,
		Meta: MetaData{
			Title: post.Title,
		},
	}
	p := &PostData{
		Post: newPost(post),
	}
	d.Data = p
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
	case `private`:
		return `[私密]`
	case `partial`:
		return `[部分可见]`
	case `draft`:
		return `[草稿]`
	default:
		panic(`unknown post status`)
	}
}

func (p *Post) EntryClass() string {
	var s []string
	if p.Metas.TextIndent {
		s = append(s, `auto-indent`)
	}
	return strings.Join(s, " ")
}

func (p *Post) IsPublic() bool  { return p.Post.Status == models.PostStatusPublic }
func (p *Post) IsPartial() bool { return p.Post.Status == models.PostStatusPartial }
func (p *Post) IsPrivate() bool { return p.Post.Status == models.PostStatusPrivate }
func (p *Post) IsDraft() bool   { return p.Post.Status == models.PostStatusDraft }

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
<geo-link name="{{.Name}}" wgs84="{{.WGS84Array}}" gcj02="{{.GCJ02Array}}"></geo-link>
`))

type _GeoData struct {
	Name string
	// [latitude,longitude]
	wgs84 [2]float32
	gcj02 [2]float32
}

func (g _GeoData) WGS84Array() string {
	return fmt.Sprintf(`[%f,%f]`, g.wgs84[0], g.wgs84[1])
}
func (g _GeoData) GCJ02Array() string {
	return fmt.Sprintf(`[%f,%f]`, g.gcj02[0], g.gcj02[1])
}

func geoData(g *proto.Metas_Geo) _GeoData {
	lat, lng := wgs2gcj.WGStoGCJ(float64(g.Latitude), float64(g.Longitude))
	return _GeoData{
		Name:  g.Name,
		wgs84: [2]float32{g.Latitude, g.Longitude},
		gcj02: [2]float32{float32(lat), float32(lng)},
	}
}

func (d *Post) GeoElement() template.HTML {
	g := d.GetMetas().GetGeo()
	if g != nil && !g.Private && g.Longitude != 0 && g.Latitude != 0 && g.Name != "" {
		var buf bytes.Buffer
		if err := geoTmpl.Execute(&buf, geoData(g)); err != nil {
			log.Println(err)
		}
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
