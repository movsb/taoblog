package main

import (
	"database/sql"
	"fmt"
	"log"
)

type OptionManager struct {
	db *sql.DB
}

type Option struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func newOptionsModel(db *sql.DB) *OptionManager {
	return &OptionManager{
		db: db,
	}
}

func (o *OptionManager) Has(name string) error {
	query := `SELECT name FROM options WHERE name=? LIMIT 1`
	val := ""
	row := o.db.QueryRow(query, name)
	return row.Scan(&val)
}

func (o *OptionManager) Get(name string) (string, error) {
	query := `SELECT value FROM options WHERE name=? LIMIT 1`
	row := o.db.QueryRow(query, name)
	val := ""
	err := row.Scan(&val)
	return val, err
}

func (o *OptionManager) GetDef(name string, def string) string {
	val, err := o.Get(name)
	if err == nil {
		return val
	}
	return def
}

func (o *OptionManager) Set(name string, val interface{}) error {
	strVal := fmt.Sprint(val)

	query := ""
	var err error

	if o.Has(name) == nil {
		query = `UPDATE options SET value=? WHERE name=? LIMIT 1`
		_, err = o.db.Exec(query, strVal, name)
	} else {
		query = `INSERT INTO options (name,value) VALUES (?,?)`
		_, err = o.db.Exec(query, name, strVal)
	}
	log.Println(query)
	return err
}

func (o *OptionManager) Del(name string) error {
	query := `DELETE FROM options WHERE name=? LIMIT 1`
	_, err := o.db.Exec(query, name)
	return err
}

func (o *OptionManager) List() ([]Option, error) {
	items := make([]Option, 0)
	query := `SELECT name,value FROM options`
	rows, err := o.db.Query(query)
	if err != nil {
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var item Option
		if err := rows.Scan(&item.Name, &item.Value); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
