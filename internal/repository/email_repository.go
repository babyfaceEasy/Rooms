package repository

import (
	"context"

	"temp_backend/internal/domain"
)

// EmailRepository defines methods for email persistence and operations.
type EmailRepository interface {
	// SaveEmailLog creates a new email log entry in the database.
	SaveEmailLog(ctx context.Context, email *domain.Email) (*domain.Email, error)

	// UpdateEmailStatus updates the status and optional error message of a logged email.
	UpdateEmailStatus(ctx context.Context, emailID string, status string, errorMsg *string) error

	// GetByID retrieves an email log by ID.
	GetByID(ctx context.Context, id string) (*domain.Email, error)

	// ListByUserID retrieves all email logs for a specific user.
	ListByUserID(ctx context.Context, userID string) ([]*domain.Email, error)
}
