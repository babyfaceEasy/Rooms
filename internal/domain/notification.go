package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification represents a notification sent to a user
type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"` // Recipient (post creator)
	PostID    primitive.ObjectID `bson:"post_id"`
	Reason    ReportReason       `bson:"reason"` // Why it was reported
	Message   string             `bson:"message"`
	IsRead    bool               `bson:"is_read"`
	CreatedAt time.Time          `bson:"created_at"`
}
