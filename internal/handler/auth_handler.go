package handler

import (
	"errors"
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email and password are required",
		})
	}

	accessToken, refreshToken, err := h.svc.Login(c.UserContext(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid email or password",
			})
		}
		return fmt.Errorf("login: %w", err)
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "refresh_token is required",
		})
	}

	accessToken, err := h.svc.RefreshAccessToken(c.UserContext(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrTokenExpired) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "refresh token has expired",
			})
		}
		if errors.Is(err, domain.ErrTokenInvalid) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid refresh token",
			})
		}
		return fmt.Errorf("refresh access token: %w", err)
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
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	err := h.svc.Logout(c.UserContext(), userIDStr)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid user id",
			})
		}
		return fmt.Errorf("logout: %w", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
