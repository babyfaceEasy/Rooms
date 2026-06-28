package domain

import "errors"

// Application-level sentinel errors. They are intentionally broad so that
// upper layers can map them to the correct HTTP status codes.
var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("resource already exists")
)
