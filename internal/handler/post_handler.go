package handler

import (
	"mime/multipart"
	"path/filepath"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"
	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreatePostRequest represents the request to create a post
type CreatePostRequest struct {
	Text string `form:"text" json:"text"`
}

// PostResponse represents a post in the response
type PostResponse struct {
	ID               string  `json:"id"`
	RoomID           string  `json:"room_id"`
	RoomCode         string  `json:"room_code"`
	RoomName         string  `json:"room_name"`
	UserID           string  `json:"user_id"`
	UserName         string  `json:"user_name"`
	Text             string  `json:"text"`
	Image            *string `json:"image,omitempty"`
	Video            *string `json:"video,omitempty"`
	Audio            *string `json:"audio,omitempty"`
	ValidationsCount int     `json:"validations_count"`
	HasValidated     bool    `json:"has_validated"`
	RespectsCount    int     `json:"respects_count"`
	HasRespected     bool    `json:"has_respected"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// PostHandler handles post-related endpoints
type PostHandler struct {
	svc        service.PostService
	storage    repository.ObjectStorage
	roomRepo   repository.RoomRepository
	userRepo   repository.UserRepository
	sseManager *service.SSEManager
}

// NewPostHandler creates a new post handler
func NewPostHandler(svc service.PostService, storage repository.ObjectStorage, roomRepo repository.RoomRepository, userRepo repository.UserRepository, sseManager *service.SSEManager) *PostHandler {
	return &PostHandler{
		svc:        svc,
		storage:    storage,
		roomRepo:   roomRepo,
		userRepo:   userRepo,
		sseManager: sseManager,
	}
}

// CreatePost creates a new post with optional file upload
func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
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

	// Parse form data
	text := c.FormValue("text")
	if text == "" {
		return domain.ErrPostTextRequired
	}

	// Parse room_code from form data
	roomCode := c.FormValue("room_code")
	if roomCode == "" {
		return &domain.AppError{Code: "ROOM_CODE_REQUIRED", Message: "room_code is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get room by code
	room, err := h.roomRepo.GetByCode(c.Context(), roomCode)
	if err != nil {
		return err
	}

	// Verify user is room member
	isMember, err := h.roomRepo.IsUserMember(c.Context(), room.ID, userObjID)
	if err != nil {
		return domain.ErrInternalServer
	}
	if !isMember {
		return domain.ErrNotRoomMember
	}

	var imageURL, videoURL, audioURL *string

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
				return &domain.AppError{Code: "MEDIA_PROCESSING_FAILED", Message: "Failed to process image", HTTPStatus: fiber.StatusBadRequest}
			}
			defer file.Close()

			// Validate image type
			if !isValidImageType(image.Filename) {
				return &domain.AppError{Code: "INVALID_MEDIA_TYPE", Message: "Invalid image type. Allowed: jpg, jpeg, png, gif, webp", HTTPStatus: fiber.StatusBadRequest}
			}

			// Upload to S3
			url, err := h.storage.PutObject(c.Context(), "posts/images/"+userID+"/"+image.Filename, file, image.Size, image.Header.Get("Content-Type"))
			if err != nil {
				return &domain.AppError{Code: "MEDIA_UPLOAD_FAILED", Message: "Failed to upload image", HTTPStatus: fiber.StatusInternalServerError}
			}
			imageURL = &url
		}

		// Try to upload video
		if videos := form.File["video"]; len(videos) > 0 {
			video := videos[0]
			file, err := video.Open()
			if err != nil {
				return &domain.AppError{Code: "MEDIA_PROCESSING_FAILED", Message: "Failed to process video", HTTPStatus: fiber.StatusBadRequest}
			}
			defer file.Close()

			// Validate video type
			if !isValidVideoType(video.Filename) {
				return &domain.AppError{Code: "INVALID_MEDIA_TYPE", Message: "Invalid video type. Allowed: mp4, webm, mov, avi", HTTPStatus: fiber.StatusBadRequest}
			}

			// Upload to S3
			url, err := h.storage.PutObject(c.Context(), "posts/videos/"+userID+"/"+video.Filename, file, video.Size, video.Header.Get("Content-Type"))
			if err != nil {
				return &domain.AppError{Code: "MEDIA_UPLOAD_FAILED", Message: "Failed to upload video", HTTPStatus: fiber.StatusInternalServerError}
			}
			videoURL = &url
		}

		// Try to upload audio
		if audios := form.File["audio"]; len(audios) > 0 {
			audio := audios[0]
			file, err := audio.Open()
			if err != nil {
				return &domain.AppError{Code: "MEDIA_PROCESSING_FAILED", Message: "Failed to process audio", HTTPStatus: fiber.StatusBadRequest}
			}
			defer file.Close()

			// Validate audio type
			if !isValidAudioType(audio.Filename) {
				return &domain.AppError{Code: "INVALID_MEDIA_TYPE", Message: "Invalid audio type. Allowed: mp3, wav, m4a, aac, flac, ogg", HTTPStatus: fiber.StatusBadRequest}
			}

			// Upload to S3
			url, err := h.storage.PutObject(c.Context(), "posts/audio/"+userID+"/"+audio.Filename, file, audio.Size, audio.Header.Get("Content-Type"))
			if err != nil {
				return &domain.AppError{Code: "MEDIA_UPLOAD_FAILED", Message: "Failed to upload audio", HTTPStatus: fiber.StatusInternalServerError}
			}
			audioURL = &url
		}
	}

	// Create post via service with roomID
	post, err := h.svc.CreatePost(c.Context(), text, userObjID, room.ID, imageURL, videoURL, audioURL)
	if err != nil {
		return err
	}

	// Look up user name for the response
	user, err := h.userRepo.GetByID(c.Context(), userObjID)
	var userName string
	if err == nil && user != nil {
		userName = user.Name
	}

	// Convert domain Post to PostResponse using helper
	response := h.toPostResponse(post, room, userName, userObjID)

	return c.Status(fiber.StatusCreated).JSON(map[string]interface{}{
		"data":    response,
		"message": "post created successfully",
		"status":  fiber.StatusCreated,
	})
}

// GetPost retrieves a post by ID
func (h *PostHandler) GetPost(c *fiber.Ctx) error {
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

	postID := c.Params("id")
	if postID == "" {
		return &domain.AppError{Code: "POST_ID_REQUIRED", Message: "Post ID is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Convert post ID from string to ObjectID
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_POST_ID", Message: "Invalid post ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get post via service
	post, err := h.svc.GetPost(c.Context(), postObjID)
	if err != nil {
		return err
	}

	// Verify user is member of the room
	isMember, err := h.roomRepo.IsUserMember(c.Context(), post.RoomID, userObjID)
	if err != nil {
		return domain.ErrInternalServer
	}
	if !isMember {
		return domain.ErrNotRoomMember
	}

	// Get room to enrich response
	room, err := h.roomRepo.GetByID(c.Context(), post.RoomID)
	if err != nil {
		return err
	}

	// Look up user name for the response
	postUser, err := h.userRepo.GetByID(c.Context(), post.UserID)
	var userName string
	if err == nil && postUser != nil {
		userName = postUser.Name
	}

	// Convert domain Post to PostResponse using helper
	response := h.toPostResponse(post, room, userName, userObjID)

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":   response,
		"status": fiber.StatusOK,
	})
}

// GetPostsByRoomCode retrieves paginated posts for a room (room code in URL parameter)
func (h *PostHandler) GetPostsByRoomCode(c *fiber.Ctx) error {
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

	// Get room code from URL parameter
	roomCode := c.Params("code")
	if roomCode == "" {
		return &domain.AppError{Code: "ROOM_CODE_REQUIRED", Message: "room_code is required", HTTPStatus: fiber.StatusBadRequest}
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

	// Get room by code to verify it exists and user is a member
	room, err := h.roomRepo.GetByCode(c.Context(), roomCode)
	if err != nil {
		return err
	}

	// Verify user is room member
	isMember, err := h.roomRepo.IsUserMember(c.Context(), room.ID, userObjID)
	if err != nil {
		return domain.ErrInternalServer
	}
	if !isMember {
		return domain.ErrNotRoomMember
	}

	// Get paginated posts for the room via service
	posts, total, err := h.svc.GetPostsByRoomID(c.Context(), room.ID, page, limit)
	if err != nil {
		return err
	}

	// Batch fetch user names for all posts
	var userIDs []primitive.ObjectID
	seen := make(map[string]bool)
	for _, post := range posts {
		uid := post.UserID.Hex()
		if !seen[uid] {
			seen[uid] = true
			userIDs = append(userIDs, post.UserID)
		}
	}
	userMap := make(map[string]string)
	if len(userIDs) > 0 {
		users, err := h.userRepo.GetByIDs(c.Context(), userIDs)
		if err == nil {
			for _, u := range users {
				userMap[u.ID.Hex()] = u.Name
			}
		}
	}

	// Convert domain Posts to PostResponses
	var responses []*PostResponse
	for _, post := range posts {
		responses = append(responses, h.toPostResponse(post, room, userMap[post.UserID.Hex()], userObjID))
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    responses,
		"count":   len(responses),
		"page":    page,
		"limit":   limit,
		"total":   total,
		"message": "posts retrieved successfully",
		"status":  fiber.StatusOK,
	})
}

// DeletePost deletes a post (owner only)
func (h *PostHandler) DeletePost(c *fiber.Ctx) error {
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

	postID := c.Params("id")
	if postID == "" {
		return &domain.AppError{Code: "POST_ID_REQUIRED", Message: "Post ID is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Convert post ID from string to ObjectID
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_POST_ID", Message: "Invalid post ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get post to verify room membership
	post, err := h.svc.GetPost(c.Context(), postObjID)
	if err != nil {
		return err
	}

	// Verify user is member of the room (final check)
	isMember, err := h.roomRepo.IsUserMember(c.Context(), post.RoomID, userObjID)
	if err != nil {
		return domain.ErrInternalServer
	}
	if !isMember {
		return domain.ErrNotRoomMember
	}

	// Delete post via service
	if err := h.svc.DeletePost(c.Context(), postObjID, userObjID); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    nil,
		"message": "post deleted successfully",
		"status":  fiber.StatusOK,
	})
}

// ValidatePost marks a post as valid by the authenticated user.
func (h *PostHandler) ValidatePost(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return domain.ErrUnauthorized
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_USER_ID", Message: "Invalid user ID", HTTPStatus: fiber.StatusBadRequest}
	}

	postID := c.Params("id")
	if postID == "" {
		return &domain.AppError{Code: "POST_ID_REQUIRED", Message: "Post ID is required", HTTPStatus: fiber.StatusBadRequest}
	}

	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_POST_ID", Message: "Invalid post ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get post to verify room membership
	post, err := h.svc.GetPost(c.Context(), postObjID)
	if err != nil {
		return err
	}

	// Verify user is member of the room
	isMember, err := h.roomRepo.IsUserMember(c.Context(), post.RoomID, userObjID)
	if err != nil {
		return domain.ErrInternalServer
	}
	if !isMember {
		return domain.ErrNotRoomMember
	}

	// Add validation
	updatedPost, err := h.svc.ValidatePost(c.Context(), postObjID, userObjID)
	if err != nil {
		return err
	}

	// Look up user name for the response
	postUser, err := h.userRepo.GetByID(c.Context(), updatedPost.UserID)
	var userName string
	if err == nil && postUser != nil {
		userName = postUser.Name
	}

	room, err := h.roomRepo.GetByID(c.Context(), updatedPost.RoomID)
	if err != nil {
		return err
	}

	response := h.toPostResponse(updatedPost, room, userName, userObjID)
	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    response,
		"message": "post validated successfully",
		"status":  fiber.StatusOK,
	})
}

// RemoveValidation removes the authenticated user's validation from a post.
func (h *PostHandler) RemoveValidation(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return domain.ErrUnauthorized
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_USER_ID", Message: "Invalid user ID", HTTPStatus: fiber.StatusBadRequest}
	}

	postID := c.Params("id")
	if postID == "" {
		return &domain.AppError{Code: "POST_ID_REQUIRED", Message: "Post ID is required", HTTPStatus: fiber.StatusBadRequest}
	}

	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_POST_ID", Message: "Invalid post ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get post to verify room membership
	post, err := h.svc.GetPost(c.Context(), postObjID)
	if err != nil {
		return err
	}

	// Verify user is member of the room
	isMember, err := h.roomRepo.IsUserMember(c.Context(), post.RoomID, userObjID)
	if err != nil {
		return domain.ErrInternalServer
	}
	if !isMember {
		return domain.ErrNotRoomMember
	}

	// Remove validation
	if err := h.svc.RemoveValidation(c.Context(), postObjID, userObjID); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    nil,
		"message": "validation removed successfully",
		"status":  fiber.StatusOK,
	})
}

// RespectPost marks a post as respected by the authenticated user.
func (h *PostHandler) RespectPost(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return domain.ErrUnauthorized
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_USER_ID", Message: "Invalid user ID", HTTPStatus: fiber.StatusBadRequest}
	}

	postID := c.Params("id")
	if postID == "" {
		return &domain.AppError{Code: "POST_ID_REQUIRED", Message: "Post ID is required", HTTPStatus: fiber.StatusBadRequest}
	}

	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_POST_ID", Message: "Invalid post ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get post to verify room membership
	post, err := h.svc.GetPost(c.Context(), postObjID)
	if err != nil {
		return err
	}

	// Verify user is member of the room
	isMember, err := h.roomRepo.IsUserMember(c.Context(), post.RoomID, userObjID)
	if err != nil {
		return domain.ErrInternalServer
	}
	if !isMember {
		return domain.ErrNotRoomMember
	}

	// Add respect
	updatedPost, err := h.svc.RespectPost(c.Context(), postObjID, userObjID)
	if err != nil {
		return err
	}

	// Look up user name for the response
	postUser, err := h.userRepo.GetByID(c.Context(), updatedPost.UserID)
	var userName string
	if err == nil && postUser != nil {
		userName = postUser.Name
	}

	room, err := h.roomRepo.GetByID(c.Context(), updatedPost.RoomID)
	if err != nil {
		return err
	}

	response := h.toPostResponse(updatedPost, room, userName, userObjID)
	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    response,
		"message": "post respected successfully",
		"status":  fiber.StatusOK,
	})
}

// RemoveRespect removes the authenticated user's respect from a post.
func (h *PostHandler) RemoveRespect(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return domain.ErrUnauthorized
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_USER_ID", Message: "Invalid user ID", HTTPStatus: fiber.StatusBadRequest}
	}

	postID := c.Params("id")
	if postID == "" {
		return &domain.AppError{Code: "POST_ID_REQUIRED", Message: "Post ID is required", HTTPStatus: fiber.StatusBadRequest}
	}

	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_POST_ID", Message: "Invalid post ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get post to verify room membership
	post, err := h.svc.GetPost(c.Context(), postObjID)
	if err != nil {
		return err
	}

	// Verify user is member of the room
	isMember, err := h.roomRepo.IsUserMember(c.Context(), post.RoomID, userObjID)
	if err != nil {
		return domain.ErrInternalServer
	}
	if !isMember {
		return domain.ErrNotRoomMember
	}

	// Remove respect
	if err := h.svc.RemoveRespect(c.Context(), postObjID, userObjID); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    nil,
		"message": "respect removed successfully",
		"status":  fiber.StatusOK,
	})
}

// handleError delegates to the global error handler.
func (h *PostHandler) handleError(c *fiber.Ctx, err error) error {
	return err
}

// toPostResponse converts a domain Post and Room to a PostResponse
func (h *PostHandler) toPostResponse(post *domain.Post, room *domain.Room, userName string, currentUserID primitive.ObjectID) *PostResponse {
	return &PostResponse{
		ID:               post.ID.Hex(),
		RoomID:           post.RoomID.Hex(),
		RoomCode:         room.Code,
		RoomName:         room.Name,
		UserID:           post.UserID.Hex(),
		UserName:         userName,
		Text:             post.Text,
		Image:            post.Image,
		Video:            post.Video,
		Audio:            post.Audio,
		ValidationsCount: len(post.Validations),
		HasValidated:     containsObjectID(post.Validations, currentUserID),
		RespectsCount:    len(post.Respects),
		HasRespected:     containsObjectID(post.Respects, currentUserID),
		CreatedAt:        post.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        post.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// containsObjectID checks if a target ObjectID exists in a slice.
func containsObjectID(slice []primitive.ObjectID, target primitive.ObjectID) bool {
	for _, id := range slice {
		if id == target {
			return true
		}
	}
	return false
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

// isValidAudioType checks if the file is a valid audio type
func isValidAudioType(filename string) bool {
	ext := filepath.Ext(filename)
	validTypes := map[string]bool{
		".mp3":  true,
		".wav":  true,
		".m4a":  true,
		".aac":  true,
		".flac": true,
		".ogg":  true,
	}
	return validTypes[ext]
}

// StreamNewPosts streams new post events for a room using Server-Sent Events
func (h *PostHandler) StreamNewPosts(c *fiber.Ctx) error {
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

	// Get room code from query parameter
	roomCode := c.Query("room_code")
	if roomCode == "" {
		return &domain.AppError{Code: "ROOM_CODE_REQUIRED", Message: "room_code query parameter is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get room by code
	room, err := h.roomRepo.GetByCode(c.Context(), roomCode)
	if err != nil {
		return err
	}

	// Verify user is room member
	isMember, err := h.roomRepo.IsUserMember(c.Context(), room.ID, userObjID)
	if err != nil {
		return domain.ErrInternalServer
	}
	if !isMember {
		return domain.ErrNotRoomMember
	}

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	// Subscribe to events for this room
	events, subID := h.sseManager.Subscribe(room.ID.Hex())
	defer h.sseManager.Unsubscribe(room.ID.Hex(), subID)

	// Stream events to client
	for {
		select {
		case event, ok := <-events:
			if !ok {
				return nil
			}
			// Send event as SSE format
			if err := c.JSON(event); err != nil {
				return err
			}
		case <-c.Context().Done():
			return nil
		}
	}
}
