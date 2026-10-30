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

type mongoDeviceTokenRepository struct {
	collection *mongo.Collection
}

// NewMongoDeviceTokenRepository creates a new MongoDB device token repository
func NewMongoDeviceTokenRepository(db *mongo.Database) DeviceTokenRepository {
	return &mongoDeviceTokenRepository{
		collection: db.Collection("device_tokens"),
	}
}

// Create stores a new device token
func (r *mongoDeviceTokenRepository) Create(ctx context.Context, token *domain.DeviceToken) error {
	if token.ID.IsZero() {
		token.ID = primitive.NewObjectID()
	}
	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to create device token: %w", err)
	}
	return nil
}

// GetByID retrieves a device token by ID
func (r *mongoDeviceTokenRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.DeviceToken, error) {
	filter := bson.M{"_id": id}
	result := r.collection.FindOne(ctx, filter)

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get device token: %w", result.Err())
	}

	token := &domain.DeviceToken{}
	if err := result.Decode(token); err != nil {
		return nil, fmt.Errorf("failed to decode device token: %w", err)
	}

	return token, nil
}

// GetByUserID retrieves all device tokens for a user (active and inactive)
func (r *mongoDeviceTokenRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.DeviceToken, error) {
	filter := bson.M{"user_id": userID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query device tokens: %w", err)
	}
	defer cursor.Close(ctx)

	var tokens []*domain.DeviceToken
	if err := cursor.All(ctx, &tokens); err != nil {
		return nil, fmt.Errorf("failed to decode device tokens: %w", err)
	}

	return tokens, nil
}

// GetActiveByUserID retrieves only active device tokens for a user
func (r *mongoDeviceTokenRepository) GetActiveByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.DeviceToken, error) {
	filter := bson.M{
		"user_id":   userID,
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query active device tokens: %w", err)
	}
	defer cursor.Close(ctx)

	var tokens []*domain.DeviceToken
	if err := cursor.All(ctx, &tokens); err != nil {
		return nil, fmt.Errorf("failed to decode device tokens: %w", err)
	}

	return tokens, nil
}

// Update updates an existing device token
func (r *mongoDeviceTokenRepository) Update(ctx context.Context, token *domain.DeviceToken) error {
	token.UpdatedAt = time.Now()

	filter := bson.M{"_id": token.ID}
	update := bson.M{
		"$set": bson.M{
			"token":      token.Token,
			"platform":   token.Platform,
			"is_active":  token.IsActive,
			"updated_at": token.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update device token: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete removes a device token
func (r *mongoDeviceTokenRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete device token: %w", err)
	}

	if result.DeletedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// DeleteByUserID removes all device tokens for a user
func (r *mongoDeviceTokenRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	filter := bson.M{"user_id": userID}

	_, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete device tokens for user: %w", err)
	}

	return nil
}

// MarkInactive deactivates a device token
func (r *mongoDeviceTokenRepository) MarkInactive(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to mark device token inactive: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// DeleteStaleTokens removes device tokens older than expiryDays
func (r *mongoDeviceTokenRepository) DeleteStaleTokens(ctx context.Context, expiryDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -expiryDays)
	filter := bson.M{
		"created_at": bson.M{
			"$lt": cutoffTime,
		},
	}

	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete stale device tokens: %w", err)
	}

	// Log how many were deleted if needed by caller
	_ = result.DeletedCount

	return nil
}

// CheckTokenExists checks if a token already exists for a user
func (r *mongoDeviceTokenRepository) CheckTokenExists(ctx context.Context, userID primitive.ObjectID, token string) (bool, error) {
	filter := bson.M{
		"user_id": userID,
		"token":   token,
	}

	result := r.collection.FindOne(ctx, filter)

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, fmt.Errorf("failed to check token existence: %w", result.Err())
	}

	return true, nil
}
