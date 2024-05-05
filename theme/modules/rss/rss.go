package rss

// https://validator.w3.org/feed/check.cgi?url=https%3A%2F%2Fblog.twofei.com%2Frss#l9

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

//go:embed rss.xml
var tmpl string

// Article ...
type Article struct {
	*protocols.Post
	Date    string
	Content template.HTML
	Link    string
	GUID    string
}

// RSS ...
type RSS struct {
	config _Config

	Name        string
	Description string
	Link        string
	Articles    []*Article

	tmpl *template.Template
	svc  *service.Service
	auth *auth.Auth
}

type Option func(r *RSS)

func WithArticleCount(n int) Option {
	return func(r *RSS) {
		r.config.articleCount = n
	}
}

type _Config struct {
	articleCount int
}

// New ...
func New(svc *service.Service, auth *auth.Auth, options ...Option) *RSS {
	r := &RSS{
		config: _Config{
			articleCount: 10,
		},

		Name:        svc.Name(),
		Description: svc.Description(),
		Link:        svc.HomeURL(),
		svc:         svc,
		auth:        auth,
	}

	for _, opt := range options {
		opt(r)
	}

	r.tmpl = template.Must(template.New(`rss`).Parse(tmpl))

	return r
}

// ServeHTTP ...
func (r *RSS) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	articles := r.svc.GetLatestPosts(
		req.Context(),
		"id,title,date",
		int64(r.config.articleCount),
		true,
	)

	rssArticles := make([]*Article, 0, len(articles))
	for _, article := range articles {
		rssArticle := Article{
			Post:    article,
			Date:    time.Unix(int64(article.Date), 0).Local().Format(time.RFC1123),
			Content: template.HTML(cdata(article.Content)),
			Link:    fmt.Sprintf("%s/%d/", r.svc.HomeURL(), article.Id),
		}
		rssArticles = append(rssArticles, &rssArticle)
	}

	cr := *r
	cr.Articles = rssArticles

	w.Header().Set("Content-Type", "application/xml")
	fmt.Fprintln(w, `<?xml version="1.0" encoding="UTF-8"?>`)

	if err := cr.tmpl.Execute(w, cr); err != nil {
		log.Println("failed to write rss", err)
	}
}

func cdata(s string) string {
	s = strings.Replace(s, "]]>", "]]]]><!CDATA[>", -1)
	return "<![CDATA[" + s + "]]>"
}
