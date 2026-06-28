package domain

import (
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Room struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name      string               `bson:"name" json:"name"`
	Code      string               `bson:"code" json:"code"`
	CreatedBy primitive.ObjectID   `bson:"created_by" json:"created_by"`
	Members   []primitive.ObjectID `bson:"members" json:"members"`
	CreatedAt time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time            `bson:"updated_at" json:"updated_at"`
}

// ValidateName validates room name
func (r *Room) ValidateName() error {
	if len(r.Name) == 0 {
		return ErrInvalidInput
	}
	if len(r.Name) > 50 {
		return ErrInvalidInput
	}
	return nil
}

// ValidateCode validates room code format and length
func (r *Room) ValidateCode() error {
	if len(r.Code) == 0 {
		return ErrInvalidInput
	}
	if len(r.Code) > 50 {
		return ErrInvalidInput
	}

	// Check if code is alphanumeric (including underscore and hyphen)
	alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !alphanumericRegex.MatchString(r.Code) {
		return ErrInvalidInput
	}

	return nil
}

// Validate validates the entire room
func (r *Room) Validate() error {
	if err := r.ValidateName(); err != nil {
		return err
	}
	if err := r.ValidateCode(); err != nil {
		return err
	}
	return nil
}
