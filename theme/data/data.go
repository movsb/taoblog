package data

import (
	"html/template"
	"io"

	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
)

// Data holds all data for rendering the site.
type Data struct {
	svc *service.Service

	// all configuration.
	Config *config.Config

	// current login user, non-nil.
	User *auth.User

	// The response writer.
	Writer io.Writer

	// The template
	Template *template.Template

	// Metadata
	Meta *MetaData

	// If it is home page.
	Home *HomeData

	// If it is a post.
	Post *PostData

	// If it is the Search.
	Search *SearchData

	// If it is the Posts.
	Posts *PostsData

	// If it is the Tags.
	Tags *TagsData

	// If it is the tag.
	Tag *TagData
}

// MetaData ...
type MetaData struct {
	Title string
}

// PostData ...
type PostData struct {
	Post        *Post
	Pingbacks   []*models.Pingback
	PingbackURL string
}

// NewDataForPost ...
func NewDataForPost(cfg *config.Config, user *auth.User, service *service.Service, post *protocols.Post) *Data {
	d := &Data{
		Config: cfg,
		User:   user,
		Meta: &MetaData{
			Title: post.Title,
		},
	}
	p := &PostData{
		Post:        newPost(post),
		PingbackURL: service.HomeURL() + `/v3/xmlrpc`,
	}
	d.Post = p
	if cfg.Site.ShowRelatedPosts {
		p.Post.Related = service.GetRelatedPosts(post.Id)
	}
	if cfg.Site.ShowPingbacks {
		p.Pingbacks = service.GetPingbacks(post.Id)
	}
	p.Post.Tags = service.GetPostTags(p.Post.Id)
	return d
}
