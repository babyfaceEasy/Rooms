package repository

import (
	"context"
	"time"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoCommentRepository implements CommentRepository using MongoDB
type MongoCommentRepository struct {
	collection *mongo.Collection
}

// NewMongoCommentRepository creates a new MongoCommentRepository
func NewMongoCommentRepository(collection *mongo.Collection) CommentRepository {
	return &MongoCommentRepository{
		collection: collection,
	}
}

// Create inserts a new comment
func (m *MongoCommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	result, err := m.collection.InsertOne(ctx, comment)
	if err != nil {
		return err
	}

	comment.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves a comment by ID (excluding soft-deleted)
func (m *MongoCommentRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Comment, error) {
	var comment domain.Comment

	err := m.collection.FindOne(ctx, bson.M{
		"_id":        id,
		"deleted_at": nil,
	}).Decode(&comment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrCommentNotFound
		}
		return nil, err
	}

	return &comment, nil
}

// GetByPostID retrieves paginated comments for a post (excluding soft-deleted).
// sortOrder accepts "asc" (oldest first) or "desc" (newest first, default).
// Returns the comments, total count matching the filter, and any error.
func (m *MongoCommentRepository) GetByPostID(ctx context.Context, postID primitive.ObjectID, page, limit int, sortOrder string) ([]*domain.Comment, int64, error) {
	filter := bson.M{
		"post_id":    postID,
		"deleted_at": nil,
	}

	total, err := m.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	sort := -1 // default: newest first
	if sortOrder == "asc" {
		sort = 1
	}

	skip := (page - 1) * limit
	opts := options.Find().
		SetSort(bson.M{"created_at": sort}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := m.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var comments []*domain.Comment
	if err = cursor.All(ctx, &comments); err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// DeleteComment soft-deletes a comment by ID
func (m *MongoCommentRepository) DeleteComment(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now().UTC()
	result, err := m.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"deleted_at": &now}},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrCommentNotFound
	}

	return nil
}
