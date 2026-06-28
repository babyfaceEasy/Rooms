package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"temp_backend/internal/domain"
)

// MongoRefreshTokenRepository implements RefreshTokenRepository for MongoDB.
type MongoRefreshTokenRepository struct {
	collection *mongo.Collection
}

// NewMongoRefreshTokenRepository creates a new MongoRefreshTokenRepository and ensures indexes.
func NewMongoRefreshTokenRepository(db *mongo.Database) (*MongoRefreshTokenRepository, error) {
	collection := db.Collection("refresh_tokens")

	// Ensure compound index on (user_id, expires_at) for efficient queries
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "expires_at", Value: 1},
		},
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	// Also create index on token_hash for lookups
	hashIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "token_hash", Value: 1},
		},
	}

	_, err = collection.Indexes().CreateOne(context.Background(), hashIndexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create token_hash index: %w", err)
	}

	return &MongoRefreshTokenRepository{collection: collection}, nil
}

// Create inserts a new refresh token into the database.
func (r *MongoRefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	token.ID = primitive.NewObjectID()
	token.CreatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to insert refresh token: %w", err)
	}

	return nil
}

// GetByTokenHash retrieves a refresh token by its hash.
func (r *MongoRefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	err := r.collection.FindOne(ctx, bson.M{"token_hash": tokenHash}).Decode(&token)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("refresh token not found: %w", domain.ErrTokenInvalid)
		}
		return nil, fmt.Errorf("failed to query refresh token: %w", err)
	}

	// Check if token has expired
	if time.Now().UTC().After(token.ExpiresAt) {
		return nil, fmt.Errorf("refresh token expired: %w", domain.ErrTokenExpired)
	}

	return &token, nil
}

// Delete removes a refresh token by ID.
func (r *MongoRefreshTokenRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("refresh token not found: %w", domain.ErrTokenInvalid)
	}
	return nil
}

// DeleteByUserID removes all refresh tokens for a user.
func (r *MongoRefreshTokenRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return fmt.Errorf("failed to delete user refresh tokens: %w", err)
	}
	return nil
}
