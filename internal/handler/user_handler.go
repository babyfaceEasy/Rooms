package handler

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"
	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserHandler exposes HTTP endpoints for user management.
type UserHandler struct {
	svc          service.UserService
	emailService service.EmailService
	storage      repository.ObjectStorage
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc service.UserService, emailService service.EmailService, storage repository.ObjectStorage) *UserHandler {
	return &UserHandler{svc: svc, emailService: emailService, storage: storage}
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
	ID             string `json:"id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	ProfilePicture string `json:"profile_picture,omitempty"`
	CreatedAt      string `json:"created_at"`
}

// ProfileResponse represents a user profile response (minimal data).
type ProfileResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Code           string `json:"code"`
	ProfilePicture string `json:"profile_picture"`
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

	user, verificationToken, err := h.svc.Register(c.UserContext(), req.Name, req.Email, req.Password, req.AgeVerified)
	if err != nil {
		// Map domain errors to HTTP status codes
		return err
	}

	// Send verification email asynchronously (gracefully degrade if it fails)
	go func() {
		// Extract first word from name for personalization
		firstName := strings.Fields(user.Name)
		userFirstName := user.Name
		if len(firstName) > 0 {
			userFirstName = firstName[0]
		}
		dynamicData := map[string]string{
			"user_name":          userFirstName,
			"user_email":         user.Email,
			"verification_url":   "https://tempbackend.com/verify-email?token=" + verificationToken,
			"verification_token": verificationToken,
			"year":               strconv.Itoa(time.Now().Year()),
		}
		_, _ = h.emailService.SendVerificationEmail(context.Background(), user.ID, user.Email, dynamicData)
		// Email service logs internally, so we just ignore the result here for async send
	}()

	response := UserResponse{
		ID:             user.ID.Hex(),
		Code:           user.Code,
		Name:           user.Name,
		Email:          user.Email,
		ProfilePicture: user.ProfilePicture,
		CreatedAt:      user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// VerifyEmailRequest represents the request payload for email verification.
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// VerifyEmail handles POST /api/v1/auth/verify-email requests.
func (h *UserHandler) VerifyEmail(c *fiber.Ctx) error {
	var req VerifyEmailRequest

	if err := c.BodyParser(&req); err != nil {
		return fmt.Errorf("parse verify email request: %w", err)
	}

	if req.Token == "" {
		return &domain.AppError{Code: "INVALID_INPUT", Message: "Verification token is required", HTTPStatus: fiber.StatusBadRequest}
	}

	if err := h.svc.VerifyEmail(c.UserContext(), req.Token); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Email verified successfully",
	})
}

// ResendVerificationEmailRequest represents the request payload for resending verification email.
type ResendVerificationEmailRequest struct {
	Email string `json:"email"`
}

// ResendVerificationEmail handles POST /api/v1/auth/resend-verification-email requests.
func (h *UserHandler) ResendVerificationEmail(c *fiber.Ctx) error {
	var req ResendVerificationEmailRequest

	if err := c.BodyParser(&req); err != nil {
		return fmt.Errorf("parse resend verification email request: %w", err)
	}

	if req.Email == "" {
		return &domain.AppError{Code: "INVALID_INPUT", Message: "Email is required", HTTPStatus: fiber.StatusBadRequest}
	}

	verificationToken, err := h.svc.ResendVerificationEmail(c.UserContext(), req.Email)
	if err != nil {
		return err
	}

	// Send verification email asynchronously (gracefully degrade if it fails)
	if verificationToken != "" {
		go func() {
			dynamicData := map[string]string{
				"user_name":          "", // We could fetch user here, but keeping it generic for resend
				"user_email":         req.Email,
				"verification_url":   "https://tempbackend.com/verify-email?token=" + verificationToken,
				"verification_token": verificationToken,
				"year":               strconv.Itoa(time.Now().Year()),
			}
			_, _ = h.emailService.SendVerificationEmail(context.Background(), primitive.ObjectID{}, req.Email, dynamicData)
			// Email service logs internally, so we just ignore the result here for async send
		}()
	}

	// Always return success to prevent email enumeration
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "If an account with this email exists and is not verified, a verification email has been sent.",
	})
}

// ForgotPasswordRequest represents the request payload for password reset initiation.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPassword handles POST /api/v1/auth/forgot-password requests.
func (h *UserHandler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return fmt.Errorf("parse forgot password request: %w", err)
	}

	if req.Email == "" {
		return &domain.AppError{Code: "INVALID_INPUT", Message: "Email is required", HTTPStatus: fiber.StatusBadRequest}
	}

	resetToken, userID, err := h.svc.ForgotPassword(c.UserContext(), req.Email)
	if err != nil {
		return err
	}

	// Only send email if a token was actually generated (user exists)
	if resetToken != "" {
		// Send password reset email asynchronously (gracefully degrade if it fails)
		go func() {
			dynamicData := map[string]string{
				"reset_url":   "https://tempbackend.com/reset-password?token=" + resetToken,
				"reset_token": resetToken,
				"year":        strconv.Itoa(time.Now().Year()),
			}
			_, _ = h.emailService.SendPasswordResetEmail(context.Background(), userID, req.Email, dynamicData)
			// Email service logs internally, so we just ignore the result here for async send
		}()
	}

	// Always return success to prevent email enumeration
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "If an account with this email exists, a password reset link has been sent.",
	})
}

// ResetPasswordRequest represents the request payload for password reset.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ResetPassword handles POST /api/v1/auth/reset-password requests.
func (h *UserHandler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return fmt.Errorf("parse reset password request: %w", err)
	}

	if req.Token == "" {
		return &domain.AppError{Code: "INVALID_INPUT", Message: "Reset token is required", HTTPStatus: fiber.StatusBadRequest}
	}
	if req.NewPassword == "" {
		return &domain.AppError{Code: "INVALID_INPUT", Message: "New password is required", HTTPStatus: fiber.StatusBadRequest}
	}

	if err := h.svc.ResetPassword(c.UserContext(), req.Token, req.NewPassword); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password has been reset successfully. Please log in with your new password.",
	})
}

// GetUser handles GET /api/v1/users/:id requests.
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")

	user, err := h.svc.GetUserByID(c.UserContext(), id)
	if err != nil {
		return err
	}

	response := UserResponse{
		ID:             user.ID.Hex(),
		Code:           user.Code,
		Name:           user.Name,
		Email:          user.Email,
		ProfilePicture: user.ProfilePicture,
		CreatedAt:      user.CreatedAt.Format("2006-01-02T15:04:05Z"),
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
		ID:             user.ID.Hex(),
		Name:           user.Name,
		Email:          user.Email,
		Code:           user.Code,
		ProfilePicture: user.ProfilePicture,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// UpdateProfile handles PATCH /api/v1/profile requests.
// Updates the authenticated user's profile (name and optional profile picture).
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	// Extract user ID from JWT context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return domain.ErrUnauthorized
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return fmt.Errorf("invalid user id type in context")
	}

	// Parse name from form data
	name := c.FormValue("name")
	if name == "" {
		return &domain.AppError{Code: "INVALID_INPUT", Message: "name is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Handle optional profile picture upload
	var profilePictureURL string
	file, err := c.FormFile("profile_picture")
	if err != nil {
		if !errors.Is(err, fasthttp.ErrMissingFile) {
			return &domain.AppError{Code: "MEDIA_PROCESSING_FAILED", Message: "Failed to process profile picture", HTTPStatus: fiber.StatusBadRequest}
		}
		// No file uploaded — that's fine, it's optional
	} else {
		// Validate image type
		ext := filepath.Ext(file.Filename)
		validTypes := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
		if !validTypes[ext] {
			return &domain.AppError{Code: "INVALID_MEDIA_TYPE", Message: "Invalid image type. Allowed: jpg, jpeg, png, gif, webp", HTTPStatus: fiber.StatusBadRequest}
		}

		// Open the file
		src, err := file.Open()
		if err != nil {
			return &domain.AppError{Code: "MEDIA_PROCESSING_FAILED", Message: "Failed to process profile picture", HTTPStatus: fiber.StatusBadRequest}
		}
		defer src.Close()

		// Upload to S3
		key := "profile_pictures/" + userIDStr + "/" + file.Filename
		url, err := h.storage.PutObject(c.UserContext(), key, src, file.Size, file.Header.Get("Content-Type"))
		if err != nil {
			return &domain.AppError{Code: "MEDIA_UPLOAD_FAILED", Message: "Failed to upload profile picture", HTTPStatus: fiber.StatusInternalServerError}
		}
		profilePictureURL = url
	}

	// Update user profile
	user, err := h.svc.UpdateProfile(c.UserContext(), userIDStr, name, profilePictureURL)
	if err != nil {
		return err
	}

	response := ProfileResponse{
		ID:             user.ID.Hex(),
		Name:           user.Name,
		Email:          user.Email,
		Code:           user.Code,
		ProfilePicture: user.ProfilePicture,
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
