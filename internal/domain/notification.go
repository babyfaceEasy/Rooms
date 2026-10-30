package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationType defines the type of notification
type NotificationType string

const (
	NotificationTypeAdded    NotificationType = "added"
	NotificationTypeJoined   NotificationType = "joined"
	NotificationTypeActivity NotificationType = "activity"
)

// NotificationStatus tracks the delivery status of a push notification
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
)

// Notification represents a notification sent to a user
type Notification struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	UserID           primitive.ObjectID `bson:"user_id"`
	DeviceTokenID    primitive.ObjectID `bson:"device_token_id,omitempty"`
	PatchID          primitive.ObjectID `bson:"patch_id,omitempty"`
	PostID           primitive.ObjectID `bson:"post_id,omitempty"`
	Reason           ReportReason       `bson:"reason,omitempty"`
	Message          string             `bson:"message"`
	NotificationType NotificationType   `bson:"notification_type"`
	Status           NotificationStatus `bson:"status"`
	IsRead           bool               `bson:"is_read"`
	SentAt           *time.Time         `bson:"sent_at,omitempty"`
	ErrorMessage     string             `bson:"error_message,omitempty"`
	FCMMessageID     string             `bson:"fcm_message_id,omitempty"`
	CreatedAt        time.Time          `bson:"created_at"`
}

