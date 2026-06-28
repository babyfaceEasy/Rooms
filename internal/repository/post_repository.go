package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/internal/domain"
)

// PostRepository defines the interface for post persistence
type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Post, error)
	DeletePost(ctx context.Context, id primitive.ObjectID) error
}
