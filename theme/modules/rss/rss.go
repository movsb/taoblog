package rss

// https://validator.w3.org/feed/check.cgi?url=https%3A%2F%2Fblog.twofei.com%2Frss#l9

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
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
	svc  protocols.TaoBlogServer
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
func New(svc protocols.TaoBlogServer, auth *auth.Auth, options ...Option) *RSS {
	info := utils.Must(svc.GetInfo(context.Background(), &protocols.GetInfoRequest{}))

	r := &RSS{
		config: _Config{
			articleCount: 10,
		},

		Name:        info.Name,
		Description: info.Description,
		Link:        info.Home,
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
	rsp, err := r.svc.ListPosts(req.Context(), &protocols.ListPostsRequest{
		Limit:   int32(r.config.articleCount),
		OrderBy: `date desc`,
		Kinds:   []string{`post`},
		ContentOptions: &protocols.PostContentOptions{
			WithContent:      true,
			RenderCodeBlocks: false,
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rssArticles := make([]*Article, 0, len(rsp.Posts))
	for _, article := range rsp.Posts {
		rssArticle := Article{
			Post:    article,
			Date:    time.Unix(int64(article.Date), 0).Local().Format(time.RFC1123),
			Content: template.HTML(cdata(article.Content)),
			Link:    fmt.Sprintf("%s/%d/", r.Link, article.Id),
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
