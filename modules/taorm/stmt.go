package taorm

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type Stmt struct {
	db      *DB
	model   interface{}
	name    string
	fields  string
	wheres  []_Where
	ors     []_Where
	groupBy string
	orderBy string
	limit   int64
	offset  int64
}

type DB struct {
	db *sql.DB
}

func (db *DB) Model(model interface{}, name string) *Stmt {
	stmt := &Stmt{
		db:     db,
		model:  model,
		name:   name,
		limit:  -1,
		offset: -1,
	}

	stmt.initPrimaryKey()

	return stmt
}

func (s *Stmt) Select(fields string) *Stmt {
	s.fields = fields
	return s
}

func (s *Stmt) Where(query string, args ...interface{}) *Stmt {
	w := _Where{
		query: query,
		args:  args,
	}
	s.wheres = append(s.wheres, w)
	return s
}

func (s *Stmt) WhereIf(cond bool, query string, args ...interface{}) *Stmt {
	if cond {
		s.Where(query, args...)
	}
	return s
}

func (s *Stmt) Or(query string, args ...interface{}) *Stmt {
	w := _Where{
		query: query,
		args:  args,
	}
	s.ors = append(s.ors, w)
	return s
}

func (s *Stmt) GroupBy(groupBy string) *Stmt {
	s.groupBy = groupBy
	return s
}

func (s *Stmt) OrderBy(orderBy string) *Stmt {
	s.orderBy = orderBy
	return s
}

func (s *Stmt) Limit(limit int64) *Stmt {
	s.limit = limit
	return s
}

func (s *Stmt) Offset(offset int64) *Stmt {
	s.offset = offset
	return s
}

func (s *Stmt) initPrimaryKey() {
	ty := reflect.TypeOf(s.model)
	value := reflect.ValueOf(s.model)
	if ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
		value = value.Elem()
	}
	if ty.Kind() != reflect.Struct {
		panic("not struct")
	}
	for i := 0; i < ty.NumField(); i++ {
		f := value.Field(i)
		columnName := getColumnName(ty.Field(i))
		if columnName == "id" {
			id := f.Interface().(int64)
			if id > 0 {
				s.Where("id = ?", id)
			}
			break
		}
	}
}

func (s *Stmt) buildWheres() (string, []interface{}) {
	var args []interface{}
	sb := bytes.NewBuffer(nil)
	fw := func(format string, args ...interface{}) {
		sb.WriteString(fmt.Sprintf(format, args...))
	}
	if len(s.wheres)+len(s.ors) > 0 {
		sb.WriteString(" WHERE ")
		for i, w := range s.wheres {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			query, xargs := w.Rebuild()
			fw("(%s)", query)
			args = append(args, xargs...)
		}
		for i, w := range s.ors {
			if i > 0 {
				sb.WriteString(" OR ")
			}
			query, xargs := w.Rebuild()
			fw("(%s)", query)
			args = append(args, xargs...)
		}
	}
	return sb.String(), args
}

func (s *Stmt) buildSelect() string {
	if s.fields == "" {
		s.fields = "*"
	}
	panicIf(s.name == "", "model is empty")
	return fmt.Sprintf(`SELECT %s FROM %s`, s.fields, s.name)
}

func (s *Stmt) buildGroupBy() (groupBy string) {
	if s.groupBy != "" {
		groupBy = fmt.Sprintf(` GROUP BY %s`, s.groupBy)
	}
	return
}

func (s *Stmt) buildOrderBy() (orderBy string) {
	if s.orderBy != "" {
		orderBy = fmt.Sprintf(` ORDER BY %s`, s.orderBy)
	}
	return
}

func (s *Stmt) buildLimit() (limit string) {
	if s.limit >= 0 {
		limit += fmt.Sprintf(" LIMIT %d", s.limit)
		if s.offset >= 0 {
			limit += fmt.Sprintf(" OFFSET %d", s.offset)
		}
	}
	return
}

func (s *Stmt) Find(out interface{}) {
	var query string
	var args = []interface{}{}

	query += s.buildSelect()

	whereQuery, whereArgs := s.buildWheres()

	query += whereQuery
	args = append(args, whereArgs...)

	query += s.buildGroupBy()
	query += s.buildOrderBy()
	query += s.buildLimit()

	fmt.Printf(strings.Replace(query, "?", "%v", -1)+"\n", args...)
	MustQueryRows(out, s.db.db, query, args...)
}
