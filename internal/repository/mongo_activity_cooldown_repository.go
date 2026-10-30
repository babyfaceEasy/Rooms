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

type mongoActivityCooldownRepository struct {
	collection *mongo.Collection
}

// NewMongoActivityCooldownRepository creates a new MongoDB activity cooldown repository
func NewMongoActivityCooldownRepository(db *mongo.Database) ActivityCooldownRepository {
	return &mongoActivityCooldownRepository{
		collection: db.Collection("activity_cooldowns"),
	}
}

// GetOrCreate retrieves an existing cooldown record or creates a new one
func (r *mongoActivityCooldownRepository) GetOrCreate(ctx context.Context, userID, patchID primitive.ObjectID) (*domain.ActivityCooldown, error) {
	filter := bson.M{
		"user_id":  userID,
		"patch_id": patchID,
	}

	result := r.collection.FindOne(ctx, filter)

	if result.Err() == nil {
		cooldown := &domain.ActivityCooldown{}
		if err := result.Decode(cooldown); err != nil {
			return nil, fmt.Errorf("failed to decode activity cooldown: %w", err)
		}
		return cooldown, nil
	}

	if result.Err() != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("failed to query activity cooldown: %w", result.Err())
	}

	// Create new cooldown record
	now := time.Now()
	cooldown := &domain.ActivityCooldown{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		PatchID:   patchID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := r.collection.InsertOne(ctx, cooldown)
	if err != nil {
		// If insertion fails due to duplicate, try to fetch again
		if mongo.IsDuplicateKeyError(err) {
			return r.GetOrCreate(ctx, userID, patchID)
		}
		return nil, fmt.Errorf("failed to create activity cooldown: %w", err)
	}

	return cooldown, nil
}

// UpdateLastActivityAt updates when activity last occurred
func (r *mongoActivityCooldownRepository) UpdateLastActivityAt(ctx context.Context, userID, patchID primitive.ObjectID, timestamp time.Time) error {
	filter := bson.M{
		"user_id":  userID,
		"patch_id": patchID,
	}

	update := bson.M{
		"$set": bson.M{
			"last_activity_at": timestamp,
			"updated_at":       time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update last activity time: %w", err)
	}

	if result.MatchedCount == 0 && result.UpsertedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// UpdateLastNotificationSentAt updates when notification was sent
func (r *mongoActivityCooldownRepository) UpdateLastNotificationSentAt(ctx context.Context, userID, patchID primitive.ObjectID, timestamp time.Time) error {
	filter := bson.M{
		"user_id":  userID,
		"patch_id": patchID,
	}

	update := bson.M{
		"$set": bson.M{
			"last_notification_sent_at": timestamp,
			"updated_at":                time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update last notification sent time: %w", err)
	}

	if result.MatchedCount == 0 && result.UpsertedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// UpdateLastCheckedAt updates when user viewed the patch
func (r *mongoActivityCooldownRepository) UpdateLastCheckedAt(ctx context.Context, userID, patchID primitive.ObjectID, timestamp time.Time) error {
	filter := bson.M{
		"user_id":  userID,
		"patch_id": patchID,
	}

	update := bson.M{
		"$set": bson.M{
			"last_checked_at": timestamp,
			"updated_at":      time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update last checked time: %w", err)
	}

	if result.MatchedCount == 0 && result.UpsertedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// ShouldSendNotification checks if enough time has passed since last notification
func (r *mongoActivityCooldownRepository) ShouldSendNotification(ctx context.Context, userID, patchID primitive.ObjectID, cooldownMinutes int) (bool, error) {
	cooldown, err := r.GetOrCreate(ctx, userID, patchID)
	if err != nil {
		return false, err
	}

	// If no notification has been sent yet, send it
	if cooldown.LastNotificationSentAt == nil {
		return true, nil
	}

	// Check if cooldown period has passed
	elapsed := time.Since(*cooldown.LastNotificationSentAt)
	cooldownDuration := time.Duration(cooldownMinutes) * time.Minute

	return elapsed >= cooldownDuration, nil
}

// Reset clears all timestamps for a cooldown record
func (r *mongoActivityCooldownRepository) Reset(ctx context.Context, userID, patchID primitive.ObjectID) error {
	filter := bson.M{
		"user_id":  userID,
		"patch_id": patchID,
	}

	update := bson.M{
		"$set": bson.M{
			"last_notification_sent_at": nil,
			"last_activity_at":          nil,
			"last_checked_at":           nil,
			"updated_at":                time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to reset activity cooldown: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete removes an activity cooldown record
func (r *mongoActivityCooldownRepository) Delete(ctx context.Context, userID, patchID primitive.ObjectID) error {
	filter := bson.M{
		"user_id":  userID,
		"patch_id": patchID,
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete activity cooldown: %w", err)
	}

	if result.DeletedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}
