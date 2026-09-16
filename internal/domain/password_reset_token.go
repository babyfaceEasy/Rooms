package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PasswordResetToken represents a token used to reset a user's password.
type PasswordResetToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Token     string             `bson:"token" json:"token"`
	ExpiresAt time.Time          `bson:"expires_at" json:"expires_at"`
	UsedAt    *time.Time         `bson:"used_at,omitempty" json:"used_at,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}