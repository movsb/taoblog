package models

// Tag is a tag.
type Tag struct {
	ID    int64
	Name  string
	Alias int64
}

// ObjectTag ...
type ObjectTag struct {
	ID     int64
	PostID int64
	TagID  int64
}

// TagWithCount is a tag with associated post count.
type TagWithCount struct {
	Tag
	Count int64
}
