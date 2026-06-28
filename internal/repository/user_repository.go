package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/internal/domain"
)

// UserRepository defines the contract for user data access operations.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	SoftDelete(ctx context.Context, id primitive.ObjectID) error
}
