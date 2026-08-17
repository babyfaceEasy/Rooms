package domain

import (
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrCommentTextRequired = &AppError{Code: "COMMENT_TEXT_REQUIRED", Message: "Comment text is required", HTTPStatus: http.StatusBadRequest}
	ErrCommentTextTooLong  = &AppError{Code: "COMMENT_TEXT_TOO_LONG", Message: "Comment text exceeds maximum length of 1000 characters", HTTPStatus: http.StatusBadRequest}
	ErrCommentNotFound     = &AppError{Code: "COMMENT_NOT_FOUND", Message: "Comment not found", HTTPStatus: http.StatusNotFound}
	ErrUnauthorizedComment = &AppError{Code: "UNAUTHORIZED_COMMENT", Message: "You are not authorized to perform this action on this comment", HTTPStatus: http.StatusForbidden}
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
