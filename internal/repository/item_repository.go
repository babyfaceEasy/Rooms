package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/internal/domain"
)

// ItemRepository defines the persistence contract for domain.Item.
type ItemRepository interface {
	Create(ctx context.Context, item *domain.Item) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Item, error)
	List(ctx context.Context, page, limit int64) ([]domain.Item, error)
	Update(ctx context.Context, item *domain.Item) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
