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
	"github.com/movsb/taoblog/protocols/clients"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
)

//go:embed rss.xml
var tmpl string

type Article struct {
	*proto.Post
	Date    Date
	Content template.HTML
}

type Date time.Time

func (d Date) String() string {
	return time.Time(d).Format(time.RFC1123)
}

func dateFrom(t int32, l *time.Location) Date {
	return Date(time.Unix(int64(t), 0).In(l))
}

type Data struct {
	Name          string
	Description   string
	Home          string
	LastBuildDate Date
	Articles      []*Article
}

type RSS struct {
	config _Config

	tmpl   *template.Template
	client *clients.ProtoClient

	loc utils.CurrentTimezoneGetter

	// 是否仅列出公开文章，默认是。
	onlyPublic bool
}

type Option func(r *RSS)

func WithArticleCount(n int) Option {
	return func(r *RSS) {
		r.config.articleCount = n
	}
}

func WithCurrentLocationGetter(loc utils.CurrentTimezoneGetter) Option {
	return func(r *RSS) {
		r.loc = loc
	}
}

type _Config struct {
	articleCount int
}

func New(auther *auth.Auth, client *clients.ProtoClient, options ...Option) http.Handler {
	r := &RSS{
		config: _Config{
			articleCount: 10,
		},
		tmpl:   template.Must(template.New(`rss`).Parse(tmpl)),
		client: client,

		onlyPublic: true,
	}

	for _, opt := range options {
		opt(r)
	}

	if r.loc == nil {
		r.loc = utils.LocalTimezoneGetter{}
	}

	return utils.IIF[http.Handler](r.onlyPublic, utils.StripCredentialsHandler(r), r)
}

// ServeHTTP ...
func (r *RSS) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	info := utils.Must1(r.client.Blog.GetInfo(context.Background(), &proto.GetInfoRequest{}))

	data := Data{
		Name:          info.Name,
		Description:   info.Description,
		Home:          info.Home,
		LastBuildDate: dateFrom(info.LastPostedAt, r.loc.GetCurrentTimezone()),
	}

	rsp, err := r.client.Blog.ListPosts(
		// 默认应该是移除了凭证信息的，查看 http.Handler 实现。
		// 否则单测过不了。
		req.Context(),
		&proto.ListPostsRequest{
			Limit:          int32(r.config.articleCount),
			OrderBy:        `date desc`,
			Kinds:          []string{`post`},
			WithLink:       proto.LinkKind_LinkKindFull,
			ContentOptions: co.For(co.Rss),
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rssArticles := make([]*Article, 0, len(rsp.Posts))
	for _, article := range rsp.Posts {
		rssArticle := Article{
			Post: article,
			// TODO 尊重文章本身的时区。
			Date:    dateFrom(article.Date, r.loc.GetCurrentTimezone()),
			Content: template.HTML(cdata(article.Content)),
		}
		rssArticles = append(rssArticles, &rssArticle)
	}

	data.Articles = rssArticles

	w.Header().Set("Content-Type", "application/xml")
	fmt.Fprintln(w, `<?xml version="1.0" encoding="UTF-8"?>`)

	if err := r.tmpl.Execute(w, data); err != nil {
		log.Println("failed to write rss", err)
	}
}

func cdata(s string) string {
	s = strings.Replace(s, "]]>", "]]]]><!CDATA[>", -1)
	return "<![CDATA[" + s + "]]>"
}
