package service

import (
	"context"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminLogService defines admin logging operations
type AdminLogService interface {
	// LogAction logs an admin action
	LogAction(ctx context.Context, adminID primitive.ObjectID, adminEmail string, action domain.AdminLogAction, resourceType domain.AdminLogResourceType, resourceID *primitive.ObjectID, changes map[string]interface{}, ipAddress string, userAgent string) error

	// GetLogs retrieves logs (paginated)
	GetLogs(ctx context.Context, limit int, offset int) ([]*domain.AdminLog, error)

	// GetAdminLogs retrieves logs for a specific admin (paginated)
	GetAdminLogs(ctx context.Context, adminID primitive.ObjectID, limit int, offset int) ([]*domain.AdminLog, error)

	// GetResourceLogs retrieves logs for a specific resource
	GetResourceLogs(ctx context.Context, resourceType domain.AdminLogResourceType, resourceID primitive.ObjectID, limit int, offset int) ([]*domain.AdminLog, error)
}

type adminLogService struct {
	logRepo repository.AdminLogRepository
}

// NewAdminLogService creates a new admin log service
func NewAdminLogService(logRepo repository.AdminLogRepository) AdminLogService {
	return &adminLogService{
		logRepo: logRepo,
	}
}

// LogAction logs an admin action
func (s *adminLogService) LogAction(ctx context.Context, adminID primitive.ObjectID, adminEmail string, action domain.AdminLogAction, resourceType domain.AdminLogResourceType, resourceID *primitive.ObjectID, changes map[string]interface{}, ipAddress string, userAgent string) error {
	log := &domain.AdminLog{
		AdminID:      adminID,
		AdminEmail:   adminEmail,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Changes:      changes,
		Status:       domain.AdminLogStatusSuccess,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	return s.logRepo.Create(ctx, log)
}

// GetLogs retrieves logs (paginated)
func (s *adminLogService) GetLogs(ctx context.Context, limit int, offset int) ([]*domain.AdminLog, error) {
	return s.logRepo.List(ctx, limit, offset)
}

// GetAdminLogs retrieves logs for a specific admin (paginated)
func (s *adminLogService) GetAdminLogs(ctx context.Context, adminID primitive.ObjectID, limit int, offset int) ([]*domain.AdminLog, error) {
	return s.logRepo.ListByAdmin(ctx, adminID, limit, offset)
}

// GetResourceLogs retrieves logs for a specific resource
func (s *adminLogService) GetResourceLogs(ctx context.Context, resourceType domain.AdminLogResourceType, resourceID primitive.ObjectID, limit int, offset int) ([]*domain.AdminLog, error) {
	return s.logRepo.ListByResource(ctx, resourceType, resourceID, limit, offset)
}
