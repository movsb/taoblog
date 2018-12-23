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
	Type          string
	Category      uint `taorm:"name:taxonomy"`
	Status        string
	PageView      uint
	CommentStatus uint
	Comments      uint
	Metas         string
	Source        string
	SourceType    string
	Tags          []string

	_Metas map[string]interface{}
}

func (p *Post) decodeMetas() {
	if p._Metas == nil {
		json.Unmarshal([]byte(p.Metas), &p._Metas)
	}
}

func (p *Post) Serialize() *protocols.Post {
	p.decodeMetas()

	return &protocols.Post{
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
		Metas:         p._Metas,
		Source:        p.Source,
		SourceType:    p.SourceType,
		Tags:          p.Tags,
	}
}

type Posts []*Post

func (ps Posts) Serialize() []*protocols.Post {
	sp := []*protocols.Post{}
	for _, p := range ps {
		sp = append(sp, p.Serialize())
	}
	return sp
}

type PostForLatest struct {
	ID    int64
	Title string
	Type  string
}

func (p *PostForLatest) Serialize() *protocols.PostForLatest {
	return &protocols.PostForLatest{
		ID:    p.ID,
		Title: p.Title,
		Type:  p.Type,
	}
}

type PostForLatests []*PostForLatest

func (ps PostForLatests) Serialize() []*protocols.PostForLatest {
	sp := []*protocols.PostForLatest{}
	for _, p := range ps {
		sp = append(sp, p.Serialize())
	}
	return sp
}

type PostForRelated struct {
	ID        int64
	Title     string
	Relevance uint
}

func (p *PostForRelated) Serialize() *protocols.PostForRelated {
	return &protocols.PostForRelated{
		ID:        p.ID,
		Title:     p.Title,
		Relevance: p.Relevance,
	}
}

type PostForRelateds []*PostForRelated

func (ps PostForRelateds) Serialize() []*protocols.PostForRelated {
	sp := []*protocols.PostForRelated{}
	for _, p := range ps {
		sp = append(sp, p.Serialize())
	}
	return sp
}
