package protocols

// Comment ...
type Comment struct {
	// From Comment
	ID      int64  `json:"id"`
	Parent  int64  `json:"parent"`
	Root    int64  `json:"root"`
	PostID  int64  `json:"post_id"`
	Author  string `json:"author"`
	Email   string `json:"email,omitempty"`
	URL     string `json:"url"`
	IP      string `json:"ip,omitempty"`
	Date    string `json:"date"`
	Content string `json:"content"`

	// Owned
	Children []*Comment `json:"children"`
	Avatar   string     `json:"avatar"`
	IsAdmin  bool       `json:"is_admin"`
}

type ListCommentsMode int

const (
	ListCommentsModeTree = iota
	ListCommentsModeFlat
)

type ListCommentsRequest struct {
	Mode    ListCommentsMode
	PostID  int64
	Fields  string
	Limit   int64
	Offset  int64
	OrderBy string
}
