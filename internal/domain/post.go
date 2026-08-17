package domain

import (
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrPostTextRequired = &AppError{Code: "POST_TEXT_REQUIRED", Message: "Post text is required", HTTPStatus: 400}
	ErrPostTextTooLong  = &AppError{Code: "POST_TEXT_TOO_LONG", Message: "Post text exceeds maximum length (5000 characters)", HTTPStatus: 400}
	ErrPostTextTooShort = &AppError{Code: "POST_TEXT_TOO_SHORT", Message: "Post text must be at least 1 character", HTTPStatus: 400}
	ErrPostNotFound     = &AppError{Code: "POST_NOT_FOUND", Message: "Post not found", HTTPStatus: 404}
	ErrUnauthorizedPost = &AppError{Code: "UNAUTHORIZED_POST", Message: "You are not authorized to perform this action on this post", HTTPStatus: http.StatusForbidden}
)

// Post represents a social media post
type Post struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	RoomID    primitive.ObjectID `bson:"room_id"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Text      string             `bson:"text"`
	Image     *string            `bson:"image,omitempty"` // S3 URL
	Video     *string            `bson:"video,omitempty"` // S3 URL
	Audio     *string            `bson:"audio,omitempty"` // S3 URL
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty"` // Soft delete
}

// ValidatePost validates the Post entity
func (p *Post) Validate() error {
	if p.RoomID.IsZero() {
		return ErrInvalidInput
	}

	if p.Text == "" {
		return ErrPostTextRequired
	}

	if len(p.Text) < 1 {
		return ErrPostTextTooShort
	}

	if len(p.Text) > 5000 {
		return ErrPostTextTooLong
	}

	return nil
}
