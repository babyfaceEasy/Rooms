package handler

import (
	"fmt"

	"temp_backend/internal/domain"
	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminAuthHandler handles admin authentication
type AdminAuthHandler struct {
	authService service.AdminAuthService
	userService service.AdminUserService
}

// NewAdminAuthHandler creates a new admin auth handler
func NewAdminAuthHandler(authService service.AdminAuthService, userService service.AdminUserService) *AdminAuthHandler {
	return &AdminAuthHandler{
		authService: authService,
		userService: userService,
	}
}

// Login handles admin login request
// POST /api/v1/admin/auth/login
func (h *AdminAuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	accessToken, refreshToken, err := h.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		appErr, ok := err.(*domain.AppError)
		if ok {
			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "login failed",
		})
	}

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

// InitialSetup handles initial password setup request
// POST /api/v1/admin/auth/initial-setup
func (h *AdminAuthHandler) InitialSetup(c *fiber.Ctx) error {
	var req struct {
		Email     string `json:"email" binding:"required"`
		SetupToken string `json:"setup_token" binding:"required"`
		Password  string `json:"password" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":     "password too short",
			"code":      domain.ErrAdminInvalidPassword.Code,
			"min_chars": 8,
		})
	}

	accessToken, refreshToken, err := h.userService.CompleteSetup(c.Context(), req.Email, req.SetupToken, req.Password)
	if err != nil {
		appErr, ok := err.(*domain.AppError)
		if ok {
			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("setup failed: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"message":       "Password setup completed successfully",
	})
}

// ForgotPassword handles forgot password request
// POST /api/v1/admin/auth/forgot-password
func (h *AdminAuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	resetToken, err := h.authService.ForgotPassword(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to process request",
		})
	}

	// In production, this token would be sent via email
	// For now, we return it in response (only for development)
	return c.JSON(fiber.Map{
		"message": "Password reset instructions sent to email",
		"reset_token": resetToken, // Only for dev/testing - remove in production
	})
}

// ResetPassword handles password reset completion
// POST /api/v1/admin/auth/reset-password
func (h *AdminAuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req struct {
		ResetToken string `json:"reset_token" binding:"required"`
		Password   string `json:"password" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":     "password too short",
			"code":      domain.ErrAdminInvalidPassword.Code,
			"min_chars": 8,
		})
	}

	err := h.authService.ResetPassword(c.Context(), req.ResetToken, req.Password)
	if err != nil {
		appErr, ok := err.(*domain.AppError)
		if ok {
			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "password reset failed",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password reset successfully",
	})
}

// RefreshToken handles token refresh request
// POST /api/v1/admin/auth/refresh
func (h *AdminAuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	accessToken, err := h.authService.RefreshAdminToken(c.Context(), req.RefreshToken)
	if err != nil {
		appErr, ok := err.(*domain.AppError)
		if ok {
			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "token refresh failed",
		})
	}

	return c.JSON(fiber.Map{
		"access_token": accessToken,
		"token_type":   "Bearer",
	})
}

// Logout handles logout request
// POST /api/v1/admin/auth/logout
func (h *AdminAuthHandler) Logout(c *fiber.Ctx) error {
	adminID := c.Locals("admin_id")
	if adminID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "admin not authenticated",
		})
	}

	adminIDObj, ok := adminID.(primitive.ObjectID)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "invalid admin id format",
		})
	}

	// Logout invalidates all refresh tokens
	_ = h.authService.Logout(c.Context(), adminIDObj)

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}
