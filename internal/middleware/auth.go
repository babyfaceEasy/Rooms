package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"temp_backend/internal/service"
)

// AuthMiddleware validates JWT tokens and protects routes.
func AuthMiddleware(authService service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		// Validate Bearer scheme
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		tokenString := parts[1]

		// Validate token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		// Attach user info to request context
		c.Locals("user_id", claims.UserID.Hex())
		c.Locals("user_email", claims.Email)
		c.Locals("claims", claims)

		return c.Next()
	}
}
