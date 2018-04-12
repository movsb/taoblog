package main

import (
	"database/sql"
	"fmt"
)

type xOptionsModel struct {
	db *sql.DB
}

type OptionItem struct {
	Name  string
	Value string
}

func newOptionsModel(db *sql.DB) *xOptionsModel {
	return &xOptionsModel{
		db: db,
	}
}

func (o *xOptionsModel) Has(name string) error {
	query := `SELECT name FROM options WHERE name=? LIMIT 1`
	val := ""
	row := o.db.QueryRow(query, name)
	return row.Scan(&val)
}

func (o *xOptionsModel) Get(name string) (string, error) {
	query := `SELECT value FROM options WHERE name=? LIMIT 1`
	row := o.db.QueryRow(query, name)
	val := ""
	err := row.Scan(&val)
	return val, err
}

func (o *xOptionsModel) GetDef(name string, def string) string {
	val, err := o.Get(name)
	if err == nil {
		return val
	}
	return def
}

func (o *xOptionsModel) Set(name string, val interface{}) error {
	strVal := fmt.Sprint(val)

	query := ""

	if o.Has(name) == nil {
		query = `UPDATE options SET value=? WHERE name=? LIMIT 1`
	} else {
		query = `INSERT INTO options (name,value) VALUES (?,?)`
	}
	_, err := o.db.Exec(query, name, strVal)
	return err
}

func (o *xOptionsModel) Del(name string) error {
	query := `DELETE FROM options WHERE name=? LIMIT 1`
	_, err := o.db.Exec(query, name)
	return err
}

func (o *xOptionsModel) List() ([]OptionItem, error) {
	items := make([]OptionItem, 0)
	query := `SELECT name,value FROM options`
	rows, err := o.db.Query(query)
	if err != nil {
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var item OptionItem
		if err := rows.Scan(&item.Name, &item.Value); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
