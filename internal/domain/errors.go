package domain

import "net/http"

// AppError is a structured application error with a machine-readable code,
// a user-facing message, an HTTP status, and an optional wrapped internal cause.
type AppError struct {
	Code       string // machine-readable, e.g. "ROOM_NOT_FOUND"
	Message    string // user-facing display text
	HTTPStatus int    // HTTP status code
	Err        error  // wrapped internal cause
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Application-level sentinel errors. Upper layers extract the *AppError with
// errors.As to get the code, message, and HTTP status.
var (
	// 4xx errors
	ErrNotFound                = &AppError{Code: "NOT_FOUND", Message: "Resource not found", HTTPStatus: http.StatusNotFound}
	ErrInvalidInput            = &AppError{Code: "INVALID_INPUT", Message: "The provided input is invalid", HTTPStatus: http.StatusBadRequest}
	ErrUnauthorized            = &AppError{Code: "UNAUTHORIZED", Message: "Authentication is required", HTTPStatus: http.StatusUnauthorized}
	ErrForbidden               = &AppError{Code: "FORBIDDEN", Message: "You do not have permission to perform this action", HTTPStatus: http.StatusForbidden}
	ErrConflict                = &AppError{Code: "CONFLICT", Message: "Resource already exists", HTTPStatus: http.StatusConflict}
	ErrEmailAlreadyExists      = &AppError{Code: "EMAIL_ALREADY_EXISTS", Message: "An account with this email already exists", HTTPStatus: http.StatusConflict}
	ErrInvalidEmail            = &AppError{Code: "INVALID_EMAIL", Message: "Invalid email format", HTTPStatus: http.StatusBadRequest}
	ErrInvalidPassword         = &AppError{Code: "INVALID_PASSWORD", Message: "Invalid password", HTTPStatus: http.StatusBadRequest}
	ErrAgeVerificationRequired = &AppError{Code: "AGE_VERIFICATION_REQUIRED", Message: "Age verification is required", HTTPStatus: http.StatusBadRequest}
	ErrUserNotFound            = &AppError{Code: "USER_NOT_FOUND", Message: "User not found", HTTPStatus: http.StatusNotFound}
	ErrInvalidCredentials      = &AppError{Code: "INVALID_CREDENTIALS", Message: "Invalid email or password", HTTPStatus: http.StatusUnauthorized}
	ErrTokenExpired            = &AppError{Code: "TOKEN_EXPIRED", Message: "Token has expired", HTTPStatus: http.StatusUnauthorized}
	ErrTokenInvalid            = &AppError{Code: "TOKEN_INVALID", Message: "Invalid token", HTTPStatus: http.StatusUnauthorized}
	ErrTokenNotFound           = &AppError{Code: "TOKEN_NOT_FOUND", Message: "Token not found", HTTPStatus: http.StatusNotFound}
	ErrCodeAlreadyExists       = &AppError{Code: "CODE_ALREADY_EXISTS", Message: "Room code already exists", HTTPStatus: http.StatusConflict}
	ErrRoomNotFound            = &AppError{Code: "ROOM_NOT_FOUND", Message: "This room doesn't exist or has been deleted", HTTPStatus: http.StatusNotFound}
	ErrNotRoomMember           = &AppError{Code: "NOT_ROOM_MEMBER", Message: "You are not a member of this room", HTTPStatus: http.StatusForbidden}
	ErrCannotJoinOwnRoom       = &AppError{Code: "CANNOT_JOIN_OWN_ROOM", Message: "You are the creator of this room and are already a member", HTTPStatus: http.StatusBadRequest}
	ErrOwnerCannotLeaveRoom    = &AppError{Code: "OWNER_CANNOT_LEAVE", Message: "Room owners cannot leave their own room. Delete the room instead.", HTTPStatus: http.StatusForbidden}

	// 5xx errors
	ErrInternalServer  = &AppError{Code: "INTERNAL_ERROR", Message: "Something went wrong. Please try again later.", HTTPStatus: http.StatusInternalServerError}
	ErrEmailSendFailed = &AppError{Code: "EMAIL_SEND_FAILED", Message: "Failed to send email", HTTPStatus: http.StatusInternalServerError}
)