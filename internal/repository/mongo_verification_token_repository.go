package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoVerificationTokenRepository implements VerificationTokenRepository for MongoDB.
type MongoVerificationTokenRepository struct {
	collection *mongo.Collection
}

// NewMongoVerificationTokenRepository creates a new MongoVerificationTokenRepository and ensures indexes.
func NewMongoVerificationTokenRepository(db *mongo.Database) (*MongoVerificationTokenRepository, error) {
	collection := db.Collection("verification_tokens")

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

	return &MongoVerificationTokenRepository{collection: collection}, nil
}

// Create inserts a new verification token into the database.
func (r *MongoVerificationTokenRepository) Create(ctx context.Context, token *domain.EmailVerificationToken) error {
	token.ID = primitive.NewObjectID()
	token.CreatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to insert verification token: %w", err)
	}

	return nil
}

// GetByToken retrieves a verification token by its token string.
func (r *MongoVerificationTokenRepository) GetByToken(ctx context.Context, token string) (*domain.EmailVerificationToken, error) {
	var vt domain.EmailVerificationToken
	err := r.collection.FindOne(ctx, bson.M{"token": token}).Decode(&vt)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("verification token not found: %w", domain.ErrInvalidInput)
		}
		return nil, fmt.Errorf("failed to query verification token: %w", err)
	}
	return &vt, nil
}

// GetByUserID retrieves the most recent verification token for a user (if exists).
func (r *MongoVerificationTokenRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) (*domain.EmailVerificationToken, error) {
	var vt domain.EmailVerificationToken
	opts := options.FindOne().SetSort(bson.M{"created_at": -1})

	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}, opts).Decode(&vt)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // No token found; this is not an error
		}
		return nil, fmt.Errorf("failed to query verification token by user_id: %w", err)
	}
	return &vt, nil
}

// MarkUsed marks a verification token as used.
func (r *MongoVerificationTokenRepository) MarkUsed(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now().UTC()
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{"used_at": now},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("verification token not found: %w", domain.ErrInvalidInput)
	}
	return nil
}

// DeleteByUserID removes all verification tokens for a user.
func (r *MongoVerificationTokenRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return fmt.Errorf("failed to delete verification tokens: %w", err)
	}
	return nil
}
