package data

import (
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	proto "github.com/movsb/taoblog/protocols"
)

// Data holds all data for rendering the site.
type Data struct {
	svc proto.TaoBlogServer

	// all configuration.
	// It's not safe to export all configs outside.
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

	// If it is a post (or page).
	Post *PostData

	// If it is the Search.
	Search *SearchData

	// If it is the Posts.
	Posts *PostsData

	// If it is the Tags.
	Tags *TagsData

	// If it is the tag.
	Tag *TagData

	// 碎碎念、叽叽喳喳🦜
	Tweets *TweetsData

	Error *ErrorData
}

func (d *Data) Title() string {
	if d.Meta != nil && d.Meta.Title != "" {
		return fmt.Sprintf(`%s - %s`, d.Meta.Title, d.Config.Site.Name)
	}
	return d.Config.Site.Name
}

func (d *Data) SiteName() string {
	return d.Config.Site.Name
}

func (d *Data) TweetName() string {
	return fmt.Sprintf(`%s的%s`, d.Config.Comment.Author, TweetName)
}

func (d *Data) BodyClass() string {
	c := []string{}
	if d.Post != nil {
		if d.Post.Post.Wide() {
			c = append(c, `wide`)
		}
	}
	if d.Tweets != nil {
		c = append(c, `wide`)
	}
	return strings.Join(c, ` `)
}

func (d *Data) Author() string {
	if d.Config.Comment.Author != `` {
		return d.Config.Comment.Author
	}
	return ``
}

// MetaData ...
type MetaData struct {
	Title string // 实际上应该为站点标题，但是好像成了文章标题？
}

type ErrorData struct {
	Message string
}
