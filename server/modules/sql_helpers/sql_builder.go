package sql_helpers

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

type _Where struct {
	query string
	args  []interface{}
}

func (w *_Where) Rebuild() (query string, args []interface{}, err error) {
	sb := bytes.NewBuffer(nil)
	var i int
	for _, c := range w.query {
		switch c {
		case '?':
			if i >= len(w.args) {
				err = fmt.Errorf("err where args count")
				return
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
		err = fmt.Errorf("err where args count")
		return
	}
	return sb.String(), args, nil
}

type _Table struct {
	name  string
	alias string
}

type Select struct {
	fields  string
	tables  []*_Table
	wheres  []*_Where
	ors     []*_Where
	groupBy string
	orderBy string
	limit   int64
	offset  int64
}

func (s *Select) Select(fields string) *Select {
	s.fields = fields
	return s
}

func (s *Select) From(table string, alias string) *Select {
	s.tables = append(s.tables, &_Table{
		name:  table,
		alias: alias,
	})
	return s
}

func (s *Select) Where(query string, args ...interface{}) *Select {
	w := &_Where{
		query: query,
		args:  args,
	}
	s.wheres = append(s.wheres, w)
	return s
}

func (s *Select) WhereIf(condition bool, query string, args ...interface{}) *Select {
	if condition {
		s.Where(query, args...)
	}
	return s
}

func (s *Select) Or(query string, args ...interface{}) *Select {
	w := &_Where{
		query: query,
		args:  args,
	}
	s.ors = append(s.ors, w)
	return s
}

func (s *Select) GroupBy(groupBy string) *Select {
	s.groupBy = groupBy
	return s
}

func (s *Select) OrderBy(orderBy string) *Select {
	s.orderBy = orderBy
	return s
}

func (s *Select) Limit(limit int64) *Select {
	s.limit = limit
	return s
}

func (s *Select) Offset(offset int64) *Select {
	s.offset = offset
	return s
}

func (s *Select) SQL() (string, []interface{}) {
	sb := bytes.NewBuffer(nil)
	fw := func(format string, args ...interface{}) {
		sb.WriteString(fmt.Sprintf(format, args...))
	}
	var args = []interface{}{}

	sb.WriteString("SELECT ")
	sb.WriteString(s.fields)
	sb.WriteString(" FROM ")
	for i, t := range s.tables {
		if i > 0 {
			sb.WriteRune(',')
		}
		sb.WriteString(t.name)
		if t.alias != "" {
			fw(" %s", t.alias)
		}
	}
	if len(s.wheres)+len(s.ors) > 0 {
		sb.WriteString(" WHERE ")
		for i, w := range s.wheres {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			query, xargs, err := w.Rebuild()
			if err != nil {
				panic(err)
			}
			fw("(%s)", query)
			args = append(args, xargs...)
		}
		for i, w := range s.ors {
			if i > 0 {
				sb.WriteString(" OR ")
			}
			query, xargs, err := w.Rebuild()
			if err != nil {
				panic(err)
			}
			fw("(%s)", query)
			args = append(args, xargs...)
		}
	}
	if s.groupBy != "" {
		fw(" GROUP BY %s", s.groupBy)
	}
	if s.orderBy != "" {
		fw(" ORDER BY %s", s.orderBy)
	}
	if s.limit > 0 {
		fw(" LIMIT %d", s.limit)
		if s.offset >= 0 {
			fw(" OFFSET %d", s.offset)
		}
	}
	str := sb.String()
	fmt.Printf(strings.Replace(str, "?", "%v", -1)+"\n", args...)
	return str, args
}

func NewSelect() *Select {
	return &Select{
		limit:  -1,
		offset: -1,
	}
}
