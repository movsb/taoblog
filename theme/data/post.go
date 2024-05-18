package data

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
	"github.com/xeonx/timeago"
)

// PostData ...
type PostData struct {
	Post     *Post
	Comments []*protocols.Comment
}

func (d *PostData) CommentsAsJsonArray() template.JS {
	if d.Comments == nil {
		d.Comments = make([]*protocols.Comment, 0)
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
func NewDataForPost(cfg *config.Config, user *auth.User, service *service.Service, post *protocols.Post, comments []*protocols.Comment) *Data {
	d := &Data{
		Config: cfg,
		User:   user,
		Meta: &MetaData{
			Title: post.Title,
		},
	}
	p := &PostData{
		Post: newPost(post),
	}
	d.Post = p
	if cfg.Site.ShowRelatedPosts {
		p.Post.Related = service.GetRelatedPosts(post.Id)
	}
	p.Post.Tags = service.GetPostTags(p.Post.Id)
	p.Post.link = service.GetLink(post.Id)
	p.Comments = comments
	return d
}

// Post ...
type Post struct {
	*protocols.Post
	ID      int64
	Content template.HTML
	Related []*models.PostForRelated
	Metas   models.PostMeta
	link    string
}

func newPost(post *protocols.Post) *Post {
	p := &Post{
		Post:    post,
		ID:      post.Id,
		Content: template.HTML(post.Content),
		Metas:   *models.PostMetaFrom(post.Metas),
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

// Link ...
func (p *Post) Link() string {
	if p.link != "" {
		return p.link
	}
	return fmt.Sprintf("/%d/", p.ID)
}

// DateString ...
func (p *Post) DateString() string {
	t := time.Unix(int64(p.Date), 0).Local()
	y, m, d := t.Date()
	return fmt.Sprintf("%d年%02d月%02d日", y, m, d)
}
func (p *Post) ShortDateString() string {
	return timeago.Chinese.Format(time.Unix(int64(p.Date), 0))
}
func (p *Post) CommentString() string {
	if p.Comments == 0 {
		return `没有评论`
	}
	return fmt.Sprintf(`%d 条评论`, p.Comments)
}

// ModifiedString ...
func (p *Post) ModifiedString() string {
	t := time.Unix(int64(p.Modified), 0).Local()
	y, m, d := t.Date()
	return fmt.Sprintf("%d年%02d月%02d日", y, m, d)
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
	if p.Type == `tweet` {
		return true
	}
	return p.Metas.Wide
}

func (d *Post) GeoString() string {
	g := d.Metas.Geo
	if g != nil && g.Longitude > 0 && g.Latitude > 0 && g.Name != "" {
		return g.Name
	}
	return ""
}
