package service

import (
	"context"
	"time"

	"temp_backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActivityCooldownService manages the 30-minute cooldown logic for activity notifications
type ActivityCooldownService interface {
	// ShouldSendActivityNotification checks if we should send an activity notification
	// Returns true if this is the first activity or if cooldown period has passed
	ShouldSendActivityNotification(ctx context.Context, userID, patchID primitive.ObjectID) (bool, error)

	// RecordActivity updates the activity timestamp for a patch
	RecordActivity(ctx context.Context, userID, patchID primitive.ObjectID) error

	// RecordNotificationSent updates when we sent a notification for activity
	RecordNotificationSent(ctx context.Context, userID, patchID primitive.ObjectID) error

	// RecordPatchViewed updates when user viewed the patch (resets cooldown)
	RecordPatchViewed(ctx context.Context, userID, patchID primitive.ObjectID) error
}

type activityCooldownService struct {
	cooldownRepo     repository.ActivityCooldownRepository
	notificationRepo repository.NotificationRepository
	cooldownMinutes  int
}

// NewActivityCooldownService creates a new activity cooldown service
func NewActivityCooldownService(
	cooldownRepo repository.ActivityCooldownRepository,
	notificationRepo repository.NotificationRepository,
	cooldownMinutes int,
) ActivityCooldownService {
	if cooldownMinutes <= 0 {
		cooldownMinutes = 30 // Default to 30 minutes
	}
	return &activityCooldownService{
		cooldownRepo:     cooldownRepo,
		notificationRepo: notificationRepo,
		cooldownMinutes:  cooldownMinutes,
	}
}

// ShouldSendActivityNotification determines if we should send an activity notification
// based on the 30-minute cooldown window
func (s *activityCooldownService) ShouldSendActivityNotification(ctx context.Context, userID, patchID primitive.ObjectID) (bool, error) {
	return s.cooldownRepo.ShouldSendNotification(ctx, userID, patchID, s.cooldownMinutes)
}

// RecordActivity updates the activity timestamp
func (s *activityCooldownService) RecordActivity(ctx context.Context, userID, patchID primitive.ObjectID) error {
	now := time.Now()
	return s.cooldownRepo.UpdateLastActivityAt(ctx, userID, patchID, now)
}

// RecordNotificationSent updates when we sent a notification
func (s *activityCooldownService) RecordNotificationSent(ctx context.Context, userID, patchID primitive.ObjectID) error {
	now := time.Now()
	return s.cooldownRepo.UpdateLastNotificationSentAt(ctx, userID, patchID, now)
}

// RecordPatchViewed updates when user viewed the patch
// This potentially resets the cooldown for future notifications
func (s *activityCooldownService) RecordPatchViewed(ctx context.Context, userID, patchID primitive.ObjectID) error {
	now := time.Now()
	return s.cooldownRepo.UpdateLastCheckedAt(ctx, userID, patchID, now)
}
