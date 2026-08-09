package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Email represents an email record in the database for tracking email sending attempts.
type Email struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID            primitive.ObjectID `bson:"user_id" json:"user_id"`
	RecipientEmail    string             `bson:"recipient_email" json:"recipient_email"`
	TemplateID        string             `bson:"template_id" json:"template_id"`
	DynamicData       map[string]string  `bson:"dynamic_data" json:"dynamic_data"`
	Status            string             `bson:"status" json:"status"` // "pending", "sent", "failed"
	ErrorMessage      *string            `bson:"error_message" json:"error_message"`
	SendgridMessageID *string            `bson:"sendgrid_message_id" json:"sendgrid_message_id"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	SentAt            *time.Time         `bson:"sent_at" json:"sent_at"`
}

// Email statuses
const (
	EmailStatusPending = "pending"
	EmailStatusSent    = "sent"
	EmailStatusFailed  = "failed"
)
