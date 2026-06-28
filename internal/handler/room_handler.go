package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/internal/domain"
	"temp_backend/internal/service"
)

// CreateRoomRequest represents the request to create a room
type CreateRoomRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// RoomResponse represents a room in the response
type RoomResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// RoomHandler handles room-related endpoints
type RoomHandler struct {
	svc service.RoomService
}

// NewRoomHandler creates a new room handler
func NewRoomHandler(svc service.RoomService) *RoomHandler {
	return &RoomHandler{svc: svc}
}

// CreateRoom creates a new room
func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	// Extract user ID from context
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]interface{}{
			"error":  "unauthorized",
			"status": fiber.StatusUnauthorized,
		})
	}

	// Convert user ID from string to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "invalid user id",
			"status": fiber.StatusBadRequest,
		})
	}

	// Parse request body
	var req CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "invalid input",
			"status": fiber.StatusBadRequest,
		})
	}

	// Create room via service
	room, err := h.svc.CreateRoom(c.Context(), req.Name, req.Code, userObjID)
	if err != nil {
		return h.handleError(c, err)
	}

	// Convert domain Room to RoomResponse
	response := &RoomResponse{
		ID:        room.ID.Hex(),
		Name:      room.Name,
		Code:      room.Code,
		CreatedBy: room.CreatedBy.Hex(),
		CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: room.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return c.Status(fiber.StatusCreated).JSON(map[string]interface{}{
		"data":    response,
		"message": "room created successfully",
		"status":  fiber.StatusCreated,
	})
}

// handleError maps service errors to HTTP responses
func (h *RoomHandler) handleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "invalid input",
			"status": fiber.StatusBadRequest,
		})
	case errors.Is(err, domain.ErrCodeAlreadyExists):
		return c.Status(fiber.StatusConflict).JSON(map[string]interface{}{
			"error":  "room code already exists",
			"status": fiber.StatusConflict,
		})
	case errors.Is(err, domain.ErrRoomNotFound):
		return c.Status(fiber.StatusNotFound).JSON(map[string]interface{}{
			"error":  "room not found",
			"status": fiber.StatusNotFound,
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
			"error":  "internal server error",
			"status": fiber.StatusInternalServerError,
		})
	}
}
