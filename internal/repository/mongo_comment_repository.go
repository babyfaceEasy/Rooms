package repository

import (
	"context"
	"time"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

// GetByPostID retrieves all comments for a post (excluding soft-deleted)
func (m *MongoCommentRepository) GetByPostID(ctx context.Context, postID primitive.ObjectID) ([]*domain.Comment, error) {
	filter := bson.M{
		"post_id":    postID,
		"deleted_at": nil,
	}

	cursor, err := m.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []*domain.Comment
	if err = cursor.All(ctx, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// DeleteComment soft-deletes a comment by ID
func (m *MongoCommentRepository) DeleteComment(ctx context.Context, id primitive.ObjectID) error {
	result, err := m.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"deleted_at": primitive.NewDateTimeFromTime(ctx.Value("now").(time.Time))}},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrCommentNotFound
	}

	return nil
}
