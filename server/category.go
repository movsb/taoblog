package main

// Category is a post category
type Category struct {
	ID       int64       `json:"id"`
	Name     string      `json:"name"`
	Slug     string      `json:"slug"`
	Parent   int64       `json:"parent"`
	Ancestor int64       `json:"ancestor"`
	Children []*Category `json:"children"`
}

type CategoryForGetAll struct {
	ID    int64 `json:"id"`
	Count int64 `json:"count"`
}

// Create creates category into database.
func (z *Category) Create(tx Querier) error {
	return nil
}
