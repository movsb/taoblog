package models

// Category is a post category
type Category struct {
	ID       int64       `json:"id"`
	Name     string      `json:"name"`
	Slug     string      `json:"slug"`
	ParentID int64       `json:"parent"`
	Path     string      `json:"path"`
	Children []*Category `json:"children"`
}
