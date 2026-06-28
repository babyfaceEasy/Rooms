package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Item represents a stored entity that has metadata persisted in MongoDB and
// an optional binary object stored in S3/MinIO.
type Item struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	FileKey     string             `json:"file_key,omitempty" bson:"file_key,omitempty"`
	FileName    string             `json:"file_name,omitempty" bson:"file_name,omitempty"`
	ContentType string             `json:"content_type,omitempty" bson:"content_type,omitempty"`
	Size        int64              `json:"size,omitempty" bson:"size,omitempty"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}
