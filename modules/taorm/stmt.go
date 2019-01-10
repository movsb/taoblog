package taorm

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// _Where ...
type _Where struct {
	query string
	args  []interface{}
}

func (w *_Where) Rebuild() (query string, args []interface{}) {
	sb := bytes.NewBuffer(nil)
	var i int
	for _, c := range w.query {
		switch c {
		case '?':
			if i >= len(w.args) {
				panic(fmt.Errorf("err where args count"))
			}
			value := reflect.ValueOf(w.args[i])
			if value.Kind() == reflect.Slice {
				n := value.Len()
				sb.WriteString(createSQLInMarks(n))
				for j := 0; j < n; j++ {
					args = append(args, value.Index(j).Interface())
				}
			} else {
				sb.WriteByte('?')
				args = append(args, w.args[i])
			}
			i++
		default:
			sb.WriteRune(c)
		}
	}
	if i != len(w.args) {
		panic(fmt.Errorf("err where args count"))
	}
	return sb.String(), args
}

// Stmt is an SQL statement.
type Stmt struct {
	db         *DB
	model      interface{}
	tableNames []string
	fields     string
	ands       []_Where
	ors        []_Where
	groupBy    string
	orderBy    string
	limit      int64
	offset     int64
}

type DB struct {
	rdb *sql.DB    // raw db
	cdb _SQLCommon // common db
}

func NewDB(db *sql.DB) *DB {
	t := &DB{
		rdb: db,
		cdb: db,
	}
	return t
}

func (db *DB) Common() _SQLCommon {
	return db.cdb
}

// TxCall calls callback within transaction.
// It automatically catches and re-throws exceptions.
func (db *DB) TxCall(callback func(tx *DB) error) error {
	var err error

	rtx, err := db.rdb.Begin()
	if err != nil {
		return err
	}

	tx := &DB{
		rdb: db.rdb,
		cdb: rtx,
	}

	catchCall := func() (except interface{}) {
		defer func() {
			except = recover()
		}()
		err = callback(tx)
		return
	}

	if except := catchCall(); except != nil {
		rtx.Rollback()
		panic(except)
	}

	if err != nil {
		rtx.Rollback()
		return err
	}

	if err := rtx.Commit(); err != nil {
		rtx.Rollback()
		return err
	}

	return nil
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
	s.ands = append(s.ands, w)
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

// noWheres returns true if no SQL conditions.
// Includes and, or.
func (s *Stmt) noWheres() bool {
	return len(s.ands)+len(s.ors) <= 0
}

func (s *Stmt) buildWheres() (string, []interface{}) {
	if s.noWheres() {
		return "", nil
	}

	var args []interface{}
	sb := bytes.NewBuffer(nil)
	fw := func(format string, args ...interface{}) {
		sb.WriteString(fmt.Sprintf(format, args...))
	}
	sb.WriteString(" WHERE ")
	for i, w := range s.ands {
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

// Exec ...
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.cdb.Exec(query, args...)
}

// MustExec ...
func (db *DB) MustExec(query string, args ...interface{}) sql.Result {
	result, err := db.Exec(query, args...)
	if err != nil {
		panic(err)
	}
	return result
}

// Create ...
func (s *Stmt) Create() error {
	fields, args := collectDataFromModel(s.model)
	if len(fields) == 0 {
		return ErrNoFields
	}

	var query string
	query += s.buildCreate()
	query += fmt.Sprintf(` (%s) VALUES (%s)`,
		strings.Join(fields, ","),
		createSQLInMarks(len(fields)),
	)

	dumpSQL(query, args...)

	result, err := s.db.cdb.Exec(query, args...)
	if err != nil {
		return WrapError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	setPrimaryKeyValue(s.model, id)

	return nil
}

// MustCreate ...
func (s *Stmt) MustCreate() {
	if err := s.Create(); err != nil {
		panic(err)
	}
}

// Find ...
func (s *Stmt) Find(out interface{}) error {
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
	return QueryRows(out, s.db.cdb, query, args...)
}

// MustFind ...
func (s *Stmt) MustFind(out interface{}) {
	if err := s.Find(out); err != nil {
		panic(err)
	}
}

func (s *Stmt) updateMap(fields map[string]interface{}, anyway bool) error {
	var query string
	var args []interface{}

	query += s.buildUpdate()

	var updates []string
	var values []interface{}

	if len(fields) == 0 {
		return ErrNoFields
	}

	for field, value := range fields {
		pair := fmt.Sprintf("%s=?", field)
		updates = append(updates, pair)
		values = append(values, value)
	}

	query += strings.Join(updates, ",")
	args = append(args, values...)

	if !anyway && s.noWheres() {
		return ErrNoWhere
	}

	whereQuery, whereArgs := s.buildWheres()
	query += whereQuery
	args = append(args, whereArgs...)

	query += s.buildLimit()

	dumpSQL(query, args...)

	_, err := s.db.cdb.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}

// UpdateMap ...
func (s *Stmt) UpdateMap(updates map[string]interface{}) error {
	return s.updateMap(updates, false)
}

// UpdateMapAnyway ...
func (s *Stmt) UpdateMapAnyway(updates map[string]interface{}) error {
	return s.updateMap(updates, true)
}

// MustUpdateMap ...
func (s *Stmt) MustUpdateMap(updates map[string]interface{}) {
	if err := s.updateMap(updates, false); err != nil {
		panic(err)
	}
}

// MustUpdateMapAnyway ...
func (s *Stmt) MustUpdateMapAnyway(updates map[string]interface{}) {
	if err := s.updateMap(updates, true); err != nil {
		panic(err)
	}
}

func (s *Stmt) _delete(anyway bool) error {
	var query string
	var args []interface{}

	query += s.buildDelete()

	if !anyway && s.noWheres() {
		return ErrNoWhere
	}

	whereQuery, whereArgs := s.buildWheres()
	query += whereQuery
	args = append(args, whereArgs...)

	query += s.buildLimit()

	dumpSQL(query, args...)

	_, err := s.db.cdb.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}

// Delete ...
func (s *Stmt) Delete() error {
	return s._delete(false)
}

// DeleteAnyway ...
func (s *Stmt) DeleteAnyway() error {
	return s._delete(true)
}

// MustDelete ...
func (s *Stmt) MustDelete() {
	if err := s.Delete(); err != nil {
		panic(err)
	}
}

// MustDeleteAnyway ...
func (s *Stmt) MustDeleteAnyway() {
	if err := s.DeleteAnyway(); err != nil {
		panic(err)
	}
}
