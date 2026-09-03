package handler

import (
	"bufio"
	"encoding/json"
	"fmt"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"
	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CommentHandler handles HTTP requests for comments
type CommentHandler struct {
	svc        service.CommentService
	userRepo   repository.UserRepository
	postRepo   repository.PostRepository
	sseManager *service.SSEManager
}

// NewCommentHandler creates a new CommentHandler
func NewCommentHandler(svc service.CommentService, userRepo repository.UserRepository, postRepo repository.PostRepository, sseManager *service.SSEManager) *CommentHandler {
	return &CommentHandler{
		svc:        svc,
		userRepo:   userRepo,
		postRepo:   postRepo,
		sseManager: sseManager,
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
	UserName  string `json:"user_name"`
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
	comment, err := h.svc.CreateComment(c.UserContext(), postObjID, userObjID, req.Text)
	if err != nil {
		return err
	}

	// Look up user name for the response
	commentUser, err := h.userRepo.GetByID(c.UserContext(), userObjID)
	var userName string
	if err == nil && commentUser != nil {
		userName = commentUser.Name
	}

	// Convert to response
	response := h.toCommentResponse(comment, userName)

	// Publish SSE event to post subscribers
	h.sseManager.PublishCommentCreated(postObjID.Hex(), comment.ID.Hex(), response)

	return c.Status(fiber.StatusCreated).JSON(map[string]interface{}{
		"data":    response,
		"message": "comment created successfully",
		"status":  fiber.StatusCreated,
	})
}

// GetCommentsByPostID retrieves paginated comments for a post
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

	// Parse pagination query params
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Parse sort order (asc = oldest first, desc = newest first)
	sortOrder := c.Query("sort", "desc")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Get paginated comments via service
	comments, total, err := h.svc.GetCommentsByPostID(c.UserContext(), postObjID, page, limit, sortOrder)
	if err != nil {
		return err
	}

	// Batch fetch user names for all comments
	var userIDs []primitive.ObjectID
	seen := make(map[string]bool)
	for _, comment := range comments {
		uid := comment.UserID.Hex()
		if !seen[uid] {
			seen[uid] = true
			userIDs = append(userIDs, comment.UserID)
		}
	}
	userMap := make(map[string]string)
	if len(userIDs) > 0 {
		users, err := h.userRepo.GetByIDs(c.UserContext(), userIDs)
		if err == nil {
			for _, u := range users {
				userMap[u.ID.Hex()] = u.Name
			}
		}
	}

	// Convert to responses
	var responses []*CommentResponse
	for _, comment := range comments {
		responses = append(responses, h.toCommentResponse(comment, userMap[comment.UserID.Hex()]))
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    responses,
		"count":   len(responses),
		"page":    page,
		"limit":   limit,
		"total":   total,
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
	err = h.svc.DeleteComment(c.UserContext(), commentObjID, userObjID)
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
func (h *CommentHandler) toCommentResponse(comment *domain.Comment, userName string) *CommentResponse {
	return &CommentResponse{
		ID:        comment.ID.Hex(),
		PostID:    comment.PostID.Hex(),
		UserID:    comment.UserID.Hex(),
		UserName:  userName,
		Text:      comment.Text,
		CreatedAt: comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: comment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// StreamNewComments streams new comment events for a post using Server-Sent Events
func (h *CommentHandler) StreamNewComments(c *fiber.Ctx) error {
	// Extract user ID from context
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return domain.ErrUnauthorized
	}

	// Get post ID from URL parameter
	postID := c.Params("id")
	if postID == "" {
		return &domain.AppError{Code: "POST_ID_REQUIRED", Message: "Post ID is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Convert post ID from string to ObjectID
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_POST_ID", Message: "Invalid post ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get post to verify it exists
	post, err := h.postRepo.GetByID(c.UserContext(), postObjID)
	if err != nil {
		return err
	}

	// Verify user is member of the room
	// Note: We'll add roomRepo to CommentHandler later if needed for member verification
	// For now, comment stream is open to anyone with post access

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	// Subscribe to events for this post
	events, subID := h.sseManager.SubscribeToPost(post.ID.Hex())

	// Capture the disconnect context channel BEFORE moving into the stream writer
	notifyDone := c.Context().Done()

	// Execute streaming writer loop
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Ensure cleanup happens when this streaming function finishes
		defer h.sseManager.UnsubscribeFromPost(post.ID.Hex(), subID)

		for {
			select {
			case event, ok := <-events:
				if !ok {
					return
				}

				eventJSON, err := json.Marshal(event)
				if err != nil {
					return
				}

				// Write to bufio.Writer
				_, err = fmt.Fprintf(w, "data: %s\n\n", string(eventJSON))
				if err != nil {
					return
				}

				// CRITICAL: Force flush the data out over the network immediately
				if err := w.Flush(); err != nil {
					return
				}

			case <-notifyDone:
				return
			}
		}
	})
	return nil
}

// handleCommentError delegates to the global error handler.
func (h *CommentHandler) handleCommentError(c *fiber.Ctx, err error) error {
	return err
}
