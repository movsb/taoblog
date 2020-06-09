package data

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"strings"
	"time"

	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
)

// Post ...
type Post struct {
	*protocols.Post
	ID      int64
	Content template.HTML
	Related []*models.PostForRelated
	Metas   map[string]interface{}
}

func newPost(post *protocols.Post) *Post {
	p := &Post{
		Post:    post,
		ID:      post.Id,
		Content: template.HTML(post.Content),
	}
	json.Unmarshal([]byte(post.Metas), &p.Metas)
	return p
}

func newPosts(posts []*protocols.Post) []*Post {
	ps := []*Post{}
	for _, p := range posts {
		ps = append(ps, newPost(p))
	}
	return ps
}

// NonPublic ...
func (p *Post) NonPublic() string {
	switch p.Status {
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
	return fmt.Sprintf("/%d/", p.Id)
}

// DateString ...
func (p *Post) DateString() string {
	t := time.Unix(p.Date.Seconds, 0).Local()
	y, m, d := t.Date()
	return fmt.Sprintf("%d年%d月%d日", y, m, d)
}

// ModifiedString ...
func (p *Post) ModifiedString() string {
	t := time.Unix(p.Modified.Seconds, 0).Local()
	y, m, d := t.Date()
	return fmt.Sprintf("%d年%d月%d日", y, m, d)
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
		if s, ok := i.(string); ok {
			header = s
		}
	}
	return
}

// CustomFooter ...
func (p *Post) CustomFooter() (footer string) {
	if i, ok := p.Metas["footer"]; ok {
		if s, ok := i.(string); ok {
			footer = s
		}
	}
	return
}
