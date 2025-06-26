package errors

import "errors"

var (
	ErrURLNotFound = errors.New("URL not found")
	ErrURLDeleted  = errors.New("URL is deleted")
)
