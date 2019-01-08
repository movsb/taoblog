package taorm

import (
	"database/sql"
	"reflect"
)

// QueryRows queries results.
// out can be either *Struct, *[]Struct, or *[]*Struct.
//
// For querying single row, QueryRows returns:
//   nil          : no error (got row)
//   sql.ErrNoRows: an error (no data)
// For querying multiple rows, QueryRows returns:
//   nil          : no error (but can be empty slice)
//   some error   : an error
func QueryRows(out interface{}, tx _SQLCommon, query string, args ...interface{}) error {
	var err error
	rows, err := tx.Query(query, args...)
	if err != nil {
		return err
	}

	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	ty := reflect.TypeOf(out)
	if ty.Kind() != reflect.Ptr {
		return ErrInvalidOut
	}

	ty = ty.Elem()
	switch ty.Kind() {
	case reflect.Struct:
		if rows.Next() {
			pointers := getPointers(out, columns)
			return rows.Scan(pointers...)
		}
		err = rows.Err()
		if err == nil {
			err = sql.ErrNoRows
		}
		return err
	case reflect.Slice:
		slice := reflect.MakeSlice(ty, 0, 0)
		ty = ty.Elem()
		isPtr := ty.Kind() == reflect.Ptr
		if isPtr {
			ty = ty.Elem()
		}
		if ty.Kind() != reflect.Struct {
			return ErrInvalidOut
		}
		if isPtr {
			for rows.Next() {
				elem := reflect.New(ty)
				elemPtr := elem.Interface()
				pointers := getPointers(elemPtr, columns)
				if err := rows.Scan(pointers...); err != nil {
					return err
				}
				slice = reflect.Append(slice, elem)
			}
		} else {
			elem := reflect.New(ty)
			elemPtr := elem.Interface()
			pointers := getPointers(elemPtr, columns)
			for rows.Next() {
				if err := rows.Scan(pointers...); err != nil {
					return err
				}
				slice = reflect.Append(slice, elem.Elem())
			}
		}
		reflect.ValueOf(out).Elem().Set(slice)
		return rows.Err()
	default:
		return ErrInvalidOut
	}
}

// MustQueryRows ...
func MustQueryRows(out interface{}, tx _SQLCommon, query string, args ...interface{}) {
	if err := QueryRows(out, tx, query, args...); err != nil {
		panic(err)
	}
}
