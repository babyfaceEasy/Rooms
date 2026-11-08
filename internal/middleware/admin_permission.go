package middleware

import (
	"temp_backend/internal/domain"

	"github.com/gofiber/fiber/v2"
)

// AdminPermissionMiddleware checks if admin has required permissions
func AdminPermissionMiddleware(requiredPermissions []domain.AdminPermission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get admin permissions from context
		permissionsInterface := c.Locals("admin_permissions")
		if permissionsInterface == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "admin permissions not found in context",
			})
		}

		permissions, ok := permissionsInterface.([]domain.AdminPermission)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "invalid permissions format",
			})
		}

		// Get role for super admin check
		role := c.Locals("admin_role")

		// Super admin has all permissions
		if role == domain.AdminRoleSuperAdmin {
			return c.Next()
		}

		// Check if admin has all required permissions
		permissionMap := make(map[domain.AdminPermission]bool)
		for _, p := range permissions {
			permissionMap[p] = true
		}

		for _, required := range requiredPermissions {
			if !permissionMap[required] {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "insufficient permissions",
					"code":  domain.ErrInsufficientPermissions.Code,
				})
			}
		}

		return c.Next()
	}
}
