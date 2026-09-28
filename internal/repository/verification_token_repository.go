package repository

import (
	"context"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// VerificationTokenRepository defines the contract for email verification token data access.
type VerificationTokenRepository interface {
	Create(ctx context.Context, token *domain.EmailVerificationToken) error
	GetByToken(ctx context.Context, token string) (*domain.EmailVerificationToken, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID) (*domain.EmailVerificationToken, error)
	MarkUsed(ctx context.Context, id primitive.ObjectID) error
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
}
