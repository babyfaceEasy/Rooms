package repository

import (
	"context"
	"fmt"
	"time"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoAdminLogRepository implements AdminLogRepository for MongoDB.
type MongoAdminLogRepository struct {
	collection *mongo.Collection
}

// NewMongoAdminLogRepository creates a new MongoAdminLogRepository and ensures indexes.
func NewMongoAdminLogRepository(db *mongo.Database) (*MongoAdminLogRepository, error) {
	collection := db.Collection("admin_logs")

	// Index for admin_id + created_at (for quickly finding admin's logs)
	adminCreatedIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "admin_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
	}

	_, err := collection.Indexes().CreateOne(context.Background(), adminCreatedIndexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin_id + created_at index: %w", err)
	}

	// Index for action + created_at (for finding logs by action)
	actionCreatedIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "action", Value: 1},
			{Key: "created_at", Value: -1},
		},
	}

	_, err = collection.Indexes().CreateOne(context.Background(), actionCreatedIndexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create action + created_at index: %w", err)
	}

	// Index for resource_type + resource_id (for finding logs by resource)
	resourceIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "resource_type", Value: 1},
			{Key: "resource_id", Value: 1},
		},
	}

	_, err = collection.Indexes().CreateOne(context.Background(), resourceIndexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource index: %w", err)
	}

	// TTL index: automatically delete logs older than 90 days
	ttlIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "created_at", Value: 1},
		},
		Options: options.Index().SetExpireAfterSeconds(90 * 24 * 60 * 60),
	}

	_, err = collection.Indexes().CreateOne(context.Background(), ttlIndexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create TTL index: %w", err)
	}

	return &MongoAdminLogRepository{collection: collection}, nil
}

// Create creates a new admin log entry
func (r *MongoAdminLogRepository) Create(ctx context.Context, log *domain.AdminLog) error {
	log.ID = primitive.NewObjectID()
	log.CreatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, log)
	if err != nil {
		return fmt.Errorf("failed to create admin log: %w", err)
	}

	return nil
}

// List retrieves logs (paginated)
func (r *MongoAdminLogRepository) List(ctx context.Context, limit int, offset int) ([]*domain.AdminLog, error) {
	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list logs: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []*domain.AdminLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %w", err)
	}

	if logs == nil {
		logs = make([]*domain.AdminLog, 0)
	}

	return logs, nil
}

// ListByAdmin retrieves logs for a specific admin (paginated)
func (r *MongoAdminLogRepository) ListByAdmin(ctx context.Context, adminID primitive.ObjectID, limit int, offset int) ([]*domain.AdminLog, error) {
	filter := bson.M{
		"admin_id": adminID,
	}

	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list logs by admin: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []*domain.AdminLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %w", err)
	}

	if logs == nil {
		logs = make([]*domain.AdminLog, 0)
	}

	return logs, nil
}

// ListByResource retrieves logs for a specific resource (paginated)
func (r *MongoAdminLogRepository) ListByResource(ctx context.Context, resourceType domain.AdminLogResourceType, resourceID primitive.ObjectID, limit int, offset int) ([]*domain.AdminLog, error) {
	filter := bson.M{
		"resource_type": resourceType,
		"resource_id":   resourceID,
	}

	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list logs by resource: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []*domain.AdminLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %w", err)
	}

	if logs == nil {
		logs = make([]*domain.AdminLog, 0)
	}

	return logs, nil
}

// ListByAction retrieves logs for a specific action (paginated)
func (r *MongoAdminLogRepository) ListByAction(ctx context.Context, action domain.AdminLogAction, limit int, offset int) ([]*domain.AdminLog, error) {
	filter := bson.M{
		"action": action,
	}

	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list logs by action: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []*domain.AdminLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %w", err)
	}

	if logs == nil {
		logs = make([]*domain.AdminLog, 0)
	}

	return logs, nil
}

// ListWithFilters retrieves logs with multiple filters (paginated)
func (r *MongoAdminLogRepository) ListWithFilters(ctx context.Context, filters map[string]interface{}, limit int, offset int) ([]*domain.AdminLog, error) {
	filter := bson.M{}

	// Add optional filters
	if adminID, ok := filters["admin_id"]; ok {
		filter["admin_id"] = adminID
	}
	if action, ok := filters["action"]; ok {
		filter["action"] = action
	}
	if resourceType, ok := filters["resource_type"]; ok {
		filter["resource_type"] = resourceType
	}
	if resourceID, ok := filters["resource_id"]; ok {
		filter["resource_id"] = resourceID
	}
	if status, ok := filters["status"]; ok {
		filter["status"] = status
	}

	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list logs with filters: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []*domain.AdminLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %w", err)
	}

	if logs == nil {
		logs = make([]*domain.AdminLog, 0)
	}

	return logs, nil
}
