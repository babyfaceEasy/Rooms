package handler

import (
	"context"
	"errors"
	"fmt"

	"temp_backend/internal/domain"
	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

// UserHandler exposes HTTP endpoints for user management.
type UserHandler struct {
	svc          service.UserService
	emailService service.EmailService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc service.UserService, emailService service.EmailService) *UserHandler {
	return &UserHandler{svc: svc, emailService: emailService}
}

// RegisterRequest represents the registration request payload.
type RegisterRequest struct {
	Name        string `json:"name" form:"name"`
	Email       string `json:"email" form:"email"`
	Password    string `json:"password" form:"password"`
	AgeVerified bool   `json:"age_verified" form:"age_verified"`
}

// UserResponse represents a user in HTTP responses (no password).
type UserResponse struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// ProfileResponse represents a user profile response (minimal data).
type ProfileResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Code  string `json:"code"`
}

// UpdateProfileRequest represents the request payload for updating a profile.
type UpdateProfileRequest struct {
	Name string `json:"name"`
}

// ChangePasswordRequest represents the request payload for changing password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePasswordResponse represents the response after password change.
type ChangePasswordResponse struct {
	Message string `json:"message"`
}

// Register handles POST /api/v1/auth/register requests.
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return fmt.Errorf("parse register request: %w", err)
	}

	user, err := h.svc.Register(c.UserContext(), req.Name, req.Email, req.Password, req.AgeVerified)
	if err != nil {
		// Map domain errors to HTTP status codes
		return err
	}

	// Send verification email asynchronously (gracefully degrade if it fails)
	go func() {
		dynamicData := map[string]string{
			"user_name":  user.Name,
			"user_email": user.Email,
		}
		_ = h.emailService.SendVerificationEmail(context.Background(), user.ID, user.Email, dynamicData)
	}()

	response := UserResponse{
		ID:        user.ID.Hex(),
		Code:      user.Code,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// GetUser handles GET /api/v1/users/:id requests.
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")

	user, err := h.svc.GetUserByID(c.UserContext(), id)
	if err != nil {
		return err
	}

	response := UserResponse{
		ID:        user.ID.Hex(),
		Code:      user.Code,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// DeleteUser handles DELETE /api/v1/users/:id requests.
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.svc.DeleteUser(c.UserContext(), id)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ViewProfile handles GET /api/v1/profile requests.
// Returns the authenticated user's profile (name and email).
func (h *UserHandler) ViewProfile(c *fiber.Ctx) error {
	// Extract user ID from JWT context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return domain.ErrUnauthorized
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return fmt.Errorf("invalid user id type in context")
	}

	user, err := h.svc.GetUserByID(c.UserContext(), userIDStr)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("get user profile: %w", err)
	}

	response := ProfileResponse{
		Name:  user.Name,
		Email: user.Email,
		Code:  user.Code,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// UpdateProfile handles PATCH /api/v1/profile requests.
// Updates the authenticated user's profile (currently only name).
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	var req UpdateProfileRequest

	// Extract user ID from JWT context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return domain.ErrUnauthorized
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return fmt.Errorf("invalid user id type in context")
	}

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return domain.ErrInvalidInput
	}

	// Update user profile
	user, err := h.svc.UpdateProfile(c.UserContext(), userIDStr, req.Name)
	if err != nil {
		// Map domain errors to HTTP status codes
		return err
	}

	response := ProfileResponse{
		Name:  user.Name,
		Email: user.Email,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// ChangePassword handles POST /api/v1/profile/change-password requests.
// Changes the authenticated user's password and invalidates all refresh tokens.
func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	var req ChangePasswordRequest

	// Extract user ID from JWT context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return domain.ErrUnauthorized
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return fmt.Errorf("invalid user id type in context")
	}

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return domain.ErrInvalidInput
	}

	// Validate request fields
	if req.CurrentPassword == "" {
		return &domain.AppError{Code: "INVALID_INPUT", Message: "current_password is required", HTTPStatus: fiber.StatusBadRequest}
	}
	if req.NewPassword == "" {
		return &domain.AppError{Code: "INVALID_INPUT", Message: "new_password is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Change password
	err := h.svc.ChangePassword(c.UserContext(), userIDStr, req.CurrentPassword, req.NewPassword)
	if err != nil {
		// Map domain errors to HTTP status codes
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ChangePasswordResponse{
		Message: "password changed successfully, please log in again",
	})
}

// DeleteAccount handles DELETE /api/v1/profile requests.
// Soft-deletes the authenticated user's account and invalidates all refresh tokens.
func (h *UserHandler) DeleteAccount(c *fiber.Ctx) error {
	// Extract user ID from JWT context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return domain.ErrUnauthorized
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return fmt.Errorf("invalid user id type in context")
	}

	// Delete the account (soft delete)
	err := h.svc.DeleteAccount(c.UserContext(), userIDStr)
	if err != nil {
		// Map domain errors to HTTP status codes
		return err
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
