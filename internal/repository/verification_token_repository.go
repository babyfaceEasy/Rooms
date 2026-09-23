package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/internal/domain"
)

// VerificationTokenRepository defines the contract for email verification token data access.
type VerificationTokenRepository interface {
	Create(ctx context.Context, token *domain.EmailVerificationToken) error
	GetByToken(ctx context.Context, token string) (*domain.EmailVerificationToken, error)
	MarkUsed(ctx context.Context, id primitive.ObjectID) error
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
}