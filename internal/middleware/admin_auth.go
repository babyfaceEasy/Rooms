package middleware

import (
	"strings"

	"temp_backend/internal/domain"
	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

// AdminAuthMiddleware validates JWT tokens for admin routes
func AdminAuthMiddleware(authService service.AdminAuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		// Extract bearer token
		const bearerScheme = "Bearer "
		if !strings.HasPrefix(authHeader, bearerScheme) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		token := authHeader[len(bearerScheme):]

		// Validate token
		claims, err := authService.ValidateAdminToken(c.Context(), token)
		if err != nil {
			appErr, ok := err.(*domain.AppError)
			if ok {
				return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
					"error": appErr.Message,
					"code":  appErr.Code,
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		// Store claims and admin ID in context
		c.Locals("admin_id", claims.AdminID)
		c.Locals("admin_email", claims.Email)
		c.Locals("admin_role", claims.Role)
		c.Locals("admin_permissions", claims.Permissions)
		c.Locals("admin_claims", claims)

		return c.Next()
	}
}
