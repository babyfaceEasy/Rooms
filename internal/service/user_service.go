package service

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// UserService defines user management operations.
type UserService interface {
	Register(ctx context.Context, name, email, password string, ageVerified bool) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	UpdateProfile(ctx context.Context, id, name, profilePictureURL string) (*domain.User, error)
	ChangePassword(ctx context.Context, id, currentPassword, newPassword string) error
	DeleteAccount(ctx context.Context, id string) error
	DeleteUser(ctx context.Context, id string) error
}

type userService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
}

// NewUserService creates a new UserService.
func NewUserService(userRepo repository.UserRepository, refreshTokenRepo repository.RefreshTokenRepository) UserService {
	return &userService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

// Register creates a new user account with validation and password hashing.
// generateUniqueCode generates a random 8-digit numeric code.
// Note: In production, consider implementing a database-level unique constraint on the code field.
func generateUniqueCode() string {
	const digits = "0123456789"
	code := make([]byte, 8)
	for i := range code {
		code[i] = digits[rand.Intn(len(digits))]
	}
	return string(code)
}

func (s *userService) Register(ctx context.Context, name, email, password string, ageVerified bool) (*domain.User, error) {
	// Validate age verification
	if !ageVerified {
		return nil, fmt.Errorf("age verification required: %w", domain.ErrAgeVerificationRequired)
	}

	// Validate inputs
	if err := validateName(name); err != nil {
		return nil, err
	}
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate unique customer code (8-digit numeric)
	code := generateUniqueCode()

	// Create user entity
	user := &domain.User{
		Code:          code,
		Name:          strings.TrimSpace(name),
		Email:         strings.ToLower(strings.TrimSpace(email)),
		PasswordHash:  string(passwordHash),
		IsAgeVerified: true,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Persist to repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID string.
func (s *userService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id %q: %w", id, domain.ErrInvalidInput)
	}
	return s.userRepo.GetByID(ctx, oid)
}

// DeleteUser removes a user by ID string.
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id %q: %w", id, domain.ErrInvalidInput)
	}
	return s.userRepo.Delete(ctx, oid)
}

// DeleteAccount performs a soft delete on the user account and invalidates all refresh tokens.
func (s *userService) DeleteAccount(ctx context.Context, id string) error {
	// Convert ID to ObjectID
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id %q: %w", id, domain.ErrInvalidInput)
	}

	// Verify user exists (GetByID also checks for soft-deleted users)
	_, err = s.userRepo.GetByID(ctx, oid)
	if err != nil {
		return err
	}

	// Soft delete the user account
	if err := s.userRepo.SoftDelete(ctx, oid); err != nil {
		return err
	}

	// Invalidate all refresh tokens for this user (logs out all devices)
	if err := s.refreshTokenRepo.DeleteByUserID(ctx, oid); err != nil {
		return fmt.Errorf("failed to invalidate refresh tokens: %w", err)
	}

	return nil
}

// UpdateProfile updates a user's profile name and optionally the profile picture URL.
func (s *userService) UpdateProfile(ctx context.Context, id, name, profilePictureURL string) (*domain.User, error) {
	// Validate inputs
	if err := validateName(name); err != nil {
		return nil, err
	}

	// Convert ID to ObjectID
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id %q: %w", id, domain.ErrInvalidInput)
	}

	// Get existing user
	user, err := s.userRepo.GetByID(ctx, oid)
	if err != nil {
		return nil, err
	}

	// Update fields
	user.Name = strings.TrimSpace(name)
	user.UpdatedAt = time.Now().UTC()
	if profilePictureURL != "" {
		user.ProfilePicture = profilePictureURL
	}

	// Persist changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword changes a user's password and invalidates all refresh tokens (logs out all devices).
func (s *userService) ChangePassword(ctx context.Context, id, currentPassword, newPassword string) error {
	// Convert ID to ObjectID
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id %q: %w", id, domain.ErrInvalidInput)
	}

	// Get existing user
	user, err := s.userRepo.GetByID(ctx, oid)
	if err != nil {
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return fmt.Errorf("invalid current password: %w", domain.ErrInvalidPassword)
	}

	// Validate new password
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	// Ensure new password is different from current
	if currentPassword == newPassword {
		return fmt.Errorf("new password must be different from current password: %w", domain.ErrInvalidPassword)
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password hash
	user.PasswordHash = string(newPasswordHash)
	user.UpdatedAt = time.Now().UTC()

	// Persist changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Invalidate all refresh tokens for this user (logs out all devices)
	if err := s.refreshTokenRepo.DeleteByUserID(ctx, oid); err != nil {
		return fmt.Errorf("failed to invalidate refresh tokens: %w", err)
	}

	return nil
}

// validateName checks that the name is valid.
func validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name is required: %w", domain.ErrInvalidInput)
	}
	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters: %w", domain.ErrInvalidInput)
	}
	if len(name) > 100 {
		return fmt.Errorf("name must be at most 100 characters: %w", domain.ErrInvalidInput)
	}
	return nil
}

// validateEmail checks that the email format is valid.
func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email is required: %w", domain.ErrInvalidEmail)
	}

	// Basic RFC 5322 validation
	const emailPattern = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailPattern, email)
	if err != nil || !matched {
		return fmt.Errorf("invalid email format: %w", domain.ErrInvalidEmail)
	}

	return nil
}

// validatePassword checks that the password meets strength requirements.
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters: %w", domain.ErrInvalidPassword)
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsSymbol(r) || unicode.IsPunct(r):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return fmt.Errorf("password must contain uppercase, lowercase, digit, and special character: %w", domain.ErrInvalidPassword)
	}

	return nil
}
