package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"temp_backend/internal/domain"
	"temp_backend/internal/service"
)

// AuthHandler exposes HTTP endpoints for authentication.
type AuthHandler struct {
	svc service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// LoginRequest represents a login request payload.
type LoginRequest struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

// TokenResponse represents a token response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// RefreshTokenRequest represents a refresh token request payload.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Login handles POST /api/v1/auth/login requests.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return fmt.Errorf("parse login request: %w", err)
	}

	if req.Email == "" || req.Password == "" {
		return &domain.AppError{Code: "INVALID_INPUT", Message: "Email and password are required", HTTPStatus: fiber.StatusBadRequest}
	}

	accessToken, refreshToken, err := h.svc.Login(c.UserContext(), req.Email, req.Password)
	if err != nil {
		return err
	}

	// Access token TTL is 1 hour (3600 seconds)
	response := TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// RefreshAccessToken handles POST /api/v1/auth/refresh requests.
func (h *AuthHandler) RefreshAccessToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest

	if err := c.BodyParser(&req); err != nil {
		return fmt.Errorf("parse refresh token request: %w", err)
	}

	if req.RefreshToken == "" {
		return &domain.AppError{Code: "INVALID_INPUT", Message: "Refresh token is required", HTTPStatus: fiber.StatusBadRequest}
	}

	accessToken, err := h.svc.RefreshAccessToken(c.UserContext(), req.RefreshToken)
	if err != nil {
		return err
	}

	response := TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   3600,
		TokenType:   "Bearer",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// Logout handles POST /api/v1/auth/logout requests (protected).
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return domain.ErrUnauthorized
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return domain.ErrInternalServer
	}

	err := h.svc.Logout(c.UserContext(), userIDStr)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
