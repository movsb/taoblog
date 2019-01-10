package taorm

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

func WrapError(err error) error {
	if myErr, ok := err.(*mysql.MySQLError); ok {
		switch myErr.Number {
		case 1062:
			return &DupKeyError{}
		}
	}
	return err
}

var (
	ErrNoWhere    = errors.New("taorm: no wheres")
	ErrNoFields   = errors.New("taorm: no fields")
	ErrInvalidOut = errors.New("taorm: invalid out")
	ErrDupKey     = errors.New("taorm: dup key")
)

type DupKeyError struct {
}

func (e DupKeyError) Error() string {
	return "dup key error"
}
