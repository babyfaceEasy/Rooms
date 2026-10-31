package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoUserRepository implements UserRepository for MongoDB.
type MongoUserRepository struct {
	collection *mongo.Collection
}

// NewMongoUserRepository creates a new MongoUserRepository and ensures indexes.
func NewMongoUserRepository(db *mongo.Database) (*MongoUserRepository, error) {
	collection := db.Collection("users")

	// Ensure unique index on email field
	emailIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "email", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context.Background(), emailIndexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create email index: %w", err)
	}

	// Ensure unique index on code field
	codeIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "code", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}

	_, err = collection.Indexes().CreateOne(context.Background(), codeIndexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create code index: %w", err)
	}

	return &MongoUserRepository{collection: collection}, nil
}

// Create inserts a new user into the database.
func (r *MongoUserRepository) Create(ctx context.Context, user *domain.User) error {
	user.ID = primitive.NewObjectID()

	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("email already exists: %w", domain.ErrEmailAlreadyExists)
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// GetByEmail retrieves a user by email address (excluding soft-deleted users).
func (r *MongoUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	filter := bson.M{
		"email":      strings.ToLower(email),
		"deleted_at": bson.M{"$eq": nil},
	}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("user not found: %w", domain.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	return &user, nil
}

// GetByID retrieves a user by MongoDB ObjectID (excluding soft-deleted users).
func (r *MongoUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	var user domain.User
	filter := bson.M{
		"_id":        id,
		"deleted_at": bson.M{"$eq": nil},
	}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("user not found: %w", domain.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	return &user, nil
}

// GetByIDs retrieves multiple users by their IDs.
func (r *MongoUserRepository) GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*domain.User, error) {
	if len(ids) == 0 {
		return []*domain.User{}, nil
	}

	filter := bson.M{
		"_id":        bson.M{"$in": ids},
		"deleted_at": bson.M{"$eq": nil},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

// GetByCode retrieves a user by their customer code (excluding soft-deleted users).
func (r *MongoUserRepository) GetByCode(ctx context.Context, code string) (*domain.User, error) {
	var user domain.User
	filter := bson.M{
		"code":       code,
		"deleted_at": bson.M{"$eq": nil},
	}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("user not found: %w", domain.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	return &user, nil
}

// Update updates an existing user record.
func (r *MongoUserRepository) Update(ctx context.Context, user *domain.User) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{
			"$set": bson.M{
				"name":               user.Name,
				"email":              strings.ToLower(user.Email),
				"password_hash":      user.PasswordHash,
				"is_age_verified":    user.IsAgeVerified,
				"is_email_verified":  user.IsEmailVerified,
				"profile_picture":    user.ProfilePicture,
				"updated_at":         user.UpdatedAt,
			},
		},
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("email already exists: %w", domain.ErrEmailAlreadyExists)
		}
		return fmt.Errorf("failed to update user: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found: %w", domain.ErrUserNotFound)
	}
	return nil
}

// Delete removes a user from the database.
func (r *MongoUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("user not found: %w", domain.ErrUserNotFound)
	}
	return nil
}

// SoftDelete marks a user as deleted without removing the record (soft delete).
func (r *MongoUserRepository) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now().UTC()
	result, err := r.collection.UpdateOne(
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
		return fmt.Errorf("failed to soft delete user: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found: %w", domain.ErrUserNotFound)
	}
	return nil
}

// SuspendUser suspends a user account with a reason
func (r *MongoUserRepository) SuspendUser(ctx context.Context, userID primitive.ObjectID, reason string, suspendedBy primitive.ObjectID) error {
	now := time.Now().UTC()
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID, "deleted_at": nil},
		bson.M{
			"$set": bson.M{
				"is_suspended":      true,
				"suspension_reason": reason,
				"suspended_at":      now,
				"suspended_by":      suspendedBy,
				"updated_at":        now,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to suspend user: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found: %w", domain.ErrUserNotFound)
	}
	return nil
}

// UnsuspendUser unsuspends a user account
func (r *MongoUserRepository) UnsuspendUser(ctx context.Context, userID primitive.ObjectID) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID, "deleted_at": nil},
		bson.M{
			"$set": bson.M{
				"is_suspended":      false,
				"suspension_reason": nil,
				"suspended_at":      nil,
				"suspended_by":      nil,
				"updated_at":        time.Now().UTC(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to unsuspend user: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found: %w", domain.ErrUserNotFound)
	}
	return nil
}

// ListSuspendedUsers retrieves all suspended users (paginated)
func (r *MongoUserRepository) ListSuspendedUsers(ctx context.Context, limit int, offset int) ([]*domain.User, error) {
	filter := bson.M{
		"is_suspended": true,
		"deleted_at":   nil,
	}

	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"suspended_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list suspended users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	if users == nil {
		users = make([]*domain.User, 0)
	}

	return users, nil
}

// ListDeletedUsers retrieves all soft-deleted users (paginated)
func (r *MongoUserRepository) ListDeletedUsers(ctx context.Context, limit int, offset int) ([]*domain.User, error) {
	filter := bson.M{
		"deleted_at": bson.M{"$ne": nil},
	}

	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"deleted_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list deleted users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	if users == nil {
		users = make([]*domain.User, 0)
	}

	return users, nil
}

// CountAllUsers counts all active users
func (r *MongoUserRepository) CountAllUsers(ctx context.Context) (int, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"deleted_at": nil})
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return int(count), nil
}

// CountDeletedUsers counts all soft-deleted users
func (r *MongoUserRepository) CountDeletedUsers(ctx context.Context) (int, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}})
	if err != nil {
		return 0, fmt.Errorf("failed to count deleted users: %w", err)
	}
	return int(count), nil
}

// CountSuspendedUsers counts all suspended users
func (r *MongoUserRepository) CountSuspendedUsers(ctx context.Context) (int, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"is_suspended": true,
		"deleted_at":   nil,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count suspended users: %w", err)
	}
	return int(count), nil
}

// CountVerifiedUsers counts all email-verified users
func (r *MongoUserRepository) CountVerifiedUsers(ctx context.Context) (int, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"is_email_verified": true,
		"deleted_at":        nil,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count verified users: %w", err)
	}
	return int(count), nil
}

// ListAllUsers retrieves all active users (paginated)
func (r *MongoUserRepository) ListAllUsers(ctx context.Context, limit int, offset int) ([]*domain.User, error) {
	filter := bson.M{
		"deleted_at": nil,
	}

	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list all users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	if users == nil {
		users = make([]*domain.User, 0)
	}

	return users, nil
}
