package protocols

// PostMetas ...
type PostMetas map[string]interface{}

type PostType string

const (
	PostTypePost PostType = "post"
	PostTypePage PostType = "page"
)

// Post ...
type Post struct {
	ID            int64     `json:"id,omitempty"`
	Date          string    `json:"date,omitempty"`
	Modified      string    `json:"modified,omitempty"`
	Title         string    `json:"title,omitempty"`
	Content       string    `json:"content,omitempty"`
	Slug          string    `json:"slug,omitempty"`
	Type          PostType  `json:"type,omitempty"`
	Category      uint      `json:"category,omitempty"`
	Status        string    `json:"status,omitempty"`
	PageView      uint      `json:"page_view,omitempty"`
	CommentStatus uint      `json:"comment_status,omitempty"`
	Comments      uint      `json:"comments,omitempty"`
	Metas         PostMetas `json:"metas,omitempty"`
	Source        string    `json:"source,omitempty"`
	SourceType    string    `json:"source_type,omitempty"`
	Tags          []string  `json:"tags,omitempty"`
}
