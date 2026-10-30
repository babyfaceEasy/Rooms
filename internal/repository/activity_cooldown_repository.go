package repository

import (
	"context"
	"time"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActivityCooldownRepository defines the interface for activity cooldown persistence
type ActivityCooldownRepository interface {
	// GetOrCreate retrieves or creates an activity cooldown record
	GetOrCreate(ctx context.Context, userID, patchID primitive.ObjectID) (*domain.ActivityCooldown, error)

	// UpdateLastActivityAt updates when activity last occurred
	UpdateLastActivityAt(ctx context.Context, userID, patchID primitive.ObjectID, timestamp time.Time) error

	// UpdateLastNotificationSentAt updates when notification was sent
	UpdateLastNotificationSentAt(ctx context.Context, userID, patchID primitive.ObjectID, timestamp time.Time) error

	// UpdateLastCheckedAt updates when user viewed the patch
	UpdateLastCheckedAt(ctx context.Context, userID, patchID primitive.ObjectID, timestamp time.Time) error

	// ShouldSendNotification checks if enough time has passed since last notification
	ShouldSendNotification(ctx context.Context, userID, patchID primitive.ObjectID, cooldownMinutes int) (bool, error)

	// Reset clears all timestamps for a cooldown record
	Reset(ctx context.Context, userID, patchID primitive.ObjectID) error

	// Delete removes an activity cooldown record
	Delete(ctx context.Context, userID, patchID primitive.ObjectID) error
}
