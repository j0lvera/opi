package crud

import "errors"

var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrBadRequest = errors.New("bad request")
	ErrInternal   = errors.New("internal server error")
)
