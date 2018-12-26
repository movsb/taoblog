package models

import (
	"encoding/json"
)

type Post struct {
	ID            int64    `json:"id"`
	Date          string   `json:"date"`
	Modified      string   `json:"modified"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	Slug          string   `json:"slug"`
	Type          string   `json:"type"`
	Category      uint     `json:"category" taorm:"name:taxonomy"`
	Status        string   `json:"status"`
	PageView      uint     `json:"page_view"`
	CommentStatus uint     `json:"comment_status"`
	Comments      uint     `json:"comments"`
	Metas         string   `json:"metas"`
	Source        string   `json:"source"`
	SourceType    string   `json:"source_type"`
	Tags          []string `json:"tags"`

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
