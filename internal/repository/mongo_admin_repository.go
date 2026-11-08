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

// MongoAdminRepository implements AdminRepository for MongoDB.
type MongoAdminRepository struct {
	collection *mongo.Collection
}

// NewMongoAdminRepository creates a new MongoAdminRepository and ensures indexes.
func NewMongoAdminRepository(db *mongo.Database) (*MongoAdminRepository, error) {
	collection := db.Collection("admins")

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

	// Index for deleted_at (soft deletes)
	deletedAtIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "deleted_at", Value: 1},
		},
	}

	_, err = collection.Indexes().CreateOne(context.Background(), deletedAtIndexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create deleted_at index: %w", err)
	}

	return &MongoAdminRepository{collection: collection}, nil
}

// Create inserts a new admin into the database.
func (r *MongoAdminRepository) Create(ctx context.Context, admin *domain.Admin) error {
	admin.ID = primitive.NewObjectID()
	admin.CreatedAt = time.Now().UTC()
	admin.UpdatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, admin)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrAdminAlreadyExists
		}
		return fmt.Errorf("failed to create admin: %w", err)
	}

	return nil
}

// GetByID retrieves an admin by ID (only active/not deleted)
func (r *MongoAdminRepository) GetByID(ctx context.Context, adminID primitive.ObjectID) (*domain.Admin, error) {
	filter := bson.M{
		"_id":        adminID,
		"deleted_at": nil,
	}

	var admin domain.Admin
	err := r.collection.FindOne(ctx, filter).Decode(&admin)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrAdminNotFound
		}
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}

	return &admin, nil
}

// GetByEmail retrieves an admin by email (only active/not deleted)
func (r *MongoAdminRepository) GetByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	filter := bson.M{
		"email":      email,
		"deleted_at": nil,
	}

	var admin domain.Admin
	err := r.collection.FindOne(ctx, filter).Decode(&admin)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrAdminNotFound
		}
		return nil, fmt.Errorf("failed to get admin by email: %w", err)
	}

	return &admin, nil
}

// List retrieves all active admins (paginated)
func (r *MongoAdminRepository) List(ctx context.Context, limit int, offset int) ([]*domain.Admin, error) {
	filter := bson.M{
		"deleted_at": nil,
	}

	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list admins: %w", err)
	}
	defer cursor.Close(ctx)

	var admins []*domain.Admin
	if err := cursor.All(ctx, &admins); err != nil {
		return nil, fmt.Errorf("failed to decode admins: %w", err)
	}

	if admins == nil {
		admins = make([]*domain.Admin, 0)
	}

	return admins, nil
}

// ListWithFilters retrieves admins with optional filters (paginated)
func (r *MongoAdminRepository) ListWithFilters(ctx context.Context, filters map[string]interface{}, limit int, offset int) ([]*domain.Admin, error) {
	filter := bson.M{
		"deleted_at": nil,
	}

	// Add optional filters
	if role, ok := filters["role"]; ok {
		filter["role"] = role
	}
	if isActive, ok := filters["is_active"]; ok {
		filter["is_active"] = isActive
	}

	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list admins with filters: %w", err)
	}
	defer cursor.Close(ctx)

	var admins []*domain.Admin
	if err := cursor.All(ctx, &admins); err != nil {
		return nil, fmt.Errorf("failed to decode admins: %w", err)
	}

	if admins == nil {
		admins = make([]*domain.Admin, 0)
	}

	return admins, nil
}

