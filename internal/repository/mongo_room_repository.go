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

// GetByID retrieves a room by its ID
func (m *MongoRoomRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Room, error) {
	var room domain.Room
	err := m.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&room)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrRoomNotFound
		}
		return nil, err
	}
	return &room, nil
}

// GetByCode retrieves a room by its code
func (m *MongoRoomRepository) GetByCode(ctx context.Context, code string) (*domain.Room, error) {
	var room domain.Room
	err := m.collection.FindOne(ctx, bson.M{"code": code}).Decode(&room)
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
