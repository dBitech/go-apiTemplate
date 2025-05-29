package repository

import "errors"

// Common repository errors
var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrInternal      = errors.New("internal repository error")
	ErrInvalidData   = errors.New("invalid data")
)
