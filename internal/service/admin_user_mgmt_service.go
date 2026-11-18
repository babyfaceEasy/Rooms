package service

import (
	"context"
	"fmt"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminUserManagementService defines user management operations for admins
type AdminUserManagementService interface {
	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, userID primitive.ObjectID) (*domain.User, error)

	// ListAllUsers retrieves all users (paginated) with total count
	ListAllUsers(ctx context.Context, limit int, offset int) (*domain.PaginatedResponse[*domain.User], error)

	// SearchUsers searches users by email or name (paginated) with total count
	SearchUsers(ctx context.Context, query string, limit int, offset int) (*domain.PaginatedResponse[*domain.User], error)

	// SuspendUser suspends a user
	SuspendUser(ctx context.Context, userID primitive.ObjectID, reason string, adminID primitive.ObjectID) error

	// UnsuspendUser unsuspends a user
	UnsuspendUser(ctx context.Context, userID primitive.ObjectID, adminID primitive.ObjectID) error

	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, userID, adminID primitive.ObjectID) error

	// GetUserAnalytics retrieves user statistics
	GetUserAnalytics(ctx context.Context) (*UserAnalytics, error)
}

type UserAnalytics struct {
	TotalUsers      int `json:"total_users"`
	DeletedUsers    int `json:"deleted_users"`
	SuspendedUsers  int `json:"suspended_users"`
	VerifiedUsers   int `json:"verified_users"`
	UnverifiedUsers int `json:"unverified_users"`
}

type adminUserManagementService struct {
	userRepo repository.UserRepository
}

// NewAdminUserManagementService creates a new admin user management service
func NewAdminUserManagementService(userRepo repository.UserRepository) AdminUserManagementService {
	return &adminUserManagementService{
		userRepo: userRepo,
	}
}

// GetUser retrieves a user by ID
func (s *adminUserManagementService) GetUser(ctx context.Context, userID primitive.ObjectID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// ListAllUsers retrieves all users (paginated) with total count
func (s *adminUserManagementService) ListAllUsers(ctx context.Context, limit int, offset int) (*domain.PaginatedResponse[*domain.User], error) {
	users, err := s.userRepo.ListAllUsers(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.userRepo.CountAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	return domain.NewPaginatedResponse(users, total, limit, offset), nil
}

// SearchUsers searches users by email or name (paginated) with total count
func (s *adminUserManagementService) SearchUsers(ctx context.Context, query string, limit int, offset int) (*domain.PaginatedResponse[*domain.User], error) {
	users, err := s.userRepo.SearchUsers(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.userRepo.CountSearchUsers(ctx, query)
	if err != nil {
		return nil, err
	}

	return domain.NewPaginatedResponse(users, total, limit, offset), nil
}

// SuspendUser suspends a user
func (s *adminUserManagementService) SuspendUser(ctx context.Context, userID primitive.ObjectID, reason string, adminID primitive.ObjectID) error {
	if reason == "" {
		reason = "Suspended by admin"
	}
	return s.userRepo.SuspendUser(ctx, userID, reason, adminID)
}

// UnsuspendUser unsuspends a user
func (s *adminUserManagementService) UnsuspendUser(ctx context.Context, userID primitive.ObjectID, adminID primitive.ObjectID) error {
	return s.userRepo.UnsuspendUser(ctx, userID)
}

// DeleteUser deletes a user
func (s *adminUserManagementService) DeleteUser(ctx context.Context, userID, adminID primitive.ObjectID) error {
	return s.userRepo.SoftDelete(ctx, userID)
}

// GetUserAnalytics retrieves user statistics
func (s *adminUserManagementService) GetUserAnalytics(ctx context.Context) (*UserAnalytics, error) {
	totalUsers, err := s.userRepo.CountAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}

	deletedUsers, err := s.userRepo.CountDeletedUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count deleted users: %w", err)
	}

	suspendedUsers, err := s.userRepo.CountSuspendedUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count suspended users: %w", err)
	}

	verifiedUsers, err := s.userRepo.CountVerifiedUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count verified users: %w", err)
	}

	unverifiedUsers := totalUsers - verifiedUsers

	return &UserAnalytics{
		TotalUsers:      totalUsers,
		DeletedUsers:    deletedUsers,
		SuspendedUsers:  suspendedUsers,
		VerifiedUsers:   verifiedUsers,
		UnverifiedUsers: unverifiedUsers,
	}, nil
}
