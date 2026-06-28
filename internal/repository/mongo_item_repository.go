package repository

import (
	"context"
	"fmt"
	"time"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoItemRepository struct {
	collection *mongo.Collection
}

// NewMongoItemRepository returns an ItemRepository implementation backed by MongoDB.
func NewMongoItemRepository(db *mongo.Database) ItemRepository {
	repo := &mongoItemRepository{
		collection: db.Collection("items"),
	}
	// Attempt to create an index on created_at once at startup. Errors are not fatal
	// because the application can still operate with collection scans locally.
	_ = repo.ensureIndexes(context.Background())
	return repo
}

func (r *mongoItemRepository) ensureIndexes(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	index := mongo.IndexModel{
		Keys: bson.D{{Key: "created_at", Value: -1}},
	}
	_, err := r.collection.Indexes().CreateOne(ctx, index)
	return err
}

func (r *mongoItemRepository) Create(ctx context.Context, item *domain.Item) error {
	if item.ID.IsZero() {
		item.ID = primitive.NewObjectID()
	}
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, item)
	return err
}

func (r *mongoItemRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Item, error) {
	var item domain.Item
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&item); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("item with id %q not found: %w", id.Hex(), domain.ErrNotFound)
		}
		return nil, err
	}
	return &item, nil
}

func (r *mongoItemRepository) List(ctx context.Context, page, limit int64) ([]domain.Item, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	opts := options.Find().
		SetSkip((page - 1) * limit).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []domain.Item
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *mongoItemRepository) Update(ctx context.Context, item *domain.Item) error {
	item.UpdatedAt = time.Now().UTC()
	filter := bson.M{"_id": item.ID}
	result, err := r.collection.ReplaceOne(ctx, filter, item)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("item with id %q not found: %w", item.ID.Hex(), domain.ErrNotFound)
	}
	return nil
}

func (r *mongoItemRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("item with id %q not found: %w", id.Hex(), domain.ErrNotFound)
	}
	return nil
}
