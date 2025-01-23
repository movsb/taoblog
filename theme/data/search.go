package data

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/protocols/go/proto"
)

// SearchData ...
type SearchData struct {
	Initialized bool
	Posts       []*SearchPost
}

type SearchPost struct {
	p *proto.SearchPostsResponse_Post
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
func NewDataForSearch(ctx context.Context, cfg *config.Config, service proto.TaoBlogServer, searcher proto.SearchServer, r *http.Request) *Data {
	q := r.URL.Query().Get(`q`)
	d := &Data{
		ctx:    ctx,
		Config: cfg,
		Meta: &MetaData{
			Title: fmt.Sprintf("%s - 搜索结果", q),
		},
	}

	rsp, err := searcher.SearchPosts(ctx, &proto.SearchPostsRequest{Search: q})
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
