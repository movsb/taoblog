package models

import (
	"encoding/json"

	"github.com/movsb/taoblog/protocols"
)

type Post struct {
	ID            int64
	Date          string
	Modified      string
	Title         string
	Content       string
	Slug          string
	Type          protocols.PostType
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

func (p *Post) ToProtocols() *protocols.Post {
	out := protocols.Post{
		ID:            p.ID,
		Date:          p.Date,
		Modified:      p.Modified,
		Title:         p.Title,
		Content:       p.Content,
		Slug:          p.Slug,
		Type:          p.Type,
		Category:      p.Category,
		Status:        p.Status,
		PageView:      p.PageView,
		CommentStatus: p.CommentStatus,
		Comments:      p.Comments,
		Source:        p.Source,
		SourceType:    p.SourceType,
	}

	json.Unmarshal([]byte(p.Metas), &out.Metas)
	return &out
}

type Posts []*Post

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

type PostForDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Count int `json:"count"`
}

// PostForManagement for post management.
type PostForManagement struct {
	ID           int64  `json:"id"`
	Date         string `json:"date"`
	Modified     string `json:"modified"`
	Title        string `json:"title"`
	PageView     uint   `json:"page_view"`
	SourceType   string `json:"source_type"`
	CommentCount uint   `json:"comment_count" taorm:"name:comments"`
}
