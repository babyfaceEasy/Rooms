package domain

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrCommentTextRequired = errors.New("comment text is required")
	ErrCommentTextTooLong  = errors.New("comment text exceeds maximum length")
	ErrCommentNotFound     = errors.New("comment not found")
	ErrUnauthorizedComment = errors.New("you are not authorized to perform this action on this comment")
)

// Comment represents a comment on a post
type Comment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	PostID    primitive.ObjectID `bson:"post_id"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Text      string             `bson:"text"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty"` // Soft delete
}

// ValidateComment validates the Comment entity
func (c *Comment) ValidateComment() error {
	if c.Text == "" {
		return ErrCommentTextRequired
	}

	if len(c.Text) < 1 {
		return ErrCommentTextRequired
	}

	if len(c.Text) > 1000 {
		return ErrCommentTextTooLong
	}

	if c.PostID.IsZero() {
		return ErrInvalidInput
	}

	if c.UserID.IsZero() {
		return ErrInvalidInput
	}

	return nil
}
