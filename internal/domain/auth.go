package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JWTClaims represents custom JWT claims for user authentication.
type JWTClaims struct {
	UserID primitive.ObjectID `json:"user_id"`
	Email  string             `json:"email"`
	jwt.RegisteredClaims
}

// AdminJWTClaims represents custom JWT claims for admin authentication.
type AdminJWTClaims struct {
	AdminID     primitive.ObjectID `json:"admin_id"`
	Email       string             `json:"email"`
	Role        AdminRole          `json:"role"`
	Permissions []AdminPermission  `json:"permissions"`
	jwt.RegisteredClaims
}

// RefreshToken represents a refresh token stored in the database for users.
type RefreshToken struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	TokenHash string             `json:"-" bson:"token_hash"`
	ExpiresAt time.Time          `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

// AdminRefreshToken represents a refresh token stored in the database for admins.
type AdminRefreshToken struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	AdminID   primitive.ObjectID `json:"admin_id" bson:"admin_id"`
	TokenHash string             `json:"-" bson:"token_hash"`
	ExpiresAt time.Time          `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}
