package repository

import (
	"context"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CommentRepository defines the interface for comment persistence
type CommentRepository interface {
	Create(ctx context.Context, comment *domain.Comment) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Comment, error)
	GetByPostID(ctx context.Context, postID primitive.ObjectID, page, limit int) ([]*domain.Comment, int64, error)
	DeleteComment(ctx context.Context, id primitive.ObjectID) error
}
