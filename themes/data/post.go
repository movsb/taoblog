package data

import (
	"fmt"
	"html"
	"html/template"
	"strings"

	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
)

// Post ...
type Post struct {
	*protocols.Post
	Content template.HTML
	Related []*models.PostForRelated
	Tags    []string
}

func newPost(post *protocols.Post) *Post {
	return &Post{
		Post:    post,
		Content: template.HTML(post.Content),
	}
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
	return fmt.Sprintf("/%d/", p.ID)
}

// DateString ...
func (p *Post) DateString() string {
	d := strings.Split(strings.Split(p.Date, " ")[0], "-")
	return fmt.Sprintf("%v年%v月%v日", d[0], d[1], d[2])
}

// ModifiedString ...
func (p *Post) ModifiedString() string {
	d := strings.Split(strings.Split(p.Modified, " ")[0], "-")
	return fmt.Sprintf("%v年%v月%v日", d[0], d[1], d[2])
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
