package repository

import (
	"context"
	"time"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationRepository defines the interface for notification persistence
type NotificationRepository interface {
	// CreateNotification creates a new notification
	CreateNotification(ctx context.Context, notification *domain.Notification) error

	// GetByUserID retrieves all notifications for a user
	GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Notification, error)

	// GetByID retrieves a notification by ID
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Notification, error)

	// MarkAsRead marks a notification as read
	MarkAsRead(ctx context.Context, notificationID primitive.ObjectID) error

	// UpdateStatus updates the delivery status of a notification
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status domain.NotificationStatus, sentAt time.Time) error

	// UpdateStatusWithError updates status and stores error message
	UpdateStatusWithError(ctx context.Context, id primitive.ObjectID, status domain.NotificationStatus, errorMsg string) error

	// UpdateFCMMessageID stores the FCM message ID for tracking
	UpdateFCMMessageID(ctx context.Context, id primitive.ObjectID, fcmMessageID string) error

	// GetByStatus retrieves notifications with a specific status
	GetByStatus(ctx context.Context, status domain.NotificationStatus) ([]*domain.Notification, error)

	// GetFailedNotifications retrieves all failed notifications for retry
	GetFailedNotifications(ctx context.Context, limit int) ([]*domain.Notification, error)

	// GetPendingNotifications retrieves pending notifications ready to send
	GetPendingNotifications(ctx context.Context, limit int) ([]*domain.Notification, error)

	// DeleteOlderThan removes notifications older than a specified time
	DeleteOlderThan(ctx context.Context, before time.Time) error
}
