package rss

// https://validator.w3.org/feed/check.cgi?url=https%3A%2F%2Fblog.twofei.com%2Frss#l9

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/theme/modules/handle304"
)

const tmpl = `
<rss version="2.0">
<channel>
	<title>{{ .Name }}</title>
	<link>{{ .Link }}</link>
	<description>{{ .Description }}</description>
	{{- range .Articles -}}
	<item>
		<title>{{ .Title }}</title>
		<link>{{ .Link }}</link>
		<pubDate>{{ .Date }}</pubDate>
		<description>{{ .Content }}</description>
	</item>
	{{ end }}
</channel>
</rss>
`

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
	Name        string
	Description string
	Link        string
	Articles    []*Article
	Config      *config.Config

	tmpl *template.Template
	svc  *service.Service
	auth *auth.Auth
}

// New ...
func New(cfg *config.Config, svc *service.Service, auth *auth.Auth) *RSS {
	r := &RSS{
		Config:      cfg,
		Name:        svc.Name(),
		Description: svc.Description(),
		Link:        svc.HomeURL(),
		svc:         svc,
		auth:        auth,
	}

	r.tmpl = template.Must(template.New(`rss`).Parse(tmpl))

	return r
}

// ServeHTTP ...
func (r *RSS) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if handle304.ArticleRequest(w, req, r.svc.LastArticleUpdateTime()) {
		return
	}

	user := r.auth.AuthCookie2(req)

	articles := r.svc.GetLatestPosts(
		user.Context(context.TODO()),
		"id,title,date,content",
		int64(r.Config.Site.RSS.ArticleCount),
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

	handle304.ArticleResponse(w, r.svc.LastArticleUpdateTime())

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))

	if err := cr.tmpl.Execute(w, cr); err != nil {
		log.Println("failed to write rss", err)
	}
}

func cdata(s string) string {
	s = strings.Replace(s, "]]>", "]]]]><!CDATA[>", -1)
	return "<![CDATA[" + s + "]]>"
}
