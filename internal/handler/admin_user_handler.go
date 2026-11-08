package handler

import (
	"strconv"

	"temp_backend/internal/domain"
	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminUserHandler handles admin user management
type AdminUserHandler struct {
	adminService service.AdminUserService
}

// NewAdminUserHandler creates a new admin user handler
func NewAdminUserHandler(adminService service.AdminUserService) *AdminUserHandler {
	return &AdminUserHandler{
		adminService: adminService,
	}
}

// GetAdmin retrieves an admin by ID
// GET /api/v1/admin/admins/:id
func (h *AdminUserHandler) GetAdmin(c *fiber.Ctx) error {
	adminIDStr := c.Params("id")
	adminID, err := primitive.ObjectIDFromHex(adminIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid admin id",
		})
	}

	admin, err := h.adminService.GetAdmin(c.Context(), adminID)
	if err != nil {
		appErr, ok := err.(*domain.AppError)
		if ok {
			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get admin",
		})
	}

	// Remove sensitive fields
	admin.PasswordHash = ""
	admin.SetupToken = ""
	admin.PasswordResetToken = ""

	return c.JSON(admin)
}

// ListAdmins retrieves all admins
// GET /api/v1/admin/admins?limit=10&offset=0
func (h *AdminUserHandler) ListAdmins(c *fiber.Ctx) error {
	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	admins, err := h.adminService.ListAdmins(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list admins",
		})
	}

	// Remove sensitive fields
	for _, admin := range admins {
		admin.PasswordHash = ""
		admin.SetupToken = ""
		admin.PasswordResetToken = ""
	}

	return c.JSON(fiber.Map{
		"data":   admins,
		"limit":  limit,
		"offset": offset,
	})
}

// CreateAdmin creates a new admin
// POST /api/v1/admin/admins
func (h *AdminUserHandler) CreateAdmin(c *fiber.Ctx) error {
	var req struct {
		Email string             `json:"email" binding:"required"`
		Name  string             `json:"name" binding:"required"`
		Role  domain.AdminRole   `json:"role" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	admin, setupToken, err := h.adminService.CreateAdmin(c.Context(), req.Email, req.Name, req.Role)
	if err != nil {
		appErr, ok := err.(*domain.AppError)
		if ok {
			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create admin",
		})
	}

	admin.PasswordHash = ""
	admin.SetupToken = ""
	admin.PasswordResetToken = ""

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"admin": admin,
		"setup_token": setupToken,
		"message": "Admin created. Send setup_token to admin via secure channel for password setup.",
	})
}

// UpdateAdmin updates an admin
// PATCH /api/v1/admin/admins/:id
func (h *AdminUserHandler) UpdateAdmin(c *fiber.Ctx) error {
	adminIDStr := c.Params("id")
	adminID, err := primitive.ObjectIDFromHex(adminIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid admin id",
		})
	}

	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.adminService.UpdateAdmin(c.Context(), adminID, updates); err != nil {
		appErr, ok := err.(*domain.AppError)
		if ok {
			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update admin",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Admin updated successfully",
	})
}

// DeleteAdmin deletes an admin
// DELETE /api/v1/admin/admins/:id
func (h *AdminUserHandler) DeleteAdmin(c *fiber.Ctx) error {
	adminIDStr := c.Params("id")
	adminID, err := primitive.ObjectIDFromHex(adminIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid admin id",
		})
	}

	requesterIDInterface := c.Locals("admin_id")
	if requesterIDInterface == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "admin not authenticated",
		})
	}

	requesterID := requesterIDInterface.(primitive.ObjectID)

	if err := h.adminService.DeleteAdmin(c.Context(), adminID, requesterID); err != nil {
		appErr, ok := err.(*domain.AppError)
		if ok {
			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete admin",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Admin deleted successfully",
	})
}
