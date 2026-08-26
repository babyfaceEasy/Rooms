package repository

import (
	"context"

	"temp_backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportRepository defines the interface for report persistence
type ReportRepository interface {
	CreateReport(ctx context.Context, report *domain.Report) error
	GetByPostIDAndUserID(ctx context.Context, postID, userID primitive.ObjectID) (*domain.Report, error)
	CountByPostID(ctx context.Context, postID primitive.ObjectID) (int, error)
	GetUserReportCountToday(ctx context.Context, userID primitive.ObjectID) (int, error)
}
