package service

import (
	"context"
	"fmt"
	"time"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationOrchestratorService orchestrates sending notifications for the three trigger types:
// 1. User added to a Patch
// 2. User joined a Patch
// 3. Activity in a Patch (with cooldown)
type NotificationOrchestratorService interface {
	// NotifyUserAdded sends notification when user is added to a patch
	NotifyUserAdded(ctx context.Context, userID primitive.ObjectID, addedByName, patchName string, patchID primitive.ObjectID) error

	// NotifyUserJoined sends notification when user joins a patch
	NotifyUserJoined(ctx context.Context, userID primitive.ObjectID, joinedByName, patchName string, patchID primitive.ObjectID) error

	// NotifyPatchActivity sends notification for activity in patch (posts, comments, reactions)
	// with smart 30-minute cooldown
	NotifyPatchActivity(ctx context.Context, userID primitive.ObjectID, patchName string, patchID primitive.ObjectID) error
}

type notificationOrchestratorService struct {
	fcmService           FCMService
	cooldownService      ActivityCooldownService
	deviceTokenRepo      repository.DeviceTokenRepository
	notificationRepo     repository.NotificationRepository
	activityCooldownRepo repository.ActivityCooldownRepository
}

// NewNotificationOrchestratorService creates a new notification orchestrator
func NewNotificationOrchestratorService(
	fcmService FCMService,
	cooldownService ActivityCooldownService,
	deviceTokenRepo repository.DeviceTokenRepository,
	notificationRepo repository.NotificationRepository,
	activityCooldownRepo repository.ActivityCooldownRepository,
) NotificationOrchestratorService {
	return &notificationOrchestratorService{
		fcmService:           fcmService,
		cooldownService:      cooldownService,
		deviceTokenRepo:      deviceTokenRepo,
		notificationRepo:     notificationRepo,
		activityCooldownRepo: activityCooldownRepo,
	}
}

// NotifyUserAdded sends "User added you to Patch" notification
func (s *notificationOrchestratorService) NotifyUserAdded(
	ctx context.Context,
	userID primitive.ObjectID,
	addedByName, patchName string,
	patchID primitive.ObjectID,
) error {
	// Create notification record
	message := fmt.Sprintf("%s added you to %s", addedByName, patchName)
	notification := &domain.Notification{
		ID:               primitive.NewObjectID(),
		UserID:           userID,
		PatchID:          patchID,
		NotificationType: domain.NotificationTypeAdded,
		Status:           domain.NotificationStatusPending,
		Message:          message,
		IsRead:           false,
		CreatedAt:        time.Now(),
	}

	// Save notification to database
	if err := s.notificationRepo.CreateNotification(ctx, notification); err != nil {
		return fmt.Errorf("failed to create added notification: %w", err)
	}

	// Send via FCM to all user's devices
	if err := s.fcmService.SendNotification(ctx, userID, notification); err != nil {
		// Log error but don't fail - notification is saved in DB
		// Update status to failed
		_ = s.notificationRepo.UpdateStatusWithError(
			ctx,
			notification.ID,
			domain.NotificationStatusFailed,
			fmt.Sprintf("FCM send failed: %v", err),
		)
		return fmt.Errorf("failed to send added notification via FCM: %w", err)
	}

	// Mark as sent
	now := time.Now()
	return s.notificationRepo.UpdateStatus(ctx, notification.ID, domain.NotificationStatusSent, now)
}

// NotifyUserJoined sends "User joined Patch" notification
func (s *notificationOrchestratorService) NotifyUserJoined(
	ctx context.Context,
	userID primitive.ObjectID,
	joinedByName, patchName string,
	patchID primitive.ObjectID,
) error {
	// Create notification record
	message := fmt.Sprintf("%s joined %s", joinedByName, patchName)
	notification := &domain.Notification{
		ID:               primitive.NewObjectID(),
		UserID:           userID,
		PatchID:          patchID,
		NotificationType: domain.NotificationTypeJoined,
		Status:           domain.NotificationStatusPending,
		Message:          message,
		IsRead:           false,
		CreatedAt:        time.Now(),
	}

	// Save notification to database
	if err := s.notificationRepo.CreateNotification(ctx, notification); err != nil {
		return fmt.Errorf("failed to create joined notification: %w", err)
	}

	// Send via FCM to all user's devices
	if err := s.fcmService.SendNotification(ctx, userID, notification); err != nil {
		// Log error but don't fail - notification is saved in DB
		// Update status to failed
		_ = s.notificationRepo.UpdateStatusWithError(
			ctx,
			notification.ID,
			domain.NotificationStatusFailed,
			fmt.Sprintf("FCM send failed: %v", err),
		)
		return fmt.Errorf("failed to send joined notification via FCM: %w", err)
	}

	// Mark as sent
	now := time.Now()
	return s.notificationRepo.UpdateStatus(ctx, notification.ID, domain.NotificationStatusSent, now)
}

// NotifyPatchActivity sends "Activity in Patch" notification with smart 30-minute cooldown
func (s *notificationOrchestratorService) NotifyPatchActivity(
	ctx context.Context,
	userID primitive.ObjectID,
	patchName string,
	patchID primitive.ObjectID,
) error {
	// Record the activity
	if err := s.cooldownService.RecordActivity(ctx, userID, patchID); err != nil {
		return fmt.Errorf("failed to record activity: %w", err)
	}

	// Check if we should send a notification (30-minute cooldown)
	shouldSend, err := s.cooldownService.ShouldSendActivityNotification(ctx, userID, patchID)
	if err != nil {
		return fmt.Errorf("failed to check cooldown: %w", err)
	}

	if !shouldSend {
		// Still in cooldown window, skip notification
		return nil
	}

	// Create notification record
	message := fmt.Sprintf("There's activity happening in %s", patchName)
	notification := &domain.Notification{
		ID:               primitive.NewObjectID(),
		UserID:           userID,
		PatchID:          patchID,
		NotificationType: domain.NotificationTypeActivity,
		Status:           domain.NotificationStatusPending,
		Message:          message,
		IsRead:           false,
		CreatedAt:        time.Now(),
	}

	// Save notification to database
	if err := s.notificationRepo.CreateNotification(ctx, notification); err != nil {
		return fmt.Errorf("failed to create activity notification: %w", err)
	}

	// Send via FCM to all user's devices
	if err := s.fcmService.SendNotification(ctx, userID, notification); err != nil {
		// Log error but don't fail - notification is saved in DB
		// Update status to failed
		_ = s.notificationRepo.UpdateStatusWithError(
			ctx,
			notification.ID,
			domain.NotificationStatusFailed,
			fmt.Sprintf("FCM send failed: %v", err),
		)
		return fmt.Errorf("failed to send activity notification via FCM: %w", err)
	}

	// Record that we sent the notification (updates cooldown timer)
	if err := s.cooldownService.RecordNotificationSent(ctx, userID, patchID); err != nil {
		// Log but don't fail
		fmt.Printf("failed to record notification sent time: %v\n", err)
	}

	// Mark as sent
	now := time.Now()
	return s.notificationRepo.UpdateStatus(ctx, notification.ID, domain.NotificationStatusSent, now)
}
