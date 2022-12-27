package pg

import "errors"

var (
	ErrInvalidOrderByColumn = errors.New("invalid OrderBy column")
	ErrInvalidOrderByFn     = errors.New("invalid OrderBy function")
	ErrInvalidFilterValue   = errors.New("invalid filter value")
)
