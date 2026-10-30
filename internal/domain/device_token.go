package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeviceTokenPlatform defines the platform type for a device
type DeviceTokenPlatform string

const (
	PlatformIOS     DeviceTokenPlatform = "ios"
	PlatformAndroid DeviceTokenPlatform = "android"
	PlatformWeb     DeviceTokenPlatform = "web"
)

// DeviceToken represents an FCM device token for a user
type DeviceToken struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty"`
	UserID    primitive.ObjectID  `bson:"user_id"`
	Token     string              `bson:"token"`
	Platform  DeviceTokenPlatform `bson:"platform"`
	IsActive  bool                `bson:"is_active"`
	CreatedAt time.Time           `bson:"created_at"`
	UpdatedAt time.Time           `bson:"updated_at"`
}
