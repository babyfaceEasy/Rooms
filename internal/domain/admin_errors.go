package domain

import "net/http"

// Admin-specific errors
var (
	// 4xx errors
	ErrAdminNotFound             = &AppError{Code: "ADMIN_NOT_FOUND", Message: "Admin not found", HTTPStatus: http.StatusNotFound}
	ErrAdminAlreadyExists        = &AppError{Code: "ADMIN_ALREADY_EXISTS", Message: "An admin with this email already exists", HTTPStatus: http.StatusConflict}
	ErrInvalidAdminRole          = &AppError{Code: "INVALID_ADMIN_ROLE", Message: "Invalid admin role", HTTPStatus: http.StatusBadRequest}
	ErrInvalidSetupToken         = &AppError{Code: "INVALID_SETUP_TOKEN", Message: "Invalid or expired setup token", HTTPStatus: http.StatusBadRequest}
	ErrSetupTokenExpired         = &AppError{Code: "SETUP_TOKEN_EXPIRED", Message: "Setup token has expired", HTTPStatus: http.StatusUnauthorized}
	ErrUnauthorizedAdmin         = &AppError{Code: "UNAUTHORIZED_ADMIN", Message: "Admin authentication is required", HTTPStatus: http.StatusUnauthorized}
	ErrInsufficientPermissions   = &AppError{Code: "INSUFFICIENT_PERMISSIONS", Message: "You do not have permission to perform this action", HTTPStatus: http.StatusForbidden}
	ErrAdminLocked               = &AppError{Code: "ADMIN_LOCKED", Message: "Account is locked due to multiple failed login attempts", HTTPStatus: http.StatusForbidden}
	ErrAdminInactive             = &AppError{Code: "ADMIN_INACTIVE", Message: "Admin account is inactive", HTTPStatus: http.StatusForbidden}
	ErrAdminPasswordResetExpired = &AppError{Code: "PASSWORD_RESET_EXPIRED", Message: "Password reset token has expired", HTTPStatus: http.StatusUnauthorized}
	ErrAdminInvalidPassword      = &AppError{Code: "ADMIN_INVALID_PASSWORD", Message: "Invalid password", HTTPStatus: http.StatusBadRequest}
	ErrCannotDeleteSelfAdmin     = &AppError{Code: "CANNOT_DELETE_SELF", Message: "You cannot delete your own admin account", HTTPStatus: http.StatusBadRequest}
	ErrCannotDeleteLastSuperAdmin = &AppError{Code: "CANNOT_DELETE_LAST_SUPER_ADMIN", Message: "Cannot delete the last super admin account", HTTPStatus: http.StatusBadRequest}
)
