package taorm

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type Stmt struct {
	db         *DB
	model      interface{}
	tableNames []string
	fields     string
	wheres     []_Where
	ors        []_Where
	groupBy    string
	orderBy    string
	limit      int64
	offset     int64
}

type DB struct {
	db *sql.DB
}

func NewDB(db *sql.DB) *DB {
	t := &DB{
		db: db,
	}
	return t
}

func (db *DB) Model(model interface{}, name string) *Stmt {
	stmt := &Stmt{
		db:         db,
		model:      model,
		tableNames: []string{name},
		limit:      -1,
		offset:     -1,
	}

	stmt.initPrimaryKey()

	return stmt
}

func (db *DB) From(table string) *Stmt {
	stmt := &Stmt{
		db:         db,
		tableNames: []string{table},
		limit:      -1,
		offset:     -1,
	}
	return stmt
}

func (s *Stmt) From(table string) *Stmt {
	s.tableNames = append(s.tableNames, table)
	return s
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

func (s *Stmt) buildCreate() string {
	panicIf(len(s.tableNames) != 1, "model length is not 1")
	return fmt.Sprintf(`INSERT INTO %s `, s.tableNames[0])
}

func (s *Stmt) buildSelect() string {
	if s.fields == "" {
		s.fields = "*"
	}
	panicIf(len(s.tableNames) == 0, "model is empty")
	return fmt.Sprintf(`SELECT %s FROM %s`, s.fields, strings.Join(s.tableNames, ","))
}

func (s *Stmt) buildUpdate() string {
	panicIf(len(s.tableNames) == 0, "model is empty")
	return fmt.Sprintf(`UPDATE %s SET `, strings.Join(s.tableNames, ","))
}

func (s *Stmt) buildDelete() string {
	panicIf(len(s.tableNames) == 0, "model is empty")
	return fmt.Sprintf(`DELETE FROM %s`, strings.Join(s.tableNames, ","))
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
	if s.limit > 0 {
		limit += fmt.Sprintf(" LIMIT %d", s.limit)
		if s.offset >= 0 {
			limit += fmt.Sprintf(" OFFSET %d", s.offset)
		}
	}
	return
}

func (s *Stmt) Create() {
	fields, values := collectDataFromModel(s.model)
	if len(fields) == 0 {
		panic("no fields to insert")
	}

	var query string
	query += s.buildCreate()
	query += fmt.Sprintf(` (%s) VALUES (%s)`,
		strings.Join(fields, ","),
		createSQLInMarks(len(fields)),
	)

	result, err := s.db.db.Exec(query, values...)
	_ = result
	if err != nil {
		panic(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}

	setPrimaryKeyValue(s.model, id)
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

	dumpSQL(query, args...)
	MustQueryRows(out, s.db.db, query, args...)
}

func (s *Stmt) UpdateField(field string, value interface{}) {
	s.UpdateMap(map[string]interface{}{
		field: value,
	})
}

func (s *Stmt) UpdateMap(fields map[string]interface{}) {
	var query string
	var args []interface{}

	query += s.buildUpdate()

	var updates []string
	var values []interface{}

	if len(fields) == 0 {
		panic("no fields to update")
	}

	for field, value := range fields {
		pair := fmt.Sprintf("%s=?", field)
		updates = append(updates, pair)
		values = append(values, value)
	}

	query += strings.Join(updates, ",")
	args = append(args, values...)

	whereQuery, whereArgs := s.buildWheres()

	query += whereQuery
	args = append(args, whereArgs...)

	query += s.buildLimit()

	dumpSQL(query, args...)

	result, err := s.db.db.Exec(query, args...)
	_ = result
	if err != nil {
		panic(err)
	}
}

func (s *Stmt) Delete() {
	var query string
	var args []interface{}

	query += s.buildDelete()

	whereQuery, whereArgs := s.buildWheres()
	query += whereQuery
	args = append(args, whereArgs...)

	query += s.buildLimit()

	dumpSQL(query, args...)

	result, err := s.db.db.Exec(query, args...)
	_ = result
	if err != nil {
		panic(err)
	}
}
