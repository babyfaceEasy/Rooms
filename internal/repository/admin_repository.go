package repository

import (
	"context"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminRepository defines operations for admin persistence
type AdminRepository interface {
	// Create creates a new admin
	Create(ctx context.Context, admin *domain.Admin) error

	// GetByID retrieves an admin by ID
	GetByID(ctx context.Context, adminID primitive.ObjectID) (*domain.Admin, error)

	// GetByEmail retrieves an admin by email
	GetByEmail(ctx context.Context, email string) (*domain.Admin, error)

	// List retrieves all active admins (paginated)
	List(ctx context.Context, limit int, offset int) ([]*domain.Admin, error)

	// ListWithFilters retrieves admins with optional filters (paginated)
	ListWithFilters(ctx context.Context, filters map[string]interface{}, limit int, offset int) ([]*domain.Admin, error)

	// Update updates an existing admin
	Update(ctx context.Context, admin *domain.Admin) error

	// UpdateLastLogin updates the last login timestamp
	UpdateLastLogin(ctx context.Context, adminID primitive.ObjectID) error

	// Delete performs a soft delete of an admin
	Delete(ctx context.Context, adminID primitive.ObjectID) error

	// SetupToken sets a setup token for initial password setup
	SetupToken(ctx context.Context, adminID primitive.ObjectID, token string, expiresAt interface{}) error

	// VerifySetupToken checks if a setup token is valid
	VerifySetupToken(ctx context.Context, adminID primitive.ObjectID, token string) (bool, error)

	// ClearSetupToken clears the setup token after successful setup
	ClearSetupToken(ctx context.Context, adminID primitive.ObjectID) error

	// SetPasswordResetToken sets a password reset token
	SetPasswordResetToken(ctx context.Context, adminID primitive.ObjectID, token string, expiresAt interface{}) error

	// VerifyPasswordResetToken checks if a password reset token is valid
	VerifyPasswordResetToken(ctx context.Context, adminID primitive.ObjectID, token string) (bool, error)

	// ClearPasswordResetToken clears the password reset token
	ClearPasswordResetToken(ctx context.Context, adminID primitive.ObjectID) error

	// UpdateFailedLoginAttempts increments failed login attempts
	UpdateFailedLoginAttempts(ctx context.Context, adminID primitive.ObjectID, attempts int, lockedUntil interface{}) error

	// ClearFailedLoginAttempts resets failed login attempts
	ClearFailedLoginAttempts(ctx context.Context, adminID primitive.ObjectID) error

	// CountSuperAdmins counts the number of active super admin accounts
	CountSuperAdmins(ctx context.Context) (int, error)
}

// AdminLogRepository defines operations for admin log persistence
type AdminLogRepository interface {
	// Create creates a new admin log entry
	Create(ctx context.Context, log *domain.AdminLog) error

	// List retrieves logs (paginated)
	List(ctx context.Context, limit int, offset int) ([]*domain.AdminLog, error)

	// ListByAdmin retrieves logs for a specific admin (paginated)
	ListByAdmin(ctx context.Context, adminID primitive.ObjectID, limit int, offset int) ([]*domain.AdminLog, error)

	// ListByResource retrieves logs for a specific resource (paginated)
	ListByResource(ctx context.Context, resourceType domain.AdminLogResourceType, resourceID primitive.ObjectID, limit int, offset int) ([]*domain.AdminLog, error)

	// ListByAction retrieves logs for a specific action (paginated)
	ListByAction(ctx context.Context, action domain.AdminLogAction, limit int, offset int) ([]*domain.AdminLog, error)

	// ListWithFilters retrieves logs with multiple filters (paginated)
	ListWithFilters(ctx context.Context, filters map[string]interface{}, limit int, offset int) ([]*domain.AdminLog, error)
}
