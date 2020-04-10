package models

// Tag is a tag.
type Tag struct {
	ID    int64
	Name  string
	Alias int64
}

// TableName ...
func (Tag) TableName() string {
	return `tags`
}

// ObjectTag ...
type ObjectTag struct {
	ID     int64
	PostID int64
	TagID  int64
}

// TableName ...
func (ObjectTag) TableName() string {
	return `post_tags`
}

// TagWithCount is a tag with associated post count.
type TagWithCount struct {
	Tag
	Count int64
}
