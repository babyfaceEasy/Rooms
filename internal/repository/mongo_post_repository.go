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

// MongoPostRepository implements PostRepository for MongoDB
type MongoPostRepository struct {
	collection *mongo.Collection
}

// NewMongoPostRepository creates a new MongoDB post repository
func NewMongoPostRepository(db *mongo.Database) *MongoPostRepository {
	collection := db.Collection("posts")
	return &MongoPostRepository{
		collection: collection,
	}
}

// Create creates a new post in the database
func (m *MongoPostRepository) Create(ctx context.Context, post *domain.Post) error {
	result, err := m.collection.InsertOne(ctx, post)
	if err != nil {
		return err
	}

	// Set the ID to the inserted ID
	post.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves a post by its ID (excluding soft-deleted)
func (m *MongoPostRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Post, error) {
	var post domain.Post
	err := m.collection.FindOne(ctx, bson.M{
		"_id":        id,
		"deleted_at": nil,
	}).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrPostNotFound
		}
		return nil, err
	}
	return &post, nil
}

// DeletePost soft-deletes a post by marking it as deleted
func (m *MongoPostRepository) DeletePost(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	result, err := m.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"deleted_at": now,
				"updated_at": now,
			},
		},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrPostNotFound
	}

	return nil
}

// AddValidation adds a user ID to the post's validations array ($addToSet prevents duplicates).
func (m *MongoPostRepository) AddValidation(ctx context.Context, postID, userID primitive.ObjectID) error {
	result, err := m.collection.UpdateOne(
		ctx,
		bson.M{"_id": postID, "deleted_at": nil},
		bson.M{
			"$addToSet": bson.M{"validations": userID},
			"$set":      bson.M{"updated_at": time.Now()},
		},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrPostNotFound
	}

	return nil
}

// RemoveValidation removes a user ID from the post's validations array.
func (m *MongoPostRepository) RemoveValidation(ctx context.Context, postID, userID primitive.ObjectID) error {
	result, err := m.collection.UpdateOne(
		ctx,
		bson.M{"_id": postID, "deleted_at": nil},
		bson.M{
			"$pull": bson.M{"validations": userID},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrPostNotFound
	}

	return nil
}

// AddRespect adds a user ID to the post's respects array ($addToSet prevents duplicates).
func (m *MongoPostRepository) AddRespect(ctx context.Context, postID, userID primitive.ObjectID) error {
	result, err := m.collection.UpdateOne(
		ctx,
		bson.M{"_id": postID, "deleted_at": nil},
		bson.M{
			"$addToSet": bson.M{"respects": userID},
			"$set":      bson.M{"updated_at": time.Now()},
		},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrPostNotFound
	}

	return nil
}

// RemoveRespect removes a user ID from the post's respects array.
func (m *MongoPostRepository) RemoveRespect(ctx context.Context, postID, userID primitive.ObjectID) error {
	result, err := m.collection.UpdateOne(
		ctx,
		bson.M{"_id": postID, "deleted_at": nil},
		bson.M{
			"$pull": bson.M{"respects": userID},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrPostNotFound
	}

	return nil
}

// GetByRoomID retrieves paginated posts for a room (excluding soft-deleted), newest first.
// Returns the posts, total count matching the filter, and any error.
func (m *MongoPostRepository) GetByRoomID(ctx context.Context, roomID primitive.ObjectID, page, limit int) ([]*domain.Post, int64, error) {
	filter := bson.M{
		"room_id":    roomID,
		"deleted_at": nil,
	}

	total, err := m.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := (page - 1) * limit
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := m.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var posts []*domain.Post
	if err = cursor.All(ctx, &posts); err != nil {
		return nil, 0, err
	}

	if posts == nil {
		posts = []*domain.Post{}
	}

	return posts, total, nil
}
