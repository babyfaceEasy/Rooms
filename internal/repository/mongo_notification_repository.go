package repository

import (
	"context"
	"time"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

	_, err := n.collection.InsertOne(ctx, notification)
	if err != nil {
		return err
	}
	return nil
}

// GetByUserID retrieves all notifications for a given user
func (n *mongoNotificationRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Notification, error) {
	filter := bson.M{"user_id": userID}

	cursor, err := n.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*domain.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
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

	_, err := n.collection.UpdateOne(ctx, filter, update)
	return err
}
