package data

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

// RssPostData ...
type RssPostData struct {
	*protocols.Post
	Date    string
	Content template.HTML
	Link    string
}

// RssData ...
type RssData struct {
	Posts []*RssPostData
}

// NewDataForRss ...
func NewDataForRss(cfg *config.Config, user *auth.User, service *service.Service) *Data {
	d := &Data{
		Config: cfg,
		User:   user,
		Meta:   &MetaData{},
	}

	rd := &RssData{}

	posts := service.GetLatestPosts(user.Context(nil), "id,title,date,content", 10)
	rssPosts := make([]*RssPostData, 0, len(posts))
	for _, post := range posts {
		content := template.HTML("<![CDATA[" + strings.Replace(string(post.Content), "]]>", "]]]]><!CDATA[>", -1) + "]]>")
		link := fmt.Sprintf("%s/%d/", service.HomeURL(), post.Id)
		rssPosts = append(rssPosts, &RssPostData{
			Post:    post,
			Date:    datetime.My2Feed(datetime.Proto2My(*post.Date)),
			Content: content,
			Link:    link,
		})
	}

	rd.Posts = rssPosts
	d.Rss = rd
	return d
}
