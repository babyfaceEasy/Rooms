package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/internal/domain"
)

// MongoRoomRepository implements RoomRepository for MongoDB
type MongoRoomRepository struct {
	collection *mongo.Collection
}

// NewMongoRoomRepository creates a new MongoDB room repository
func NewMongoRoomRepository(db *mongo.Database) *MongoRoomRepository {
	collection := db.Collection("rooms")

	// Create unique index on code
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	collection.Indexes().CreateOne(context.Background(), indexModel)

	return &MongoRoomRepository{
		collection: collection,
	}
}

// Create creates a new room in the database
func (m *MongoRoomRepository) Create(ctx context.Context, room *domain.Room) error {
	result, err := m.collection.InsertOne(ctx, room)
	if err != nil {
		// Check if it's a duplicate key error (code already exists)
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrCodeAlreadyExists
		}
		return err
	}

	// Set the ID to the inserted ID
	room.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves a room by its ID (excluding soft-deleted)
func (m *MongoRoomRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Room, error) {
	var room domain.Room
	err := m.collection.FindOne(ctx, bson.M{
		"_id":        id,
		"deleted_at": nil,
	}).Decode(&room)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrRoomNotFound
		}
		return nil, err
	}
	return &room, nil
}

// GetByCode retrieves a room by its code (excluding soft-deleted)
func (m *MongoRoomRepository) GetByCode(ctx context.Context, code string) (*domain.Room, error) {
	var room domain.Room
	err := m.collection.FindOne(ctx, bson.M{
		"code":       code,
		"deleted_at": nil,
	}).Decode(&room)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrRoomNotFound
		}
		return nil, err
	}
	return &room, nil
}

// AddUserToRoom adds a user to the room's members list
func (m *MongoRoomRepository) AddUserToRoom(ctx context.Context, roomID, userID primitive.ObjectID) error {
	result, err := m.collection.UpdateOne(
		ctx,
		bson.M{"_id": roomID},
		bson.M{
			"$addToSet": bson.M{"members": userID},
			"$set":      bson.M{"updated_at": time.Now()},
		},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrRoomNotFound
	}

	return nil
}

// RemoveUserFromRoom removes a user from the room's members list
func (m *MongoRoomRepository) RemoveUserFromRoom(ctx context.Context, roomID, userID primitive.ObjectID) error {
	result, err := m.collection.UpdateOne(
		ctx,
		bson.M{"_id": roomID},
		bson.M{
			"$pull": bson.M{"members": userID},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrRoomNotFound
	}

	return nil
}

// DeleteRoom soft-deletes a room by marking it as deleted
func (m *MongoRoomRepository) DeleteRoom(ctx context.Context, roomID primitive.ObjectID) error {
	now := time.Now()
	result, err := m.collection.UpdateOne(
		ctx,
		bson.M{"_id": roomID},
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
		return domain.ErrRoomNotFound
	}

	return nil
}

// ListUserRooms retrieves all rooms where the user is the creator or a member (excluding soft-deleted)
func (m *MongoRoomRepository) ListUserRooms(ctx context.Context, userID primitive.ObjectID) ([]*domain.Room, error) {
	cursor, err := m.collection.Find(ctx, bson.M{
		"$or": []bson.M{
			{"created_by": userID},
			{"members": userID},
		},
		"deleted_at": nil,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rooms []*domain.Room
	if err = cursor.All(ctx, &rooms); err != nil {
		return nil, err
	}

	// If no rooms found, return empty slice instead of nil
	if rooms == nil {
		rooms = []*domain.Room{}
	}

	return rooms, nil
}

