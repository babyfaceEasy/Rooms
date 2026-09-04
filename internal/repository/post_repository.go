package repository

import (
	"context"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostRepository defines the interface for post persistence
type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Post, error)
	DeletePost(ctx context.Context, id primitive.ObjectID) error
	GetByRoomID(ctx context.Context, roomID primitive.ObjectID, page, limit int) ([]*domain.Post, int64, error)
	AddValidation(ctx context.Context, postID, userID primitive.ObjectID) error
	RemoveValidation(ctx context.Context, postID, userID primitive.ObjectID) error
	AddRespect(ctx context.Context, postID, userID primitive.ObjectID) error
	RemoveRespect(ctx context.Context, postID, userID primitive.ObjectID) error
	IncrementReportCount(ctx context.Context, postID primitive.ObjectID) (int, error)
	DeleteByUserAndRoom(ctx context.Context, userID, roomID primitive.ObjectID) error
}
