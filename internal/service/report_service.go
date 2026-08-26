package service

import (
	"context"
	"fmt"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportService defines the interface for report business logic
type ReportService interface {
	ReportPost(ctx context.Context, postID, userID primitive.ObjectID, reason string, comment *string) error
}

type reportService struct {
	reportRepo          repository.ReportRepository
	notificationRepo    repository.NotificationRepository
	postRepo            repository.PostRepository
	autoDeleteThreshold int
	maxReportsPerDay    int
}

// NewReportService creates a new report service
func NewReportService(
	reportRepo repository.ReportRepository,
	notificationRepo repository.NotificationRepository,
	postRepo repository.PostRepository,
	autoDeleteThreshold int,
	maxReportsPerDay int,
) ReportService {
	return &reportService{
		reportRepo:          reportRepo,
		notificationRepo:    notificationRepo,
		postRepo:            postRepo,
		autoDeleteThreshold: autoDeleteThreshold,
		maxReportsPerDay:    maxReportsPerDay,
	}
}

// ReportPost reports a post with the specified reason and optional comment
func (s *reportService) ReportPost(ctx context.Context, postID, userID primitive.ObjectID, reason string, comment *string) error {
	// Validate report reason
	if !domain.IsValidReason(reason) {
		return domain.ErrInvalidReportReason
	}

	// Get the post to verify it exists and get the creator
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	// Check if user is trying to report their own post
	if post.UserID == userID {
		return domain.ErrCannotReportOwnPost
	}

	// Check if user already reported this post
	_, err = s.reportRepo.GetByPostIDAndUserID(ctx, postID, userID)
	if err == nil {
		// Found an existing report
		return domain.ErrReportAlreadyExists
	}
	if err != domain.ErrNotFound {
		// Some other error occurred
		return err
	}

	// Check if user exceeded daily report limit
	todayCount, err := s.reportRepo.GetUserReportCountToday(ctx, userID)
	if err != nil {
		return err
	}
	if todayCount >= s.maxReportsPerDay {
		return domain.ErrReportLimitExceeded
	}

	// Create the report
	report := &domain.Report{
		PostID:  postID,
		UserID:  userID,
		Reason:  domain.ReportReason(reason),
		Comment: comment,
	}

	if err := s.reportRepo.CreateReport(ctx, report); err != nil {
		return err
	}

	// Create notification for the post creator
	notification := &domain.Notification{
		UserID:  post.UserID,
		PostID:  postID,
		Reason:  domain.ReportReason(reason),
		Message: fmt.Sprintf("Your post has been reported for: %s", reason),
	}

	if err := s.notificationRepo.CreateNotification(ctx, notification); err != nil {
		return err
	}

	// Increment report count
	newCount, err := s.postRepo.IncrementReportCount(ctx, postID)
	if err != nil {
		return err
	}

	// If report count reaches threshold, soft delete the post
	if newCount >= s.autoDeleteThreshold {
		if err := s.postRepo.DeletePost(ctx, postID); err != nil {
			return err
		}
	}

	return nil
}
