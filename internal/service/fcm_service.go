package service

import (
	"context"
	"fmt"
	"time"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"firebase.google.com/go/v4/messaging"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FCMService handles sending notifications via Firebase Cloud Messaging
type FCMService interface {
	// SendNotification sends a notification to user's devices
	SendNotification(ctx context.Context, userID primitive.ObjectID, notification *domain.Notification) error

	// SendToDevices sends a notification to specific device tokens
	SendToDevices(ctx context.Context, tokens []string, notification *domain.Notification) error

	// IsEnabled checks if FCM is configured and enabled
	IsEnabled() bool
}

type fcmService struct {
	notificationRepo repository.NotificationRepository
	deviceTokenRepo  repository.DeviceTokenRepository
	messagingClient  interface{} // *messaging.Client from Firebase Admin SDK
	enabled          bool
}

// NewFCMService creates a new FCM service
func NewFCMService(
	notificationRepo repository.NotificationRepository,
	deviceTokenRepo repository.DeviceTokenRepository,
	messagingClient interface{},
) FCMService {
	enabled := messagingClient != nil
	return &fcmService{
		notificationRepo: notificationRepo,
		deviceTokenRepo:  deviceTokenRepo,
		messagingClient:  messagingClient,
		enabled:          enabled,
	}
}

// SendNotification sends a notification to all active devices of a user
func (s *fcmService) SendNotification(ctx context.Context, userID primitive.ObjectID, notification *domain.Notification) error {
	if !s.IsEnabled() {
		return nil
	}

	// Get all active device tokens for the user
	activeTokens, err := s.deviceTokenRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get active tokens: %w", err)
	}

	if len(activeTokens) == 0 {
		return nil // No devices to send to
	}

	// Extract token strings
	tokens := make([]string, len(activeTokens))
	for i, token := range activeTokens {
		tokens[i] = token.Token
	}

	return s.SendToDevices(ctx, tokens, notification)
}

// SendToDevices sends a notification to specific device tokens
func (s *fcmService) SendToDevices(ctx context.Context, tokens []string, notification *domain.Notification) error {
	if !s.IsEnabled() || len(tokens) == 0 {
		return nil
	}

	if s.messagingClient == nil {
		return fmt.Errorf("firebase messaging client not initialized")
	}

	// Cast to Firebase messaging client
	messagingClient, ok := s.messagingClient.(*messaging.Client)
	if !ok {
		return fmt.Errorf("invalid messaging client type")
	}

	// Build the notification message
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: s.extractTitle(notification),
			Body:  notification.Message,
		},
		Data: map[string]string{
			"notification_id":  notification.ID.Hex(),
			"type":             string(notification.NotificationType),
			"patch_id":         notification.PatchID.Hex(),
			"created_at":       notification.CreatedAt.Format(time.RFC3339),
		},
	}

	// Track which tokens succeeded and which failed
	var lastErr error
	var fcmMessageID string
	successCount := 0

	// Send to each token individually (multicast not ideal for tracking individual failures)
	for _, token := range tokens {
		msg := *message
		msg.Token = token

		resp, err := messagingClient.Send(ctx, &msg)
		if err != nil {
			lastErr = err
			// Mark device as inactive if token is invalid
			if isInvalidTokenError(err) {
				_ = s.deviceTokenRepo.MarkInactive(ctx, primitive.ObjectID{}) // TODO: pass actual token ID
			}
			continue
		}

		successCount++
		if fcmMessageID == "" {
			fcmMessageID = resp
		}
	}

	// Update notification status in database
	now := time.Now()

	if successCount > 0 {
		// At least some tokens succeeded
		if err := s.notificationRepo.UpdateStatus(ctx, notification.ID, domain.NotificationStatusSent, now); err != nil {
			return fmt.Errorf("failed to update notification status: %w", err)
		}

		if fcmMessageID != "" {
			_ = s.notificationRepo.UpdateFCMMessageID(ctx, notification.ID, fcmMessageID)
		}

		return nil
	}

	// All tokens failed
	if lastErr != nil {
		errMsg := fmt.Sprintf("FCM send failed: %v", lastErr)
		if err := s.notificationRepo.UpdateStatusWithError(ctx, notification.ID, domain.NotificationStatusFailed, errMsg); err != nil {
			return fmt.Errorf("failed to update notification error status: %w", err)
		}
		return lastErr
	}

	return nil
}

// IsEnabled checks if FCM is configured
func (s *fcmService) IsEnabled() bool {
	return s.enabled && s.messagingClient != nil
}

// extractTitle creates a notification title from the notification type
func (s *fcmService) extractTitle(notif *domain.Notification) string {
	switch notif.NotificationType {
	case domain.NotificationTypeAdded:
		return "Added to Patch"
	case domain.NotificationTypeJoined:
		return "Someone Joined"
	case domain.NotificationTypeActivity:
		return "Patch Activity"
	default:
		return "Notification"
	}
}

// isInvalidTokenError checks if an error is due to an invalid device token
func isInvalidTokenError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Firebase returns errors like "registration token is invalid"
	return errStr == "registration token is invalid" ||
		errStr == "mismatched credential" ||
		errStr == "third party auth error"
}
