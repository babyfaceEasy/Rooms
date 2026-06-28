package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/internal/domain"
)

// RefreshTokenRepository defines the contract for refresh token data access operations.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
}
