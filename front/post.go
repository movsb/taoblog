package front

import (
	"fmt"
	"html"
	"html/template"
	"strings"

	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
)

type Post struct {
	*models.Post
	Content      template.HTML
	RelatedPosts []*models.Post
	server       *service.ImplServer
}

func newPost(post *models.Post, server *service.ImplServer) *Post {
	return &Post{
		Post:    post,
		server:  server,
		Content: template.HTML(post.Content),
	}
}

func newPosts(posts []*models.Post, server *service.ImplServer) []*Post {
	ps := []*Post{}
	for _, p := range posts {
		ps = append(ps, newPost(p, server))
	}
	return ps
}

func (p *Post) DateString() string {
	d := strings.Split(strings.Split(p.Date, " ")[0], "-")
	return fmt.Sprintf("%v年%v月%v日", d[0], d[1], d[2])
}

func (p *Post) ModifiedString() string {
	d := strings.Split(strings.Split(p.Modified, " ")[0], "-")
	return fmt.Sprintf("%v年%v月%v日", d[0], d[1], d[2])
}

func (p *Post) TagsString() template.HTML {
	var ts []string
	for _, t := range p.Tags {
		et := html.EscapeString(t)
		ts = append(ts, fmt.Sprintf(`<a href="/tags/%[1]s">%[1]s</a>`, et))
	}
	return template.HTML(strings.Join(ts, " · "))
}
