package main

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/movsb/taoblog/server/modules/sql_helpers"
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

// GetCategoryByID gets a category by its ID.
func (z *CategoryManager) GetCategoryByID(tx Querier, id int64) (*Category, error) {
	query, args := sql_helpers.NewSelect().From("taxonomies", "").
		Select("*").Where("id=?", id).Limit(1).SQL()
	row := tx.QueryRow(query, args...)
	return z.scanOne(row)
}

func (z *CategoryManager) scanOne(scn RowScanner) (*Category, error) {
	var cat Category
	if err := scn.Scan(&cat.ID, &cat.Name, &cat.Slug, &cat.Parent, &cat.Ancestor); err != nil {
		if err == sql.ErrNoRows {
			return nil, &CategoryNotFoundError{}
		}
		return nil, err
	}
	return &cat, nil
}

func (z *CategoryManager) scanMulti(rows *sql.Rows) ([]*Category, error) {
	defer rows.Close()
	cats := make([]*Category, 0)
	for rows.Next() {
		cat, err := z.scanOne(rows)
		if err != nil {
			return nil, err
		}
		cats = append(cats, cat)
	}
	return cats, nil
}

// ListCategories lists categories.
func (z *CategoryManager) ListCategories(tx Querier) ([]*Category, error) {
	query, args := sql_helpers.NewSelect().From("taxonomies", "").Select("*").SQL()
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return z.scanMulti(rows)
}

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

// GetChildren gets direct descendant children.
func (z *CategoryManager) GetChildren(tx Querier, parent int64) ([]*Category, error) {
	query, args := sql_helpers.NewSelect().From("taxonomies", "").
		Select("*").Where("parent=?", parent).SQL()
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return z.scanMulti(rows)
}

func (z *CategoryManager) UpdateCategory(tx Querier, cat *Category) error {
	return nil
}

func (z *CategoryManager) CreateCategory(tx Querier, cat *Category) error {
	return nil
}

// ParseTree parses category tree from URL to get last sub-category ID.
// e.g. /path/to/folder/post.html, then tree is path/to/folder
// It will get the ID of folder
func (z *CategoryManager) ParseTree(tx Querier, tree string) (id int64, err error) {
	parts := strings.Split(tree, "/")
	query, args := sql_helpers.NewSelect().From("taxonomies", "").
		Select("*").Where("slug IN (?)", parts).SQL()
	rows, err := tx.Query(query, args...)
	if err != nil {
		return 0, err
	}
	cats, err := z.scanMulti(rows)
	if err != nil {
		return 0, err
	}
	var parent int64
	for i := 0; i < len(parts); i++ {
		found := false
		for _, cat := range cats {
			if cat.Parent == parent && cat.Slug == parts[i] {
				parent = cat.ID
				found = true
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("找不到分类：%s", parts[i])
		}
	}
	return parent, nil
}

func (z *CategoryManager) GetCountOfCategoriesAll(tx Querier) (map[int64]int64, error) {
	query, args := sql_helpers.NewSelect().
		From("posts", "").
		Select("taxonomy,count(id) count").
		GroupBy("taxonomy").SQL()
	rows, err := tx.Query(query, args...)
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
