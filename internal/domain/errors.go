package domain

import "errors"

// Application-level sentinel errors. They are intentionally broad so that
// upper layers can map them to the correct HTTP status codes.
var (
	ErrNotFound                = errors.New("resource not found")
	ErrInvalidInput            = errors.New("invalid input")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrForbidden               = errors.New("forbidden")
	ErrConflict                = errors.New("resource already exists")
	ErrEmailAlreadyExists      = errors.New("email already exists")
	ErrInvalidEmail            = errors.New("invalid email format")
	ErrInvalidPassword         = errors.New("invalid password")
	ErrAgeVerificationRequired = errors.New("age verification required")
	ErrUserNotFound            = errors.New("user not found")
	ErrInvalidCredentials      = errors.New("invalid email or password")
	ErrTokenExpired            = errors.New("token has expired")
	ErrTokenInvalid            = errors.New("invalid token")
	ErrTokenNotFound           = errors.New("token not found")
)
