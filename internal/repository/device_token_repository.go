package repository

import (
	"context"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeviceTokenRepository defines the interface for device token persistence
type DeviceTokenRepository interface {
	// Create stores a new device token
	Create(ctx context.Context, token *domain.DeviceToken) error

	// GetByID retrieves a device token by ID
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.DeviceToken, error)

	// GetByUserID retrieves all active device tokens for a user
	GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.DeviceToken, error)

	// GetActiveByUserID retrieves only active device tokens for a user
	GetActiveByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.DeviceToken, error)

	// Update updates an existing device token
	Update(ctx context.Context, token *domain.DeviceToken) error

	// Delete removes a device token
	Delete(ctx context.Context, id primitive.ObjectID) error

	// DeleteByUserID removes all device tokens for a user
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error

	// MarkInactive deactivates a device token
	MarkInactive(ctx context.Context, id primitive.ObjectID) error

	// DeleteStaleTokens removes device tokens older than expiryDays
	DeleteStaleTokens(ctx context.Context, expiryDays int) error

	// CheckTokenExists checks if a token already exists for a user
	CheckTokenExists(ctx context.Context, userID primitive.ObjectID, token string) (bool, error)
}
