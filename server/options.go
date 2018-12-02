package main

import (
	"fmt"
	"log"
	"strconv"
)

type OptionManager struct {
}

type Option struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func newOptionsModel() *OptionManager {
	return &OptionManager{}
}

func (o *OptionManager) Has(tx Querier, name string) error {
	query := `SELECT name FROM options WHERE name=? LIMIT 1`
	val := ""
	row := tx.QueryRow(query, name)
	return row.Scan(&val)
}

func (o *OptionManager) Get(tx Querier, name string) (string, error) {
	query := `SELECT value FROM options WHERE name=? LIMIT 1`
	row := tx.QueryRow(query, name)
	val := ""
	err := row.Scan(&val)
	return val, err
}

func (o *OptionManager) GetDef(tx Querier, name string, def string) string {
	val, err := o.Get(tx, name)
	if err == nil {
		return val
	}
	return def
}

func (o *OptionManager) GetDefInt(tx Querier, name string, def int64) int64 {
	val, err := o.Get(tx, name)
	if err == nil {
		num, _ := strconv.ParseInt(val, 10, 64)
		return num
	}
	return def
}

func (o *OptionManager) Set(tx Querier, name string, val interface{}) error {
	strVal := fmt.Sprint(val)

	query := ""
	var err error

	if o.Has(tx, name) == nil {
		query = `UPDATE options SET value=? WHERE name=? LIMIT 1`
		_, err = tx.Exec(query, strVal, name)
	} else {
		query = `INSERT INTO options (name,value) VALUES (?,?)`
		_, err = tx.Exec(query, name, strVal)
	}
	log.Println(query)
	return err
}

func (o *OptionManager) Del(tx Querier, name string) error {
	query := `DELETE FROM options WHERE name=? LIMIT 1`
	_, err := tx.Exec(query, name)
	return err
}

func (o *OptionManager) List(tx Querier) ([]Option, error) {
	items := make([]Option, 0)
	query := `SELECT name,value FROM options`
	rows, err := tx.Query(query)
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
