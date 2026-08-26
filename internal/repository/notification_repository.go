package repository

import (
	"context"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationRepository defines the interface for notification persistence
type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification *domain.Notification) error
	GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Notification, error)
	MarkAsRead(ctx context.Context, notificationID primitive.ObjectID) error
}
