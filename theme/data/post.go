package data

import (
	"fmt"
	"html"
	"html/template"
	"strings"
	"time"

	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
)

// PostData ...
type PostData struct {
	Post *Post
}

// NewDataForPost ...
func NewDataForPost(cfg *config.Config, user *auth.User, service *service.Service, post *protocols.Post) *Data {
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
	return d
}

// Post ...
type Post struct {
	*protocols.Post
	ID      int64
	Content template.HTML
	Related []*models.PostForRelated
	Metas   map[string]string
	link    string
}

func newPost(post *protocols.Post) *Post {
	p := &Post{
		Post:    post,
		ID:      post.Id,
		Content: template.HTML(post.Content),
		Metas:   post.Metas,
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
		return `[未公开发表] `
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

// CustomHeader ...
func (p *Post) CustomHeader() (header string) {
	if i, ok := p.Metas["header"]; ok {
		header = i
	}
	return
}

// CustomFooter ...
func (p *Post) CustomFooter() (footer string) {
	if i, ok := p.Metas["footer"]; ok {
		footer = i
	}
	return
}

func (p *Post) Outdated() bool {
	if value, ok := p.Metas[`outdated`]; ok && (value == `true` || value == `1`) {
		return true
	}
	return false
}

// 是否开启宽屏？
func (p *Post) Wide() bool {
	if value, ok := p.Metas[`wide`]; ok && (value == `true` || value == `1`) {
		return true
	}
	return false
}
