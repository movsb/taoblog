package main

import (
	"fmt"
)

// CategoryNotFoundError is
type CategoryNotFoundError struct {
	Name string
	ID   int64
}

func (z *CategoryNotFoundError) Error() string {
	return fmt.Sprintf("CategoryNotFoundError: ID=%d, Name=%s", z.ID, z.Name)
}

// CategoryManager manages categories.
type CategoryManager struct {
}

// NewCategoryManager news a CategoryManager.
func NewCategoryManager() *CategoryManager {
	return &CategoryManager{}
}

/*
// GetTree creates a category tree.
func (z *CategoryManager) GetTree(tx Querier) ([]*Category, error) {
	cats, err := z.ListCategories(tx)
	if err != nil {
		return nil, err
	}

	var makeChildren func(parent *Category)

	makeChildren = func(parent *Category) {
		for i, c := range cats {
			if c != nil && c.Parent == parent.ID {
				parent.Children = append(parent.Children, c)
				cats[i] = nil
				makeChildren(c)
			}
		}
	}

	var dummy Category
	makeChildren(&dummy)
	return dummy.Children, nil
}
*/

func (z *CategoryManager) GetCountOfCategoriesAll(tx Querier) (map[int64]int64, error) {
	query := `select taxonomy,count(id) count from posts group by taxonomy`
	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cats := make(map[int64]int64)
	for rows.Next() {
		var id, count int64
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}
		cats[id] = count
	}
	return cats, rows.Err()
}
