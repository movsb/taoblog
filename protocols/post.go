package protocols

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

type GetPostRequest struct {
	Name int64
}

type ListPostsRequest struct {
}

type ListPostsResponse struct {
	Posts []*Post `json:"posts"`
}
