package sitemap

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/movsb/taoblog/modules/auth"
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

	auther *auth.Auth
	tmpl   *template.Template
	client *clients.ProtoClient
	impl   service.ToBeImplementedByRpc
}

// New ...
func New(auther *auth.Auth, client *clients.ProtoClient, impl service.ToBeImplementedByRpc) http.Handler {
	s := &Sitemap{
		auther: auther,
		client: client,
		impl:   impl,
		tmpl:   template.Must(template.New(`sitemap`).Parse(tmpl)),
	}

	return s
}

// ServeHTTP ...
func (s *Sitemap) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rsp, err := s.impl.ListAllPostsIds(req.Context()) // 这个 Context 是被 Server 的 Interceptor 传递了 AuthContext 的。
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := s.auther.NewContextForRequestAsGateway(req)

	info := utils.Must1(s.client.Blog.GetInfo(ctx, &proto.GetInfoRequest{}))

	rssArticles := make([]*Article, 0, len(rsp))
	for _, article := range rsp {
		rssArticle := Article{
			Link: fmt.Sprintf("%s/%d/", info.Home, article),
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
