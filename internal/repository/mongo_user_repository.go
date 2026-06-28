package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"temp_backend/internal/domain"
)

// MongoUserRepository implements UserRepository for MongoDB.
type MongoUserRepository struct {
	collection *mongo.Collection
}

// NewMongoUserRepository creates a new MongoUserRepository and ensures indexes.
func NewMongoUserRepository(db *mongo.Database) (*MongoUserRepository, error) {
	collection := db.Collection("users")

	// Ensure unique index on email field
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "email", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create email index: %w", err)
	}

	return &MongoUserRepository{collection: collection}, nil
}

// Create inserts a new user into the database.
func (r *MongoUserRepository) Create(ctx context.Context, user *domain.User) error {
	user.ID = primitive.NewObjectID()

	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("email already exists: %w", domain.ErrEmailAlreadyExists)
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// GetByEmail retrieves a user by email address (excluding soft-deleted users).
func (r *MongoUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	filter := bson.M{
		"email":      strings.ToLower(email),
		"deleted_at": bson.M{"$eq": nil},
	}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("user not found: %w", domain.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	return &user, nil
}

// GetByID retrieves a user by MongoDB ObjectID (excluding soft-deleted users).
func (r *MongoUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	var user domain.User
	filter := bson.M{
		"_id":        id,
		"deleted_at": bson.M{"$eq": nil},
	}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("user not found: %w", domain.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	return &user, nil
}

// Update updates an existing user record.
func (r *MongoUserRepository) Update(ctx context.Context, user *domain.User) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{
			"$set": bson.M{
				"name":              user.Name,
				"email":             strings.ToLower(user.Email),
				"password_hash":     user.PasswordHash,
				"is_age_verified":   user.IsAgeVerified,
				"updated_at":        user.UpdatedAt,
			},
		},
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("email already exists: %w", domain.ErrEmailAlreadyExists)
		}
		return fmt.Errorf("failed to update user: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found: %w", domain.ErrUserNotFound)
	}
	return nil
}

// Delete removes a user from the database.
func (r *MongoUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("user not found: %w", domain.ErrUserNotFound)
	}
	return nil
}

// SoftDelete marks a user as deleted without removing the record (soft delete).
func (r *MongoUserRepository) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now().UTC()
	result, err := r.collection.UpdateOne(
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
		return fmt.Errorf("failed to soft delete user: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found: %w", domain.ErrUserNotFound)
	}
	return nil
}
