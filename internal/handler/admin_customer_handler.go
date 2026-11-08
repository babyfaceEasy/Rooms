package handler

import (
	"strconv"

	"temp_backend/internal/domain"
	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminCustomerHandler handles customer (user) management for admins
type AdminCustomerHandler struct {
	userMgmtService service.AdminUserManagementService
	logService      service.AdminLogService
}

// NewAdminCustomerHandler creates a new admin customer handler
func NewAdminCustomerHandler(userMgmtService service.AdminUserManagementService, logService service.AdminLogService) *AdminCustomerHandler {
	return &AdminCustomerHandler{
		userMgmtService: userMgmtService,
		logService:      logService,
	}
}

// GetUser retrieves a customer by ID
// GET /api/v1/admin/customers/:id
func (h *AdminCustomerHandler) GetUser(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	user, err := h.userMgmtService.GetUser(c.Context(), userID)
	if err != nil {
		appErr, ok := err.(*domain.AppError)
		if ok {
			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get user",
		})
	}

	// Remove sensitive fields
	user.PasswordHash = ""

	return c.JSON(user)
}

// ListUsers retrieves all customers
// GET /api/v1/admin/customers?limit=10&offset=0
func (h *AdminCustomerHandler) ListUsers(c *fiber.Ctx) error {
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

	users, err := h.userMgmtService.ListAllUsers(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list users",
		})
	}

	// Remove sensitive fields
	for _, user := range users {
		user.PasswordHash = ""
	}

	return c.JSON(fiber.Map{
		"data":   users,
		"limit":  limit,
		"offset": offset,
	})
}

// SuspendUser suspends a customer
// POST /api/v1/admin/customers/:id/suspend
func (h *AdminCustomerHandler) SuspendUser(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	adminID := c.Locals("admin_id").(primitive.ObjectID)
	adminEmail := c.Locals("admin_email").(string)

	var req struct {
		Reason string `json:"reason"`
	}
	c.BodyParser(&req)

	err = h.userMgmtService.SuspendUser(c.Context(), userID, req.Reason, adminID)
	if err != nil {
		appErr, ok := err.(*domain.AppError)
		if ok {
			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to suspend user",
		})
	}

	// Log action
	_ = h.logService.LogAction(c.Context(), adminID, adminEmail, domain.AdminLogActionUserSuspend, domain.AdminLogResourceTypeUser, &userID, map[string]interface{}{
		"reason": req.Reason,
	}, c.IP(), c.Get("User-Agent"))

	return c.JSON(fiber.Map{
		"message": "User suspended successfully",
	})
}

// UnsuspendUser unsuspends a customer
// DELETE /api/v1/admin/customers/:id/suspend
func (h *AdminCustomerHandler) UnsuspendUser(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	adminID := c.Locals("admin_id").(primitive.ObjectID)
	adminEmail := c.Locals("admin_email").(string)

	err = h.userMgmtService.UnsuspendUser(c.Context(), userID, adminID)
	if err != nil {
		appErr, ok := err.(*domain.AppError)
		if ok {
			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to unsuspend user",
		})
	}

	// Log action
	_ = h.logService.LogAction(c.Context(), adminID, adminEmail, domain.AdminLogActionUserUnsuspend, domain.AdminLogResourceTypeUser, &userID, nil, c.IP(), c.Get("User-Agent"))

	return c.JSON(fiber.Map{
		"message": "User unsuspended successfully",
	})
}

// DeleteUser deletes a customer
// DELETE /api/v1/admin/customers/:id
func (h *AdminCustomerHandler) DeleteUser(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	adminID := c.Locals("admin_id").(primitive.ObjectID)
	adminEmail := c.Locals("admin_email").(string)

	err = h.userMgmtService.DeleteUser(c.Context(), userID, adminID)
	if err != nil {
		appErr, ok := err.(*domain.AppError)
		if ok {
			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete user",
		})
	}

	// Log action
	_ = h.logService.LogAction(c.Context(), adminID, adminEmail, domain.AdminLogActionUserDelete, domain.AdminLogResourceTypeUser, &userID, nil, c.IP(), c.Get("User-Agent"))

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}
