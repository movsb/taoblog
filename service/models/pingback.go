package models

// Pingback ...
type Pingback struct {
	ID        int64
	CreatedAt int64
	PostID    int64
	Title     string
	SourceURL string
}

// TableName ...
func (Pingback) TableName() string {
	return "pingbacks"
}
