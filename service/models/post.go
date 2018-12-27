package models

import (
	"encoding/json"
)

type Post struct {
	ID            int64    `json:"id,omitempty"`
	Date          string   `json:"date,omitempty"`
	Modified      string   `json:"modified,omitempty"`
	Title         string   `json:"title,omitempty"`
	Content       string   `json:"content,omitempty"`
	Slug          string   `json:"slug,omitempty"`
	Type          string   `json:"type,omitempty"`
	Category      uint     `json:"category,omitempty" taorm:"name:taxonomy"`
	Status        string   `json:"status,omitempty"`
	PageView      uint     `json:"page_view,omitempty"`
	CommentStatus uint     `json:"comment_status,omitempty"`
	Comments      uint     `json:"comments,omitempty"`
	Metas         string   `json:"metas,omitempty"`
	Source        string   `json:"source,omitempty"`
	SourceType    string   `json:"source_type,omitempty"`
	Tags          []string `json:"tags,omitempty"`

	_Metas map[string]interface{}
}

func (p *Post) decodeMetas() {
	if p._Metas == nil {
		json.Unmarshal([]byte(p.Metas), &p._Metas)
	}
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

type PostForArchive struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}
