package taorm

import (
	"database/sql"
	"reflect"
)

// QueryRows queries results.
// out can be either *Struct, or *[]*Struct.
//
// For querying single row, QueryRows returns:
//   nil          : no error (got row)
//   sql.ErrNoRows: an error (no data)
// For querying multiple rows, QueryRows returns:
//   nil          : no error (but can be empty slice)
//   some error   : an error
func QueryRows(out interface{}, tx Querier, query string, args ...interface{}) error {
	rows, err := tx.Query(query, args...)
	if err != nil {
		return err
	}

	defer rows.Close()

	ty := reflect.TypeOf(out)
	if ty.Kind() != reflect.Ptr {
		panic("out must be pointer")
	}

	ty = ty.Elem()
	switch ty.Kind() {
	case reflect.Struct:
		return queryRow(out, rows)
	case reflect.Slice:
		return queryRows(out, rows)
	default:
		panic("unknown pointer type")
	}
}

func queryRow(out interface{}, rows *sql.Rows) (err error) {
	if rows.Next() {
		fields := getPointers(out, rows)
		err = rows.Scan(fields...)
	} else {
		err = rows.Err()
		// Next() is used to get next record.
		// err:
		//    nil: no more record
		//   !nil: error happened
		if err == nil {
			err = sql.ErrNoRows
		}
	}
	return
}

func queryRows(out interface{}, rows *sql.Rows) error {
	ty := reflect.TypeOf(out).Elem()
	slice := reflect.MakeSlice(ty, 0, 0)

	ty = ty.Elem()
	if ty.Kind() != reflect.Ptr {
		panic("must be slice of pointer to struct")
	}
	ty = ty.Elem()
	if ty.Kind() != reflect.Struct {
		panic("must be slice of pointer to struct")
	}

	for rows.Next() {
		val := reflect.New(ty)
		ptr := val.Interface()
		fields := getPointers(ptr, rows)
		if err := rows.Scan(fields...); err != nil {
			return err
		}
		slice = reflect.Append(slice, val)
	}

	reflect.ValueOf(out).Elem().Set(slice)

	return rows.Err()
}
