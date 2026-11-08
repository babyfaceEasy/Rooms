package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a customer account with authentication credentials and age verification.
type User struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Code           string             `json:"code" bson:"code"`
	Name           string             `json:"name" bson:"name"`
	Email          string             `json:"email" bson:"email"`
	PasswordHash   string             `json:"-" bson:"password_hash"`
	IsAgeVerified    bool               `json:"is_age_verified" bson:"is_age_verified"`
	IsEmailVerified  bool               `json:"is_email_verified" bson:"is_email_verified"`
	ProfilePicture   string             `json:"profile_picture,omitempty" bson:"profile_picture,omitempty"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt      *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`

	// Admin suspension
	IsSuspended      bool                `json:"is_suspended" bson:"is_suspended"`
	SuspensionReason *string             `json:"suspension_reason,omitempty" bson:"suspension_reason,omitempty"`
	SuspendedAt      *time.Time          `json:"suspended_at,omitempty" bson:"suspended_at,omitempty"`
	SuspendedBy      *primitive.ObjectID `json:"suspended_by,omitempty" bson:"suspended_by,omitempty"`
}
