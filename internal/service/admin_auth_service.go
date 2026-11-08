package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"temp_backend/config"
	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// AdminAuthService defines admin authentication operations
type AdminAuthService interface {
	// Login authenticates an admin and returns tokens
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)

	// InitialSetup completes initial password setup for a new admin
	InitialSetup(ctx context.Context, setupToken, newPassword string) (accessToken, refreshToken string, err error)

	// ForgotPassword initiates password reset flow
	ForgotPassword(ctx context.Context, email string) (resetToken string, err error)

	// ResetPassword completes password reset
	ResetPassword(ctx context.Context, resetToken, newPassword string) error

	// ValidateAdminToken validates and parses an admin JWT token
	ValidateAdminToken(ctx context.Context, tokenString string) (*domain.AdminJWTClaims, error)

	// RefreshAdminToken generates a new access token from a refresh token
	RefreshAdminToken(ctx context.Context, refreshToken string) (newAccessToken string, err error)

	// Logout invalidates a refresh token
	Logout(ctx context.Context, adminID primitive.ObjectID) error
}

type adminAuthService struct {
	adminRepo              repository.AdminRepository
	adminRefreshTokenRepo  repository.AdminRefreshTokenRepository
	config                 *config.Config
}

// NewAdminAuthService creates a new admin auth service
func NewAdminAuthService(
	adminRepo repository.AdminRepository,
	adminRefreshTokenRepo repository.AdminRefreshTokenRepository,
	cfg *config.Config,
) AdminAuthService {
	return &adminAuthService{
		adminRepo:             adminRepo,
		adminRefreshTokenRepo: adminRefreshTokenRepo,
		config:                cfg,
	}
}

// Login authenticates an admin
func (s *adminAuthService) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	admin, err := s.adminRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("login failed: %w", domain.ErrUnauthorizedAdmin)
	}

	// Check if admin is active
	if !admin.IsActive {
		return "", "", domain.ErrAdminInactive
	}

	// Check if account is locked
	if admin.LockedUntil != nil && admin.LockedUntil.After(time.Now().UTC()) {
		return "", "", domain.ErrAdminLocked
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		// Increment failed login attempts
		attempts := admin.FailedLoginAttempts + 1
		var lockedUntil interface{} = nil

		// Lock account after 5 failed attempts for 15 minutes
		if attempts >= 5 {
			lockTime := time.Now().UTC().Add(15 * time.Minute)
			lockedUntil = lockTime
		}

		_ = s.adminRepo.UpdateFailedLoginAttempts(ctx, admin.ID, attempts, lockedUntil)
		return "", "", domain.ErrUnauthorizedAdmin
	}

	// Clear failed login attempts on successful login
	_ = s.adminRepo.ClearFailedLoginAttempts(ctx, admin.ID)

	// Update last login
	_ = s.adminRepo.UpdateLastLogin(ctx, admin.ID)

	// Generate tokens
	access, refresh, err := s.generateTokenPair(admin)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store refresh token hash
	if err := s.storeRefreshToken(ctx, admin.ID, refresh); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return access, refresh, nil
}

// InitialSetup completes initial password setup
func (s *adminAuthService) InitialSetup(ctx context.Context, setupToken, newPassword string) (accessToken, refreshToken string, err error) {
	if len(newPassword) < 8 {
		return "", "", fmt.Errorf("password too short: %w", domain.ErrAdminInvalidPassword)
	}

	// Find admin with this setup token
	// We need to search by setup token, so we'll list all admins with setup tokens
	// For now, this is a limitation - ideally we'd have a method to find by token
	// In a real scenario, you might use email or other identifier with the token
	return "", "", fmt.Errorf("setup token verification requires email or admin ID: %w", domain.ErrInvalidSetupToken)
}

