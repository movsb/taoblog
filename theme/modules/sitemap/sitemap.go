package sitemap

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/theme/modules/handle304"
)

const tmpl = `
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
{{- /* trim */ -}}
{{ range .Articles }}
	<url><loc>{{ .Link }}</loc></url>
{{- end }}
</urlset>
`

// Article ...
type Article struct {
	*protocols.Post
	Link string
}

// Sitemap ...
type Sitemap struct {
	Articles []*Article

	Config *config.Config

	tmpl *template.Template
	svc  *service.Service
	auth *auth.Auth
}

// New ...
func New(cfg *config.Config, svc *service.Service, auth *auth.Auth) *Sitemap {
	s := &Sitemap{
		Config: cfg,
		svc:    svc,
		auth:   auth,
		tmpl:   template.Must(template.New(`sitemap`).Parse(tmpl)),
	}

	return s
}

// ServeHTTP ...
func (s *Sitemap) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if handle304.ArticleRequest(w, req, s.svc.LastArticleUpdateTime()) {
		return
	}

	user := s.auth.AuthRequest(req)

	articles := s.svc.MustListPosts(
		user.Context(nil),
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

	handle304.ArticleResponse(w, s.svc.LastArticleUpdateTime())

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))

	if err := cs.tmpl.Execute(w, cs); err != nil {
		panic(err)
	}
}
