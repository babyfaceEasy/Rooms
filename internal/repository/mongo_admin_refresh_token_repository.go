package repository

import (
	"context"
	"fmt"
	"time"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoAdminRefreshTokenRepository implements AdminRefreshTokenRepository for MongoDB.
type MongoAdminRefreshTokenRepository struct {
	collection *mongo.Collection
}

// NewMongoAdminRefreshTokenRepository creates a new MongoAdminRefreshTokenRepository.
func NewMongoAdminRefreshTokenRepository(db *mongo.Database) (*MongoAdminRefreshTokenRepository, error) {
	collection := db.Collection("admin_refresh_tokens")
	return &MongoAdminRefreshTokenRepository{collection: collection}, nil
}

// Create creates a new refresh token
func (r *MongoAdminRefreshTokenRepository) Create(ctx context.Context, token *domain.AdminRefreshToken) error {
	token.ID = primitive.NewObjectID()

	_, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to create admin refresh token: %w", err)
	}

	return nil
}

// Verify checks if a refresh token hash matches and is valid
func (r *MongoAdminRefreshTokenRepository) Verify(ctx context.Context, adminID primitive.ObjectID, tokenHash string) error {
	filter := bson.M{
		"admin_id":   adminID,
		"token_hash": tokenHash,
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	}

	result := r.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return domain.ErrTokenNotFound
		}
		return fmt.Errorf("failed to verify refresh token: %w", result.Err())
	}

	return nil
}

// DeleteByAdminID deletes all refresh tokens for an admin
func (r *MongoAdminRefreshTokenRepository) DeleteByAdminID(ctx context.Context, adminID primitive.ObjectID) error {
	filter := bson.M{
		"admin_id": adminID,
	}

	_, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete admin refresh tokens: %w", err)
	}

	return nil
}

// DeleteExpired deletes expired refresh tokens
func (r *MongoAdminRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	filter := bson.M{
		"expires_at": bson.M{"$lt": time.Now().UTC()},
	}

	_, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete expired admin refresh tokens: %w", err)
	}

	return nil
}
