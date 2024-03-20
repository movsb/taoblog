package data

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

// SearchData ...
type SearchData struct {
	Initialized bool
	Posts       []*SearchPost
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
	q := r.URL.Query().Get(`q`)
	d := &Data{
		Config: cfg,
		User:   user,
		Meta: &MetaData{
			Title: fmt.Sprintf("%s - 搜索结果", q),
		},
	}

	rsp, err := service.SearchPosts(context.TODO(), &protocols.SearchPostsRequest{Search: q})
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
		Initialized: rsp.Initialized,
		Posts:       posts2,
	}

	return d
}
