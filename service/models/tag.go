package models

// Tag is a tag.
type Tag struct {
	ID    int64
	Name  string
	Alias int64
}

// TagWithCount is a tag with associated post count.
type TagWithCount struct {
	Tag
	Count int64
}
