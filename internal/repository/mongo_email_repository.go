package repository

import (
	"context"
	"fmt"
	"time"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoEmailRepository implements EmailRepository using MongoDB.
type MongoEmailRepository struct {
	collection *mongo.Collection
}

// NewMongoEmailRepository creates and returns a new MongoEmailRepository.
func NewMongoEmailRepository(db *mongo.Database) (*MongoEmailRepository, error) {
	collection := db.Collection("emails")

	// Create indexes if they don't exist
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create user_id index: %w", err)
	}

	return &MongoEmailRepository{collection: collection}, nil
}

// SaveEmailLog creates a new email log entry in the database.
func (r *MongoEmailRepository) SaveEmailLog(ctx context.Context, email *domain.Email) (*domain.Email, error) {
	if email.ID.IsZero() {
		email.ID = primitive.NewObjectID()
	}
	email.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to insert email log: %w", err)
	}

	email.ID = result.InsertedID.(primitive.ObjectID)
	return email, nil
}

// UpdateEmailStatus updates the status and optional error message of a logged email.
func (r *MongoEmailRepository) UpdateEmailStatus(ctx context.Context, emailID string, status string, errorMsg *string) error {
	objectID, err := primitive.ObjectIDFromHex(emailID)
	if err != nil {
		return fmt.Errorf("invalid email ID: %w", domain.ErrInvalidInput)
	}

	updateDoc := bson.M{
		"$set": bson.M{
			"status":        status,
			"error_message": errorMsg,
		},
	}

	// Set sent_at if status is "sent"
	if status == domain.EmailStatusSent {
		now := time.Now()
		updateDoc["$set"].(bson.M)["sent_at"] = now
	}

	result, err := r.collection.UpdateByID(ctx, objectID, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update email status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("email not found: %w", domain.ErrNotFound)
	}

	return nil
}

// GetByID retrieves an email log by ID.
func (r *MongoEmailRepository) GetByID(ctx context.Context, id string) (*domain.Email, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid email ID: %w", domain.ErrInvalidInput)
	}

	var email domain.Email
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get email: %w", err)
	}

	return &email, nil
}

// ListByUserID retrieves all email logs for a specific user.
func (r *MongoEmailRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.Email, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", domain.ErrInvalidInput)
	}

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": objectID})
	if err != nil {
		return nil, fmt.Errorf("failed to list emails: %w", err)
	}
	defer cursor.Close(ctx)

	var emails []*domain.Email
	err = cursor.All(ctx, &emails)
	if err != nil {
		return nil, fmt.Errorf("failed to decode emails: %w", err)
	}

	return emails, nil
}
