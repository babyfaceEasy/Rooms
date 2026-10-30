package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActivityCooldown tracks the activity notification cooldown for a user in a specific patch
type ActivityCooldown struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty"`
	UserID                primitive.ObjectID `bson:"user_id"`
	PatchID               primitive.ObjectID `bson:"patch_id"`
	LastNotificationSentAt *time.Time         `bson:"last_notification_sent_at,omitempty"`
	LastActivityAt        *time.Time         `bson:"last_activity_at,omitempty"`
	LastCheckedAt         *time.Time         `bson:"last_checked_at,omitempty"`
	CreatedAt             time.Time          `bson:"created_at"`
	UpdatedAt             time.Time          `bson:"updated_at"`
}
