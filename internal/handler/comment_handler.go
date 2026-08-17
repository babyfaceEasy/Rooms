package handler

import (

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
		return domain.ErrUnauthorized
	}

	// Convert user ID from string to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_USER_ID", Message: "Invalid user ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Parse request body
	var req CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.ErrInvalidInput
	}

	// Validate post_id is provided
	if req.PostID == "" {
		return &domain.AppError{Code: "POST_ID_REQUIRED", Message: "Post ID is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Convert post ID from string to ObjectID
	postObjID, err := primitive.ObjectIDFromHex(req.PostID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_POST_ID", Message: "Invalid post ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Validate text is provided
	if req.Text == "" {
		return domain.ErrCommentTextRequired
	}

	// Create comment via service
	comment, err := h.svc.CreateComment(c.Context(), postObjID, userObjID, req.Text)
	if err != nil {
		return err
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
		return &domain.AppError{Code: "POST_ID_REQUIRED", Message: "Post ID is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Convert post ID from string to ObjectID
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_POST_ID", Message: "Invalid post ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get comments via service
	comments, err := h.svc.GetCommentsByPostID(c.Context(), postObjID)
	if err != nil {
		return err
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
		return domain.ErrUnauthorized
	}

	// Convert user ID from string to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_USER_ID", Message: "Invalid user ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Extract comment ID from URL parameter
	commentID := c.Params("id")
	if commentID == "" {
		return &domain.AppError{Code: "COMMENT_ID_REQUIRED", Message: "Comment ID is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Convert comment ID from string to ObjectID
	commentObjID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_COMMENT_ID", Message: "Invalid comment ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Delete comment via service
	err = h.svc.DeleteComment(c.Context(), commentObjID, userObjID)
	if err != nil {
		return err
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

// handleCommentError delegates to the global error handler.
func (h *CommentHandler) handleCommentError(c *fiber.Ctx, err error) error {
	return err
}
