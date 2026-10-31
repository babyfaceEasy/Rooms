package handler

import (
	"strconv"

	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminLogsHandler handles logs endpoints
type AdminLogsHandler struct {
	logService service.AdminLogService
}

// NewAdminLogsHandler creates a new admin logs handler
func NewAdminLogsHandler(logService service.AdminLogService) *AdminLogsHandler {
	return &AdminLogsHandler{
		logService: logService,
	}
}

// GetLogs retrieves all admin logs
// GET /api/v1/admin/logs?limit=50&offset=0
func (h *AdminLogsHandler) GetLogs(c *fiber.Ctx) error {
	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	logs, err := h.logService.GetLogs(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get logs",
		})
	}

	return c.JSON(fiber.Map{
		"data":   logs,
		"limit":  limit,
		"offset": offset,
	})
}

// GetAdminLogs retrieves logs for a specific admin
// GET /api/v1/admin/logs/admin/:admin_id?limit=50&offset=0
func (h *AdminLogsHandler) GetAdminLogs(c *fiber.Ctx) error {
	adminIDStr := c.Params("admin_id")
	adminID, err := primitive.ObjectIDFromHex(adminIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid admin id",
		})
	}

	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	logs, err := h.logService.GetAdminLogs(c.Context(), adminID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get logs",
		})
	}

	return c.JSON(fiber.Map{
		"data":   logs,
		"limit":  limit,
		"offset": offset,
	})
}
