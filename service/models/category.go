package models

// Category is a post category
type Category struct {
	ID       int64       `json:"id"`
	Name     string      `json:"name"`
	Slug     string      `json:"slug"`
	Parent   int64       `json:"parent"`
	Root     int64       `json:"root"`
	Children []*Category `json:"children"`
}
