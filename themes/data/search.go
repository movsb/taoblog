package data

import (
	"context"
	"html/template"
	"net/http"

	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

// SearchData ...
type SearchData struct {
	Posts []*SearchPost
}

type SearchPost struct {
	p *protocols.SearchPostsResponse_Post
}

func (p *SearchPost) Id() int {
	return int(p.p.Id)
}

func (p *SearchPost) Title() template.HTML {
	return template.HTML(p.p.Title)
}

func (p *SearchPost) Content() template.HTML {
	return template.HTML(p.p.Content)
}

// NewDataForSearch ...
func NewDataForSearch(cfg *config.Config, user *auth.User, service *service.Service, r *http.Request) *Data {
	d := &Data{
		Config: cfg,
		User:   user,
		Meta:   &MetaData{},
	}

	rsp, err := service.SearchPosts(context.TODO(), &protocols.SearchPostsRequest{Search: r.URL.Query().Get(`q`)})
	if err != nil {
		panic(err)
	}

	var posts2 []*SearchPost
	for _, p := range rsp.Posts {
		posts2 = append(posts2, &SearchPost{
			p: p,
		})
	}

	d.Search = &SearchData{
		Posts: posts2,
	}

	return d
}
