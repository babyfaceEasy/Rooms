package repository

import (
	"context"
	"errors"
	"time"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoReportRepository struct {
	collection *mongo.Collection
}

// NewMongoReportRepository creates a new MongoDB report repository
func NewMongoReportRepository(db *mongo.Database) ReportRepository {
	return &mongoReportRepository{
		collection: db.Collection("reports"),
	}
}

// CreateReport creates a new report in the database
func (r *mongoReportRepository) CreateReport(ctx context.Context, report *domain.Report) error {
	if report.ID.IsZero() {
		report.ID = primitive.NewObjectID()
	}
	report.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, report)
	if err != nil {
		return err
	}
	return nil
}

// GetByPostIDAndUserID retrieves a report by post ID and user ID (to check for duplicates)
func (r *mongoReportRepository) GetByPostIDAndUserID(ctx context.Context, postID, userID primitive.ObjectID) (*domain.Report, error) {
	filter := bson.M{
		"post_id": postID,
		"user_id": userID,
	}

	result := r.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, result.Err()
	}

	var report domain.Report
	if err := result.Decode(&report); err != nil {
		return nil, err
	}

	return &report, nil
}

// CountByPostID counts the number of reports for a given post
func (r *mongoReportRepository) CountByPostID(ctx context.Context, postID primitive.ObjectID) (int, error) {
	filter := bson.M{"post_id": postID}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// GetUserReportCountToday retrieves the number of reports a user has created today
func (r *mongoReportRepository) GetUserReportCountToday(ctx context.Context, userID primitive.ObjectID) (int, error) {
	// Calculate the start of today in UTC
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	filter := bson.M{
		"user_id": userID,
		"created_at": bson.M{
			"$gte": startOfDay,
		},
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}
