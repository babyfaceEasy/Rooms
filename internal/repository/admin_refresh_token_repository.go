package repository

import (
	"context"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminRefreshTokenRepository defines operations for admin refresh token persistence
type AdminRefreshTokenRepository interface {
	// Create creates a new refresh token
	Create(ctx context.Context, token *domain.AdminRefreshToken) error

	// Verify checks if a refresh token hash matches and is valid
	Verify(ctx context.Context, adminID primitive.ObjectID, tokenHash string) error

	// DeleteByAdminID deletes all refresh tokens for an admin
	DeleteByAdminID(ctx context.Context, adminID primitive.ObjectID) error

	// DeleteExpired deletes expired refresh tokens
	DeleteExpired(ctx context.Context) error
}
