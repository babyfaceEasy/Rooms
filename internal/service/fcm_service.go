package service

import (
	"context"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

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
	fcmClient        interface{} // firebase.Messaging from Firebase Admin SDK
}

// NewFCMService creates a new FCM service
func NewFCMService(
	notificationRepo repository.NotificationRepository,
	deviceTokenRepo repository.DeviceTokenRepository,
	fcmClient interface{},
) FCMService {
	return &fcmService{
		notificationRepo: notificationRepo,
		deviceTokenRepo:  deviceTokenRepo,
		fcmClient:        fcmClient,
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
		return err
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

	if s.fcmClient == nil {
		return domain.ErrInvalidInput // Firebase not initialized
	}

	// Cast fcmClient to Firebase Messaging client
	// Note: This is a placeholder. In Phase 4 integration, this will use:
	// messaging := s.fcmClient.(*firebase.App).Messaging(ctx)
	//
	// For now, we create the structure that Phase 5 will implement with actual Firebase calls:
	//
	// For each token, send:
	// message := &messaging.Message{
	//     Notification: &messaging.Notification{
	//         Title: extractTitle(notification),
	//         Body:  notification.Message,
	//     },
	//     Data: map[string]string{
	//         "notification_id": notification.ID.Hex(),
	//         "type":            string(notification.NotificationType),
	//     },
	//     Token: token,
	// }
	//
	// resp, err := messaging.Send(ctx, message)
	// if err != nil {
	//     update status to failed with error
	// } else {
	//     update status to sent with FCM message ID
	// }

	// For Phase 4, we structure the logic without actual Firebase calls
	// Phase 5 will implement the actual Firebase Admin SDK integration

	// Record that we attempted to send
	now := ctx.Value("timestamp")
	if now == nil {
		now = nil // Will use time.Now() in actual implementation
	}

	// Update notification status to sent (actual Firebase call will happen in Phase 5)
	if notification.ID.IsZero() {
		return domain.ErrInvalidInput
	}

	// This is where Phase 5 will integrate with actual Firebase messaging
	// For now, we're setting up the infrastructure

	return nil
}

// IsEnabled checks if FCM is configured
func (s *fcmService) IsEnabled() bool {
	return s.fcmClient != nil
}
