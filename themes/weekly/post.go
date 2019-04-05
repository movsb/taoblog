package weekly

import (
	"fmt"
	"html/template"

	"github.com/movsb/taoblog/protocols"

	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
)

// Post ...
type Post struct {
	*protocols.Post
	Content      template.HTML
	RelatedPosts []*models.PostForRelated
	service      *service.Service
}

func newPost(post *protocols.Post, service *service.Service) *Post {
	return &Post{
		Post:    post,
		service: service,
		Content: template.HTML(post.Content),
	}
}

func newPosts(posts []*protocols.Post, service *service.Service) []*Post {
	ps := []*Post{}
	for _, p := range posts {
		ps = append(ps, newPost(p, service))
	}
	return ps
}

// Link ...
func (p *Post) Link() string {
	return fmt.Sprintf("/%s/", p.Slug)
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
