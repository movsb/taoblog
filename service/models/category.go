package models

// Category is a post category
type Category struct {
	ID       int64
	Name     string
	Slug     string
	Parent   int64
	Ancestor int64
	Children []*Category
}
