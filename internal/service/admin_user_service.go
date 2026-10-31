package service

import (
	"context"
	"fmt"
	"time"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// AdminUserService defines admin account management operations
type AdminUserService interface {
	// GetAdmin retrieves an admin by ID
	GetAdmin(ctx context.Context, adminID primitive.ObjectID) (*domain.Admin, error)

	// ListAdmins retrieves all admins (paginated)
	ListAdmins(ctx context.Context, limit int, offset int) ([]*domain.Admin, error)

	// CreateAdmin creates a new admin with setup token for initial password setup
	CreateAdmin(ctx context.Context, email, name string, role domain.AdminRole) (*domain.Admin, string, error)

	// UpdateAdmin updates admin details
	UpdateAdmin(ctx context.Context, adminID primitive.ObjectID, updates map[string]interface{}) error

	// DeleteAdmin deletes an admin (only super_admin can delete)
	DeleteAdmin(ctx context.Context, adminID, requesterID primitive.ObjectID) error

	// ChangeAdminRole changes an admin's role
	ChangeAdminRole(ctx context.Context, adminID primitive.ObjectID, newRole domain.AdminRole) error

	// CompleteSetup completes initial password setup
	CompleteSetup(ctx context.Context, email, setupToken, password string) (accessToken, refreshToken string, err error)
}

type adminUserService struct {
	adminRepo repository.AdminRepository
	authService AdminAuthService
}

// NewAdminUserService creates a new admin user service
func NewAdminUserService(
	adminRepo repository.AdminRepository,
	authService AdminAuthService,
) AdminUserService {
	return &adminUserService{
		adminRepo: adminRepo,
		authService: authService,
	}
}

// GetAdmin retrieves an admin by ID
func (s *adminUserService) GetAdmin(ctx context.Context, adminID primitive.ObjectID) (*domain.Admin, error) {
	return s.adminRepo.GetByID(ctx, adminID)
}

// ListAdmins retrieves all admins (paginated)
func (s *adminUserService) ListAdmins(ctx context.Context, limit int, offset int) ([]*domain.Admin, error) {
	return s.adminRepo.List(ctx, limit, offset)
}

// CreateAdmin creates a new admin with setup token
func (s *adminUserService) CreateAdmin(ctx context.Context, email, name string, role domain.AdminRole) (*domain.Admin, string, error) {
	// Validate role
	tempAdmin := &domain.Admin{Email: email, Name: name, Role: role}
	if err := tempAdmin.ValidateRole(); err != nil {
		return nil, "", err
	}

	// Check if admin already exists
	existing, err := s.adminRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, "", domain.ErrAdminAlreadyExists
	}

	// Generate setup token
	setupToken := generateRandomToken(32)
	setupTokenExpiresAt := time.Now().UTC().Add(24 * time.Hour)

	// Create admin with temporary password
	admin := &domain.Admin{
		Email:                  email,
		Name:                   name,
		Role:                   role,
		Permissions:            domain.DefaultPermissionsForRole(role),
		IsActive:               true,
		RequiresPasswordSetup:  true,
		SetupToken:             setupToken,
		SetupTokenExpiresAt:    &setupTokenExpiresAt,
		PasswordHash:           "", // Will be set during setup
		FailedLoginAttempts:    0,
	}

	if err := admin.Validate(); err != nil {
		return nil, "", err
	}

	if err := s.adminRepo.Create(ctx, admin); err != nil {
		return nil, "", err
	}

	return admin, setupToken, nil
}

// UpdateAdmin updates admin details
func (s *adminUserService) UpdateAdmin(ctx context.Context, adminID primitive.ObjectID, updates map[string]interface{}) error {
	admin, err := s.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		return err
	}

	// Update allowed fields
	if name, ok := updates["name"].(string); ok && name != "" {
		admin.Name = name
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		admin.IsActive = isActive
	}

	admin.UpdatedAt = time.Now().UTC()

	return s.adminRepo.Update(ctx, admin)
}

// DeleteAdmin deletes an admin
func (s *adminUserService) DeleteAdmin(ctx context.Context, adminID, requesterID primitive.ObjectID) error {
	// Cannot delete self
	if adminID == requesterID {
		return domain.ErrCannotDeleteSelfAdmin
	}

	admin, err := s.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		return err
	}

	// Cannot delete if it's the last super admin
	if admin.Role == domain.AdminRoleSuperAdmin {
		count, err := s.adminRepo.CountSuperAdmins(ctx)
		if err != nil {
			return err
		}
		if count <= 1 {
			return domain.ErrCannotDeleteLastSuperAdmin
		}
	}

	return s.adminRepo.Delete(ctx, adminID)
}

// ChangeAdminRole changes an admin's role
func (s *adminUserService) ChangeAdminRole(ctx context.Context, adminID primitive.ObjectID, newRole domain.AdminRole) error {
	admin, err := s.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		return err
	}

	admin.Role = newRole
	admin.Permissions = domain.DefaultPermissionsForRole(newRole)
	admin.UpdatedAt = time.Now().UTC()

	return s.adminRepo.Update(ctx, admin)
}

// CompleteSetup completes initial password setup
func (s *adminUserService) CompleteSetup(ctx context.Context, email, setupToken, password string) (accessToken, refreshToken string, err error) {
	if len(password) < 8 {
		return "", "", domain.ErrAdminInvalidPassword
	}

	admin, err := s.adminRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}

	// Verify setup token
	valid, err := s.adminRepo.VerifySetupToken(ctx, admin.ID, setupToken)
	if err != nil {
		return "", "", err
	}
	if !valid {
		return "", "", domain.ErrInvalidSetupToken
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Update admin
	admin.PasswordHash = string(hash)
	admin.RequiresPasswordSetup = false
	admin.UpdatedAt = time.Now().UTC()

	if err := s.adminRepo.Update(ctx, admin); err != nil {
		return "", "", err
	}

	// Clear setup token
	if err := s.adminRepo.ClearSetupToken(ctx, admin.ID); err != nil {
		return "", "", err
	}

	// Log them in
	return s.authService.Login(ctx, email, password)
}
