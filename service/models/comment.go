package models

type Comment struct {
	ID       int64  `json:"id"`
	Parent   int64  `json:"parent"`
	Ancestor int64  `json:"ancestor"`
	PostID   int64  `json:"post_id"`
	Author   string `json:"author"`
	Email    string `json:"email"`
	URL      string `json:"url"`
	IP       string `json:"ip"`
	Date     string `json:"date"`
	Content  string `json:"content"`
}
