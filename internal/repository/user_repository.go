package repository

import (
	"context"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRepository defines the contract for user data access operations.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*domain.User, error)
	GetByCode(ctx context.Context, code string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	SoftDelete(ctx context.Context, id primitive.ObjectID) error

	// Admin operations
	SuspendUser(ctx context.Context, userID primitive.ObjectID, reason string, suspendedBy primitive.ObjectID) error
	UnsuspendUser(ctx context.Context, userID primitive.ObjectID) error
	ListSuspendedUsers(ctx context.Context, limit int, offset int) ([]*domain.User, error)
	ListDeletedUsers(ctx context.Context, limit int, offset int) ([]*domain.User, error)
	CountAllUsers(ctx context.Context) (int, error)
	CountDeletedUsers(ctx context.Context) (int, error)
	CountSuspendedUsers(ctx context.Context) (int, error)
	CountVerifiedUsers(ctx context.Context) (int, error)
	ListAllUsers(ctx context.Context, limit int, offset int) ([]*domain.User, error)
}

