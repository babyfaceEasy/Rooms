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

// MongoPasswordResetTokenRepository implements PasswordResetTokenRepository for MongoDB.
type MongoPasswordResetTokenRepository struct {
	collection *mongo.Collection
}

// NewMongoPasswordResetTokenRepository creates a new MongoPasswordResetTokenRepository and ensures indexes.
func NewMongoPasswordResetTokenRepository(db *mongo.Database) (*MongoPasswordResetTokenRepository, error) {
	collection := db.Collection("password_reset_tokens")

	tokenIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "token", Value: 1},
		},
	}

	_, err := collection.Indexes().CreateOne(context.Background(), tokenIndexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create token index: %w", err)
	}

	userIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
		},
	}

	_, err = collection.Indexes().CreateOne(context.Background(), userIndexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create user_id index: %w", err)
	}

	return &MongoPasswordResetTokenRepository{collection: collection}, nil
}

// Create inserts a new password reset token into the database.
func (r *MongoPasswordResetTokenRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	token.ID = primitive.NewObjectID()
	token.CreatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to insert password reset token: %w", err)
	}

	return nil
}

// GetByToken retrieves a password reset token by its token string.
func (r *MongoPasswordResetTokenRepository) GetByToken(ctx context.Context, token string) (*domain.PasswordResetToken, error) {
	var rt domain.PasswordResetToken
	err := r.collection.FindOne(ctx, bson.M{"token": token}).Decode(&rt)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("password reset token not found: %w", domain.ErrInvalidInput)
		}
		return nil, fmt.Errorf("failed to query password reset token: %w", err)
	}
	return &rt, nil
}

// MarkUsed marks a password reset token as used.
func (r *MongoPasswordResetTokenRepository) MarkUsed(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now().UTC()
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{"used_at": now},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to mark password reset token as used: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("password reset token not found: %w", domain.ErrInvalidInput)
	}
	return nil
}

// DeleteByUserID removes all password reset tokens for a user.
func (r *MongoPasswordResetTokenRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return fmt.Errorf("failed to delete password reset tokens: %w", err)
	}
	return nil
}