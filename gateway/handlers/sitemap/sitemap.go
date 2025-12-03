package sitemap

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/movsb/taoblog/modules/auth/user"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
)

//go:embed sitemap.xml
var tmpl string

// Article ...
type Article struct {
	// https://developers.google.com/search/docs/crawling-indexing/sitemaps/build-sitemap#general-guidelines
	// 使用绝对链接。
	Link string
}

// Sitemap ...
type Sitemap struct {
	Articles []*Article

	tmpl   *template.Template
	client *clients.ProtoClient
	impl   service.ToBeImplementedByRpc
}

func New(client *clients.ProtoClient, impl service.ToBeImplementedByRpc) http.Handler {
	s := &Sitemap{
		client: client,
		impl:   impl,
		tmpl:   template.Must(template.New(`sitemap`).Parse(tmpl)),
	}

	return s
}

func (s *Sitemap) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rsp, err := s.impl.ListAllPostsIds(user.SystemForLocal(req.Context()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	info := utils.Must1(s.client.Blog.GetInfo(context.Background(), &proto.GetInfoRequest{}))
	home, _ := url.Parse(info.Home)

	rssArticles := make([]*Article, 0, len(rsp))
	for _, article := range rsp {
		rssArticle := Article{
			Link: home.JoinPath(fmt.Sprintf(`/%d/`, article)).String(),
		}
		rssArticles = append(rssArticles, &rssArticle)
	}

	cs := *s
	cs.Articles = rssArticles

	w.Header().Set("Content-Type", "application/xml")
	fmt.Fprintln(w, `<?xml version="1.0" encoding="UTF-8"?>`)

	if err := cs.tmpl.Execute(w, cs); err != nil {
		panic(err)
	}
}
