package service

import (
	"context"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationService defines operations for querying notification data.
type NotificationService interface {
	// GetUserNotifications retrieves all notifications for a user.
	GetUserNotifications(ctx context.Context, userID primitive.ObjectID) ([]*domain.Notification, error)

	// GetUserNotificationsByPatch retrieves notifications for a user in a specific patch.
	GetUserNotificationsByPatch(ctx context.Context, userID, patchID primitive.ObjectID) ([]*domain.Notification, error)

	// GetNotificationByID retrieves a notification by its ID.
	GetNotificationByID(ctx context.Context, notificationID primitive.ObjectID) (*domain.Notification, error)
}

// notificationService is the concrete implementation of NotificationService.
type notificationService struct {
	notificationRepo repository.NotificationRepository
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(notificationRepo repository.NotificationRepository) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
	}
}

func (s *notificationService) GetUserNotifications(ctx context.Context, userID primitive.ObjectID) ([]*domain.Notification, error) {
	return s.notificationRepo.GetByUserID(ctx, userID)
}

func (s *notificationService) GetUserNotificationsByPatch(ctx context.Context, userID, patchID primitive.ObjectID) ([]*domain.Notification, error) {
	// Get all user notifications and filter by patch
	notifications, err := s.notificationRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Filter by patch ID
	var filtered []*domain.Notification
	for _, n := range notifications {
		if n.PatchID == patchID {
			filtered = append(filtered, n)
		}
	}

	return filtered, nil
}

func (s *notificationService) GetNotificationByID(ctx context.Context, notificationID primitive.ObjectID) (*domain.Notification, error) {
	return s.notificationRepo.GetByID(ctx, notificationID)
}
