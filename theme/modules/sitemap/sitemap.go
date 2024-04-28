package sitemap

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

//go:embed sitemap.xml
var tmpl string

// Article ...
type Article struct {
	*protocols.Post
	Link string
}

// Sitemap ...
type Sitemap struct {
	Articles []*Article

	tmpl *template.Template
	svc  *service.Service
	auth *auth.Auth
}

// New ...
func New(svc *service.Service, auth *auth.Auth) *Sitemap {
	s := &Sitemap{
		svc:  svc,
		auth: auth,
		tmpl: template.Must(template.New(`sitemap`).Parse(tmpl)),
	}

	return s
}

// ServeHTTP ...
func (s *Sitemap) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	user := s.auth.AuthRequest(req)

	articles := s.svc.MustListPosts(
		user.Context(context.TODO()),
		&protocols.ListPostsRequest{
			Fields:  `id`,
			OrderBy: `date DESC`,
		},
	)

	rssArticles := make([]*Article, 0, len(articles))
	for _, article := range articles {
		rssArticle := Article{
			Post: article,
			Link: fmt.Sprintf("%s/%d/", s.svc.HomeURL(), article.Id),
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
