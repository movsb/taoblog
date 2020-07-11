package models

import (
	"github.com/movsb/taoblog/protocols"
)

// Post ...
type Post struct {
	ID            int64
	Date          int32
	Modified      int32
	Title         string
	Content       string
	Slug          string
	Type          string // TODO use integer
	Category      uint
	Status        string
	PageView      uint
	CommentStatus uint
	Comments      uint
	Metas         string
	Source        string
	SourceType    string
}

// TableName ...
func (Post) TableName() string {
	return `posts`
}

// ToProtocols ...
func (p *Post) ToProtocols() *protocols.Post {
	out := protocols.Post{
		Id:            p.ID,
		Date:          p.Date,
		Modified:      p.Modified,
		Title:         p.Title,
		Content:       p.Content,
		Slug:          p.Slug,
		Category:      int64(p.Category),
		Status:        p.Status,
		PageView:      int64(p.PageView),
		CommentStatus: p.CommentStatus > 0,
		Comments:      int64(p.Comments),
		Metas:         p.Metas,
		Source:        p.Source,
		SourceType:    p.SourceType,
	}

	switch p.Type {
	case `post`:
		out.Type = protocols.PostType_PostType_Post
	case `page`:
		out.Type = protocols.PostType_PostType_Page
	}

	return &out
}

// Posts ...
type Posts []*Post

// ToProtocols ...
func (ps Posts) ToProtocols() (posts []*protocols.Post) {
	for _, post := range ps {
		posts = append(posts, post.ToProtocols())
	}
	return
}

type PostForRelated struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Relevance uint   `json:"relevance"`
}
