package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"

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
	Type          string
	Category      uint
	Status        string
	PageView      uint
	CommentStatus uint
	Comments      uint
	Metas         PostMeta
	Source        string
	SourceType    string
}

type PostMeta map[string]string

var (
	_ sql.Scanner   = (*PostMeta)(nil)
	_ driver.Valuer = (*PostMeta)(nil)
)

func (m PostMeta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *PostMeta) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), m)
	case []byte:
		return json.Unmarshal(v, m)
	}
	return errors.New(`unsupported type`)
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
		Type:          p.Type,
		Category:      int64(p.Category),
		Status:        p.Status,
		PageView:      int64(p.PageView),
		CommentStatus: p.CommentStatus > 0,
		Comments:      int64(p.Comments),
		Metas:         p.Metas,
		Source:        p.Source,
		SourceType:    p.SourceType,
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
