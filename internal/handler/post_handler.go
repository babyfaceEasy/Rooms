package handler

import (
	"errors"
	"mime/multipart"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/internal/domain"
	"temp_backend/internal/repository"
	"temp_backend/internal/service"
)

// CreatePostRequest represents the request to create a post
type CreatePostRequest struct {
	Text string `form:"text" json:"text"`
}

// PostResponse represents a post in the response
type PostResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Text      string `json:"text"`
	Image     *string `json:"image,omitempty"`
	Video     *string `json:"video,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// PostHandler handles post-related endpoints
type PostHandler struct {
	svc     service.PostService
	storage repository.ObjectStorage
}

// NewPostHandler creates a new post handler
func NewPostHandler(svc service.PostService, storage repository.ObjectStorage) *PostHandler {
	return &PostHandler{
		svc:     svc,
		storage: storage,
	}
}

// CreatePost creates a new post with optional file upload
func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
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

	// Parse form data
	text := c.FormValue("text")
	if text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "text is required",
			"status": fiber.StatusBadRequest,
		})
	}

	var imageURL, videoURL *string

	// Handle file uploads
	form, err := c.MultipartForm()
	if err != nil && err != multipart.ErrMessageTooLarge {
		// It's okay if there's no multipart form (text-only post)
	} else if err == nil {
		// Try to upload image
		if images := form.File["image"]; len(images) > 0 {
			image := images[0]
			file, err := image.Open()
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
					"error":  "failed to process image",
					"status": fiber.StatusBadRequest,
				})
			}
			defer file.Close()

			// Validate image type
			if !isValidImageType(image.Filename) {
				return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
					"error":  "invalid image type",
					"status": fiber.StatusBadRequest,
				})
			}

			// Upload to S3
			url, err := h.storage.PutObject(c.Context(), "posts/images/"+userID+"/"+image.Filename, file, image.Size, image.Header.Get("Content-Type"))
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
					"error":  "failed to upload image",
					"status": fiber.StatusInternalServerError,
				})
			}
			imageURL = &url
		}

		// Try to upload video
		if videos := form.File["video"]; len(videos) > 0 {
			video := videos[0]
			file, err := video.Open()
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
					"error":  "failed to process video",
					"status": fiber.StatusBadRequest,
				})
			}
			defer file.Close()

			// Validate video type
			if !isValidVideoType(video.Filename) {
				return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
					"error":  "invalid video type",
					"status": fiber.StatusBadRequest,
				})
			}

			// Upload to S3
			url, err := h.storage.PutObject(c.Context(), "posts/videos/"+userID+"/"+video.Filename, file, video.Size, video.Header.Get("Content-Type"))
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
					"error":  "failed to upload video",
					"status": fiber.StatusInternalServerError,
				})
			}
			videoURL = &url
		}
	}

	// Create post via service
	post, err := h.svc.CreatePost(c.Context(), text, userObjID, imageURL, videoURL)
	if err != nil {
		return h.handleError(c, err)
	}

	// Convert domain Post to PostResponse
	response := &PostResponse{
		ID:        post.ID.Hex(),
		UserID:    post.UserID.Hex(),
		Text:      post.Text,
		Image:     post.Image,
		Video:     post.Video,
		CreatedAt: post.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: post.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return c.Status(fiber.StatusCreated).JSON(map[string]interface{}{
		"data":    response,
		"message": "post created successfully",
		"status":  fiber.StatusCreated,
	})
}

// GetPost retrieves a post by ID
func (h *PostHandler) GetPost(c *fiber.Ctx) error {
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

	// Get post via service
	post, err := h.svc.GetPost(c.Context(), postObjID)
	if err != nil {
		return h.handleError(c, err)
	}

	// Convert domain Post to PostResponse
	response := &PostResponse{
		ID:        post.ID.Hex(),
		UserID:    post.UserID.Hex(),
		Text:      post.Text,
		Image:     post.Image,
		Video:     post.Video,
		CreatedAt: post.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: post.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":   response,
		"status": fiber.StatusOK,
	})
}

// DeletePost deletes a post (owner only)
func (h *PostHandler) DeletePost(c *fiber.Ctx) error {
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

	// Delete post via service
	if err := h.svc.DeletePost(c.Context(), postObjID, userObjID); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    nil,
		"message": "post deleted successfully",
		"status":  fiber.StatusOK,
	})
}

// handleError maps service errors to HTTP responses
func (h *PostHandler) handleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrPostNotFound):
		return c.Status(fiber.StatusNotFound).JSON(map[string]interface{}{
			"error":  "post not found",
			"status": fiber.StatusNotFound,
		})
	case errors.Is(err, domain.ErrPostTextRequired):
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "post text is required",
			"status": fiber.StatusBadRequest,
		})
	case errors.Is(err, domain.ErrPostTextTooLong):
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error":  "post text exceeds maximum length (5000 characters)",
			"status": fiber.StatusBadRequest,
		})
	case errors.Is(err, domain.ErrUnauthorizedPost):
		return c.Status(fiber.StatusForbidden).JSON(map[string]interface{}{
			"error":  "you are not authorized to perform this action",
			"status": fiber.StatusForbidden,
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
			"error":  "internal server error",
			"status": fiber.StatusInternalServerError,
		})
	}
}

// isValidImageType checks if the file is a valid image type
func isValidImageType(filename string) bool {
	ext := filepath.Ext(filename)
	validTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	return validTypes[ext]
}

// isValidVideoType checks if the file is a valid video type
func isValidVideoType(filename string) bool {
	ext := filepath.Ext(filename)
	validTypes := map[string]bool{
		".mp4":  true,
		".webm": true,
		".mov":  true,
		".avi":  true,
	}
	return validTypes[ext]
}
