package front

import (
	"fmt"
	"html"
	"html/template"
	"strings"

	"github.com/movsb/taoblog/protocols"
)

type Post struct {
	*protocols.Post
	Content      template.HTML
	RelatedPosts []*protocols.PostForRelated
	server       protocols.IServer
}

func newPost(post *protocols.Post, server protocols.IServer) *Post {
	return &Post{
		Post:    post,
		server:  server,
		Content: template.HTML(post.Content),
	}
}

func newPosts(posts []*protocols.Post, server protocols.IServer) []*Post {
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
