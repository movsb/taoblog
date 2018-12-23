package protocols

import "fmt"

// Post ...
type Post struct {
	ID            int64                  `json:"id"`
	Date          string                 `json:"date"`
	Modified      string                 `json:"modified"`
	Title         string                 `json:"title"`
	Content       string                 `json:"content"`
	Slug          string                 `json:"slug"`
	Type          string                 `json:"type"`
	Category      uint                   `json:"category" taorm:"name:taxonomy"`
	Status        string                 `json:"status"`
	PageView      uint                   `json:"page_view"`
	CommentStatus uint                   `json:"comment_status"`
	Comments      uint                   `json:"comments"`
	Metas         map[string]interface{} `json:"metas"`
	Source        string                 `json:"source"`
	SourceType    string                 `json:"source_type"`
	Tags          []string               `json:"tags"`
}

func (p *Post) CustomHeader() (header string) {
	if i, ok := p.Metas["header"]; ok {
		if s, ok := i.(string); ok {
			header = s
		}
	}
	return
}

func (p *Post) CustomFooter() (footer string) {
	if i, ok := p.Metas["footer"]; ok {
		if s, ok := i.(string); ok {
			footer = s
		}
	}
	return
}

type PostForLatest struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

func (p *PostForLatest) Link() string {
	return fmt.Sprintf("/%d/", p.ID)
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

type PostForArchive struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type GetPostRequest struct {
	Name int64
}

type ListPostsRequest struct {
}

type ListPostsResponse struct {
	Posts []*Post `json:"posts"`
}

type GetLatestPostsRequest struct {
	Limit int64
}

type GetLatestPostsResponse struct {
	Posts []*PostForLatest `json:"posts"`
}

type GetRelatedPostsRequest struct {
	PostID int64
}

type GetRelatedPostsResponse struct {
	Posts []*PostForRelated `json:"posts"`
}

type GetArchivePostsRequest struct {
}

type GetArchivePostsResponse struct {
	Posts []*PostForArchive `json:"posts"`
}

type IncrementPostViewRequest struct {
	PostID int64
}

type IncrementPostViewResponse struct {
	View int64
}

type GetPostByIDRequest struct {
	ID int64
}

type GetPostBySlugRequest struct {
	Category string
	Slug     string
}

type GetPostByPageRequest struct {
	Parents string
	Slug    string
}