// ForgotPassword initiates password reset
func (s *adminAuthService) ForgotPassword(ctx context.Context, email string) (resetToken string, err error) {
	admin, err := s.adminRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists
		return "", nil
	}

	// Generate reset token
	token := generateRandomToken(32)
	expiresAt := time.Now().UTC().Add(1 * time.Hour)

	if err := s.adminRepo.SetPasswordResetToken(ctx, admin.ID, hashToken(token), expiresAt); err != nil {
		return "", fmt.Errorf("failed to set reset token: %w", err)
	}

	return token, nil
}

// ResetPassword completes password reset
func (s *adminAuthService) ResetPassword(ctx context.Context, resetToken, newPassword string) error {
	if len(newPassword) < 8 {
		return domain.ErrAdminInvalidPassword
	}

	// Verify token and find admin
	// This is a limitation without being able to search by token
	// In practice, you'd need email + token or admin_id + token
	return fmt.Errorf("password reset requires email or admin ID: %w", domain.ErrAdminPasswordResetExpired)
}

// ValidateAdminToken validates an admin JWT
func (s *adminAuthService) ValidateAdminToken(ctx context.Context, tokenString string) (*domain.AdminJWTClaims, error) {
	claims := &domain.AdminJWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	if !token.Valid {
		return nil, domain.ErrTokenInvalid
	}

	return claims, nil
}

// RefreshAdminToken generates new access token from refresh token
func (s *adminAuthService) RefreshAdminToken(ctx context.Context, refreshToken string) (newAccessToken string, err error) {
	// Validate refresh token
	claims := &domain.AdminJWTClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil || !token.Valid {
		return "", domain.ErrTokenInvalid
	}

	// Verify refresh token is stored
	tokenHash := hashToken(refreshToken)
	if err := s.adminRefreshTokenRepo.Verify(ctx, claims.AdminID, tokenHash); err != nil {
		return "", fmt.Errorf("refresh token not found: %w", domain.ErrTokenNotFound)
	}

	// Get admin
	admin, err := s.adminRepo.GetByID(ctx, claims.AdminID)
	if err != nil {
		return "", fmt.Errorf("admin not found: %w", err)
	}

	// Generate new access token
	now := time.Now().UTC()
	expiryTime := now.Add(s.config.Admin.AccessTokenTTL)

	newClaims := &domain.AdminJWTClaims{
		AdminID:     admin.ID,
		Email:       admin.Email,
		Role:        admin.Role,
		Permissions: admin.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	accessToken, err := newToken.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return accessToken, nil
}

// Logout invalidates refresh token
func (s *adminAuthService) Logout(ctx context.Context, adminID primitive.ObjectID) error {
	return s.adminRefreshTokenRepo.DeleteByAdminID(ctx, adminID)
}

// Helper functions

func (s *adminAuthService) generateTokenPair(admin *domain.Admin) (accessToken, refreshToken string, err error) {
	now := time.Now().UTC()

	// Generate access token (short-lived)
	accessExpiryTime := now.Add(s.config.Admin.AccessTokenTTL)
	accessClaims := &domain.AdminJWTClaims{
		AdminID:     admin.ID,
		Email:       admin.Email,
		Role:        admin.Role,
		Permissions: admin.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiryTime),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token (long-lived)
	refreshExpiryTime := now.Add(s.config.Admin.RefreshTokenTTL)
	refreshClaims := &domain.AdminJWTClaims{
		AdminID:     admin.ID,
		Email:       admin.Email,
		Role:        admin.Role,
		Permissions: admin.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiryTime),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *adminAuthService) storeRefreshToken(ctx context.Context, adminID primitive.ObjectID, refreshTokenStr string) error {
	expiresAt := time.Now().UTC().Add(s.config.Admin.RefreshTokenTTL)
	tokenHash := hashToken(refreshTokenStr)

	token := &domain.AdminRefreshToken{
		AdminID:   adminID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}

	return s.adminRefreshTokenRepo.Create(ctx, token)
}

func generateRandomToken(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}