// Update updates an existing admin
func (r *MongoAdminRepository) Update(ctx context.Context, admin *domain.Admin) error {
	admin.UpdatedAt = time.Now().UTC()

	filter := bson.M{
		"_id":        admin.ID,
		"deleted_at": nil,
	}

	update := bson.M{
		"$set": admin,
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update admin: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrAdminNotFound
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *MongoAdminRepository) UpdateLastLogin(ctx context.Context, adminID primitive.ObjectID) error {
	filter := bson.M{
		"_id":        adminID,
		"deleted_at": nil,
	}

	update := bson.M{
		"$set": bson.M{
			"last_login": time.Now().UTC(),
			"updated_at": time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrAdminNotFound
	}

	return nil
}

// Delete performs a soft delete of an admin
func (r *MongoAdminRepository) Delete(ctx context.Context, adminID primitive.ObjectID) error {
	filter := bson.M{
		"_id":        adminID,
		"deleted_at": nil,
	}

	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now().UTC(),
			"updated_at": time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrAdminNotFound
	}

	return nil
}

// SetupToken sets a setup token for initial password setup
func (r *MongoAdminRepository) SetupToken(ctx context.Context, adminID primitive.ObjectID, token string, expiresAt interface{}) error {
	filter := bson.M{
		"_id":        adminID,
		"deleted_at": nil,
	}

	update := bson.M{
		"$set": bson.M{
			"setup_token":           token,
			"setup_token_expires_at": expiresAt,
			"requires_password_setup": true,
			"updated_at":            time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to set setup token: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrAdminNotFound
	}

	return nil
}

// VerifySetupToken checks if a setup token is valid
func (r *MongoAdminRepository) VerifySetupToken(ctx context.Context, adminID primitive.ObjectID, token string) (bool, error) {
	filter := bson.M{
		"_id":                       adminID,
		"setup_token":               token,
		"setup_token_expires_at":    bson.M{"$gt": time.Now().UTC()},
		"deleted_at":                nil,
	}

	result := r.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, fmt.Errorf("failed to verify setup token: %w", result.Err())
	}

	return true, nil
}

// ClearSetupToken clears the setup token after successful setup
func (r *MongoAdminRepository) ClearSetupToken(ctx context.Context, adminID primitive.ObjectID) error {
	filter := bson.M{
		"_id":        adminID,
		"deleted_at": nil,
	}

	update := bson.M{
		"$set": bson.M{
			"setup_token":              "",
			"setup_token_expires_at":   nil,
			"requires_password_setup": false,
			"updated_at":              time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to clear setup token: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrAdminNotFound
	}

	return nil
}

// SetPasswordResetToken sets a password reset token
func (r *MongoAdminRepository) SetPasswordResetToken(ctx context.Context, adminID primitive.ObjectID, token string, expiresAt interface{}) error {
	filter := bson.M{
		"_id":        adminID,
		"deleted_at": nil,
	}

	update := bson.M{
		"$set": bson.M{
			"password_reset_token":     token,
			"password_reset_expires_at": expiresAt,
			"updated_at":               time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to set password reset token: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrAdminNotFound
	}

	return nil
}

// VerifyPasswordResetToken checks if a password reset token is valid
func (r *MongoAdminRepository) VerifyPasswordResetToken(ctx context.Context, adminID primitive.ObjectID, token string) (bool, error) {
	filter := bson.M{
		"_id":                        adminID,
		"password_reset_token":       token,
		"password_reset_expires_at":  bson.M{"$gt": time.Now().UTC()},
		"deleted_at":                 nil,
	}

	result := r.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, fmt.Errorf("failed to verify password reset token: %w", result.Err())
	}

	return true, nil
}

// ClearPasswordResetToken clears the password reset token
func (r *MongoAdminRepository) ClearPasswordResetToken(ctx context.Context, adminID primitive.ObjectID) error {
	filter := bson.M{
		"_id":        adminID,
		"deleted_at": nil,
	}

	update := bson.M{
		"$set": bson.M{
			"password_reset_token":     "",
			"password_reset_expires_at": nil,
			"updated_at":               time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to clear password reset token: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrAdminNotFound
	}

	return nil
}

// UpdateFailedLoginAttempts increments failed login attempts and optionally locks the account
func (r *MongoAdminRepository) UpdateFailedLoginAttempts(ctx context.Context, adminID primitive.ObjectID, attempts int, lockedUntil interface{}) error {
	filter := bson.M{
		"_id":        adminID,
		"deleted_at": nil,
	}

	update := bson.M{
		"$set": bson.M{
			"failed_login_attempts": attempts,
			"locked_until":           lockedUntil,
			"updated_at":             time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update failed login attempts: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrAdminNotFound
	}

	return nil
}

// ClearFailedLoginAttempts resets failed login attempts
func (r *MongoAdminRepository) ClearFailedLoginAttempts(ctx context.Context, adminID primitive.ObjectID) error {
	filter := bson.M{
		"_id":        adminID,
		"deleted_at": nil,
	}

	update := bson.M{
		"$set": bson.M{
			"failed_login_attempts": 0,
			"locked_until":          nil,
			"updated_at":            time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to clear failed login attempts: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrAdminNotFound
	}

	return nil
}

// CountSuperAdmins counts the number of active super admin accounts
func (r *MongoAdminRepository) CountSuperAdmins(ctx context.Context) (int, error) {
	filter := bson.M{
		"role":       domain.AdminRoleSuperAdmin,
		"deleted_at": nil,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count super admins: %w", err)
	}

	return int(count), nil
}
