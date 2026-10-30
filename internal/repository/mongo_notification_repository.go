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

type mongoNotificationRepository struct {
	collection *mongo.Collection
}

// NewMongoNotificationRepository creates a new MongoDB notification repository
func NewMongoNotificationRepository(db *mongo.Database) NotificationRepository {
	return &mongoNotificationRepository{
		collection: db.Collection("notifications"),
	}
}

// CreateNotification creates a new notification in the database
func (n *mongoNotificationRepository) CreateNotification(ctx context.Context, notification *domain.Notification) error {
	if notification.ID.IsZero() {
		notification.ID = primitive.NewObjectID()
	}
	notification.CreatedAt = time.Now()
	notification.IsRead = false
	if notification.Status == "" {
		notification.Status = domain.NotificationStatusPending
	}

	_, err := n.collection.InsertOne(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

// GetByID retrieves a notification by ID
func (n *mongoNotificationRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Notification, error) {
	filter := bson.M{"_id": id}
	result := n.collection.FindOne(ctx, filter)

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get notification: %w", result.Err())
	}

	notification := &domain.Notification{}
	if err := result.Decode(notification); err != nil {
		return nil, fmt.Errorf("failed to decode notification: %w", err)
	}

	return notification, nil
}

// GetByUserID retrieves all notifications for a given user
func (n *mongoNotificationRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Notification, error) {
	filter := bson.M{"user_id": userID}

	cursor, err := n.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer cursor.Close(ctx)

	var notifications []*domain.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, fmt.Errorf("failed to decode notifications: %w", err)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (n *mongoNotificationRepository) MarkAsRead(ctx context.Context, notificationID primitive.ObjectID) error {
	filter := bson.M{"_id": notificationID}
	update := bson.M{
		"$set": bson.M{
			"is_read": true,
		},
	}

	result, err := n.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// UpdateStatus updates the delivery status of a notification
func (n *mongoNotificationRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status domain.NotificationStatus, sentAt time.Time) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":  status,
			"sent_at": sentAt,
		},
	}

	result, err := n.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// UpdateStatusWithError updates status and stores error message
func (n *mongoNotificationRepository) UpdateStatusWithError(ctx context.Context, id primitive.ObjectID, status domain.NotificationStatus, errorMsg string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":         status,
			"error_message":  errorMsg,
		},
	}

	result, err := n.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update notification with error: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// UpdateFCMMessageID stores the FCM message ID for tracking
func (n *mongoNotificationRepository) UpdateFCMMessageID(ctx context.Context, id primitive.ObjectID, fcmMessageID string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"fcm_message_id": fcmMessageID,
		},
	}

	result, err := n.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update FCM message ID: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetByStatus retrieves notifications with a specific status
func (n *mongoNotificationRepository) GetByStatus(ctx context.Context, status domain.NotificationStatus) ([]*domain.Notification, error) {
	filter := bson.M{"status": status}

	cursor, err := n.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications by status: %w", err)
	}
	defer cursor.Close(ctx)

	var notifications []*domain.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, fmt.Errorf("failed to decode notifications: %w", err)
	}

	return notifications, nil
}

// GetFailedNotifications retrieves all failed notifications for retry
func (n *mongoNotificationRepository) GetFailedNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	filter := bson.M{"status": domain.NotificationStatusFailed}
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := n.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query failed notifications: %w", err)
	}
	defer cursor.Close(ctx)

	var notifications []*domain.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, fmt.Errorf("failed to decode notifications: %w", err)
	}

	return notifications, nil
}

// GetPendingNotifications retrieves pending notifications ready to send
func (n *mongoNotificationRepository) GetPendingNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	filter := bson.M{"status": domain.NotificationStatusPending}
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := n.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending notifications: %w", err)
	}
	defer cursor.Close(ctx)

	var notifications []*domain.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, fmt.Errorf("failed to decode notifications: %w", err)
	}

	return notifications, nil
}

// DeleteOlderThan removes notifications older than a specified time
func (n *mongoNotificationRepository) DeleteOlderThan(ctx context.Context, before time.Time) error {
	filter := bson.M{
		"created_at": bson.M{
			"$lt": before,
		},
	}

	_, err := n.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete old notifications: %w", err)
	}

	return nil
}

