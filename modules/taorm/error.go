package taorm

import "errors"

var (
	ErrNoWhere    = errors.New("taorm: no wheres")
	ErrNoFields   = errors.New("taorm: no fields")
	ErrInvalidOut = errors.New("taorm: invalid out")
)
