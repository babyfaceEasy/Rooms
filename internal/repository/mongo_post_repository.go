package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/internal/domain"
)

// MongoPostRepository implements PostRepository for MongoDB
type MongoPostRepository struct {
	collection *mongo.Collection
}

// NewMongoPostRepository creates a new MongoDB post repository
func NewMongoPostRepository(db *mongo.Database) *MongoPostRepository {
	collection := db.Collection("posts")
	return &MongoPostRepository{
		collection: collection,
	}
}

// Create creates a new post in the database
func (m *MongoPostRepository) Create(ctx context.Context, post *domain.Post) error {
	result, err := m.collection.InsertOne(ctx, post)
	if err != nil {
		return err
	}

	// Set the ID to the inserted ID
	post.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves a post by its ID (excluding soft-deleted)
func (m *MongoPostRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Post, error) {
	var post domain.Post
	err := m.collection.FindOne(ctx, bson.M{
		"_id":        id,
		"deleted_at": nil,
	}).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrPostNotFound
		}
		return nil, err
	}
	return &post, nil
}

// DeletePost soft-deletes a post by marking it as deleted
func (m *MongoPostRepository) DeletePost(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	result, err := m.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"deleted_at": now,
				"updated_at": now,
			},
		},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrPostNotFound
	}

	return nil
}
