package domain

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrPostTextRequired = errors.New("post text is required")
	ErrPostTextTooLong  = errors.New("post text exceeds maximum length")
	ErrPostTextTooShort = errors.New("post text must be at least 1 character")
	ErrPostNotFound     = errors.New("post not found")
	ErrUnauthorizedPost = errors.New("you are not authorized to perform this action on this post")
	ErrNotRoomMember    = errors.New("not a member of this room")
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
