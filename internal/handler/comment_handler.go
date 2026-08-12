package handler

import (
	"errors"

	"temp_backend/internal/domain"
	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CommentHandler handles HTTP requests for comments
type CommentHandler struct {
	svc service.CommentService
}

// NewCommentHandler creates a new CommentHandler
func NewCommentHandler(svc service.CommentService) *CommentHandler {
	return &CommentHandler{
		svc: svc,
	}
}

// CreateCommentRequest represents the request body for creating a comment
type CreateCommentRequest struct {
	PostID string `json:"post_id"`
	Text   string `json:"text"`
}

// CommentResponse represents a comment in the response
type CommentResponse struct {
	ID        string `json:"id"`
	PostID    string `json:"post_id"`
	UserID    string `json:"user_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateComment creates a new comment on a post
func (h *CommentHandler) CreateComment(c *fiber.Ctx) error {
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
	var req CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "invalid request body",
			"status": fiber.StatusBadRequest,
		})
	}

	// Validate post_id is provided
	if req.PostID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "post_id is required",
			"status": fiber.StatusBadRequest,
		})
	}

	// Convert post ID from string to ObjectID
	postObjID, err := primitive.ObjectIDFromHex(req.PostID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "invalid post id",
			"status": fiber.StatusBadRequest,
		})
	}

	// Validate text is provided
	if req.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "comment text is required",
			"status": fiber.StatusBadRequest,
		})
	}

	// Create comment via service
	comment, err := h.svc.CreateComment(c.Context(), postObjID, userObjID, req.Text)
	if err != nil {
		return h.handleCommentError(c, err)
	}

	// Convert to response
	response := h.toCommentResponse(comment)

	return c.Status(fiber.StatusCreated).JSON(map[string]interface{}{
		"data":    response,
		"message": "comment created successfully",
		"status":  fiber.StatusCreated,
	})
}

// GetCommentsByPostID retrieves all comments for a post
func (h *CommentHandler) GetCommentsByPostID(c *fiber.Ctx) error {
	// Extract post ID from URL parameter
	postID := c.Params("id")
	if postID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "post id is required",
			"status": fiber.StatusBadRequest,
		})
	}

	// Convert post ID from string to ObjectID
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "invalid post id",
			"status": fiber.StatusBadRequest,
		})
	}

	// Get comments via service
	comments, err := h.svc.GetCommentsByPostID(c.Context(), postObjID)
	if err != nil {
		return h.handleCommentError(c, err)
	}

	// Convert to responses
	var responses []*CommentResponse
	for _, comment := range comments {
		responses = append(responses, h.toCommentResponse(comment))
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    responses,
		"count":   len(responses),
		"message": "comments retrieved successfully",
		"status":  fiber.StatusOK,
	})
}

// DeleteComment deletes a comment (owner only)
func (h *CommentHandler) DeleteComment(c *fiber.Ctx) error {
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

	// Extract comment ID from URL parameter
	commentID := c.Params("id")
	if commentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "comment id is required",
			"status": fiber.StatusBadRequest,
		})
	}

	// Convert comment ID from string to ObjectID
	commentObjID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "invalid comment id",
			"status": fiber.StatusBadRequest,
		})
	}

	// Delete comment via service
	err = h.svc.DeleteComment(c.Context(), commentObjID, userObjID)
	if err != nil {
		return h.handleCommentError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    nil,
		"message": "comment deleted successfully",
		"status":  fiber.StatusOK,
	})
}

// Helper methods

// toCommentResponse converts a domain.Comment to a CommentResponse
func (h *CommentHandler) toCommentResponse(comment *domain.Comment) *CommentResponse {
	return &CommentResponse{
		ID:        comment.ID.Hex(),
		PostID:    comment.PostID.Hex(),
		UserID:    comment.UserID.Hex(),
		Text:      comment.Text,
		CreatedAt: comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: comment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// handleCommentError maps domain errors to HTTP status codes
func (h *CommentHandler) handleCommentError(c *fiber.Ctx, err error) error {
	if errors.Is(err, domain.ErrCommentNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(map[string]interface{}{
			"error":  "comment not found",
			"status": fiber.StatusNotFound,
		})
	}

	if errors.Is(err, domain.ErrUnauthorizedComment) {
		return c.Status(fiber.StatusForbidden).JSON(map[string]interface{}{
			"error":  "you are not authorized to perform this action on this comment",
			"status": fiber.StatusForbidden,
		})
	}

	if errors.Is(err, domain.ErrCommentTextRequired) {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "comment text is required",
			"status": fiber.StatusBadRequest,
		})
	}

	if errors.Is(err, domain.ErrCommentTextTooLong) {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "comment text exceeds maximum length of 1000 characters",
			"status": fiber.StatusBadRequest,
		})
	}

	if errors.Is(err, domain.ErrPostNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(map[string]interface{}{
			"error":  "post not found",
			"status": fiber.StatusNotFound,
		})
	}

	// Default to 500 for unexpected errors
	return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
		"error":  "internal server error",
		"status": fiber.StatusInternalServerError,
	})
}
