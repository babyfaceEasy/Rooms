package handler

import (
	"temp_backend/internal/domain"
	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportPostRequest represents the request to report a post
type ReportPostRequest struct {
	Reason  string  `json:"reason"`  // Report reason from predefined list
	Comment *string `json:"comment"` // Optional comment from reporter
}

// ReportHandler handles post reporting endpoints
type ReportHandler struct {
	reportSvc        service.ReportService
	maxReportsPerDay int
}

// NewReportHandler creates a new report handler
func NewReportHandler(reportSvc service.ReportService, maxReportsPerDay int) *ReportHandler {
	return &ReportHandler{
		reportSvc:        reportSvc,
		maxReportsPerDay: maxReportsPerDay,
	}
}

// ReportPost reports a post for violating community guidelines
func (h *ReportHandler) ReportPost(c *fiber.Ctx) error {
	// Extract user ID from context
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return domain.ErrUnauthorized
	}

	// Convert user ID from string to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_USER_ID", Message: "Invalid user ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get post ID from URL parameter
	postIDStr := c.Params("id")
	if postIDStr == "" {
		return &domain.AppError{Code: "POST_ID_REQUIRED", Message: "Post ID is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Convert post ID from string to ObjectID
	postObjID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		return &domain.AppError{Code: "INVALID_POST_ID", Message: "Invalid post ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Parse request body
	var req ReportPostRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.ErrInvalidInput
	}

	// Validate required fields
	if req.Reason == "" {
		return &domain.AppError{Code: "INVALID_INPUT", Message: "reason is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Check if user exceeded daily report limit
	// Note: This would need to be implemented in the service or repository
	// For now, we're delegating to the service

	// Report the post
	if err := h.reportSvc.ReportPost(c.Context(), postObjID, userObjID, req.Reason, req.Comment); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    nil,
		"message": "post reported successfully",
		"status":  fiber.StatusOK,
	})
}
