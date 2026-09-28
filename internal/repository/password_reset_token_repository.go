package repository

import (
	"context"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PasswordResetTokenRepository defines the contract for password reset token data access.
type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *domain.PasswordResetToken) error
	GetByToken(ctx context.Context, token string) (*domain.PasswordResetToken, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID) (*domain.PasswordResetToken, error)
	MarkUsed(ctx context.Context, id primitive.ObjectID) error
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
}
