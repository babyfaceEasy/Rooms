package handler

import (
	"temp_backend/internal/domain"
	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateRoomRequest represents the request to create a room
type CreateRoomRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// AddUserToRoomRequest represents the request to add a user to a room
type AddUserToRoomRequest struct {
	Code string `json:"code"`
}

// AddUserToRoomByUserCodeRequest represents the request to add a user by their customer code
type AddUserToRoomByUserCodeRequest struct {
	RoomCode string `json:"room_code"`
	UserCode string `json:"user_code"`
}

// RemoveUserFromRoomByUserCodeRequest represents the request to remove a user by their customer code
type RemoveUserFromRoomByUserCodeRequest struct {
	RoomCode string `json:"room_code"`
	UserCode string `json:"user_code"`
}

// RemoveMemberFromRoomRequest represents the request to remove a member from a room
type RemoveMemberFromRoomRequest struct {
	MemberID string `json:"member_id"`
}

// RoomResponse represents a room in the response
type RoomResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Code      string   `json:"code"`
	CreatedBy string   `json:"created_by"`
	Members   []string `json:"members,omitempty"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// RoomMembersResponse represents the members of a room
type RoomMembersResponse struct {
	RoomID   string   `json:"room_id"`
	RoomCode string   `json:"room_code"`
	RoomName string   `json:"room_name"`
	Owner    string   `json:"owner"`
	Members  []string `json:"members"`
	Count    int      `json:"count"`
}

// RoomHandler handles room-related endpoints
type RoomHandler struct {
	svc service.RoomService
}

// UserDetailResponse represents user details in the response
type UserDetailResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	IsAgeVerified bool   `json:"is_age_verified"`
	Creator       bool   `json:"creator"`
	CreatedAt     string `json:"created_at"`
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
		return domain.ErrUnauthorized
	}

	// Convert user ID from string to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_USER_ID", Message: "Invalid user ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Parse request body
	var req CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.ErrInvalidInput
	}

	// Create room via service
	room, err := h.svc.CreateRoom(c.UserContext(), req.Name, req.Code, userObjID)
	if err != nil {
		return err
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

// AddUserToRoom adds a user to an existing room by code
func (h *RoomHandler) AddUserToRoom(c *fiber.Ctx) error {
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
	var req AddUserToRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.ErrInvalidInput
	}

	// Get room first to check if user is the creator or already a member
	room, err := h.svc.GetRoom(c.UserContext(), req.Code)
	if err != nil {
		return err
	}

	// Check if the user is the creator
	if room.CreatedBy == userObjID {
		return domain.ErrCannotJoinOwnRoom
	}

	// Check if user is already a member
	isAlreadyMember := false
	for _, m := range room.Members {
		if m == userObjID {
			isAlreadyMember = true
			break
		}
	}

	// If already a member, return success without calling the service
	if isAlreadyMember {
		// Convert members to string slice for response
		members := make([]string, len(room.Members))
		for i, m := range room.Members {
			members[i] = m.Hex()
		}

		response := &RoomResponse{
			ID:        room.ID.Hex(),
			Name:      room.Name,
			Code:      room.Code,
			CreatedBy: room.CreatedBy.Hex(),
			Members:   members,
			CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: room.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
			"data":    response,
			"message": "you are already a member of this room",
			"status":  fiber.StatusOK,
		})
	}

	// Add user to room via service
	room, err = h.svc.AddUserToRoom(c.UserContext(), req.Code, userObjID)
	if err != nil {
		return err
	}

	// Convert members to string slice for response
	members := make([]string, len(room.Members))
	for i, m := range room.Members {
		members[i] = m.Hex()
	}

	// Convert domain Room to RoomResponse
	response := &RoomResponse{
		ID:        room.ID.Hex(),
		Name:      room.Name,
		Code:      room.Code,
		CreatedBy: room.CreatedBy.Hex(),
		Members:   members,
		CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: room.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    response,
		"message": "user added to room successfully",
		"status":  fiber.StatusOK,
	})
}

// GetRoom retrieves room details - only owner and members can access
func (h *RoomHandler) GetRoom(c *fiber.Ctx) error {
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
		return &domain.AppError{Code: "ROOM_CODE_REQUIRED", Message: "Room code is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get room via service
	room, err := h.svc.GetRoom(c.UserContext(), roomCode)
	if err != nil {
		return err
	}

	// Check authorization: user must be owner or member
	isOwner := room.CreatedBy == userObjID
	isMember := false
	for _, m := range room.Members {
		if m == userObjID {
			isMember = true
			break
		}
	}

	if !isOwner && !isMember {
		return domain.ErrForbidden
	}

	// Convert members to string slice for response
	members := make([]string, len(room.Members))
	for i, m := range room.Members {
		members[i] = m.Hex()
	}

	// Convert domain Room to RoomResponse
	response := &RoomResponse{
		ID:        room.ID.Hex(),
		Name:      room.Name,
		Code:      room.Code,
		CreatedBy: room.CreatedBy.Hex(),
		Members:   members,
		CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: room.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    response,
		"message": "room details retrieved successfully",
		"status":  fiber.StatusOK,
	})
}

// GetRoomByID retrieves room details by ID - only owner and members can access
func (h *RoomHandler) GetRoomByID(c *fiber.Ctx) error {
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

	// Get room ID from URL parameter
	roomIDStr := c.Params("id")
	if roomIDStr == "" {
		return &domain.AppError{Code: "ROOM_ID_REQUIRED", Message: "Room ID is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Convert room ID from string to ObjectID
	roomID, err := primitive.ObjectIDFromHex(roomIDStr)
	if err != nil {
		return &domain.AppError{Code: "INVALID_ROOM_ID", Message: "Invalid room ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get room via service
	room, err := h.svc.GetRoomByID(c.UserContext(), roomID)
	if err != nil {
		return err
	}

	// Check authorization: user must be owner or member
	isOwner := room.CreatedBy == userObjID
	isMember := false
	for _, m := range room.Members {
		if m == userObjID {
			isMember = true
			break
		}
	}

	if !isOwner && !isMember {
		return domain.ErrForbidden
	}

	// Convert members to string slice for response
	members := make([]string, len(room.Members))
	for i, m := range room.Members {
		members[i] = m.Hex()
	}

	// Convert domain Room to RoomResponse
	response := &RoomResponse{
		ID:        room.ID.Hex(),
		Name:      room.Name,
		Code:      room.Code,
		CreatedBy: room.CreatedBy.Hex(),
		Members:   members,
		CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: room.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    response,
		"message": "room details retrieved successfully",
		"status":  fiber.StatusOK,
	})
}

// GetRoomMembers retrieves members of a room - only owner and members can access
func (h *RoomHandler) GetRoomMembers(c *fiber.Ctx) error {
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
		return &domain.AppError{Code: "ROOM_CODE_REQUIRED", Message: "Room code is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get room via service
	room, err := h.svc.GetRoom(c.UserContext(), roomCode)
	if err != nil {
		return err
	}

	// Check authorization: user must be owner or member
	isOwner := room.CreatedBy == userObjID
	isMember := false
	for _, m := range room.Members {
		if m == userObjID {
			isMember = true
			break
		}
	}

	if !isOwner && !isMember {
		return domain.ErrForbidden
	}

	// Convert members to string slice for response
	members := make([]string, len(room.Members))
	for i, m := range room.Members {
		members[i] = m.Hex()
	}

	// Create response
	response := &RoomMembersResponse{
		RoomID:   room.ID.Hex(),
		RoomCode: room.Code,
		RoomName: room.Name,
		Owner:    room.CreatedBy.Hex(),
		Members:  members,
		Count:    len(room.Members),
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    response,
		"message": "room members retrieved successfully",
		"status":  fiber.StatusOK,
	})
}

// LeaveRoom removes the authenticated user from a room
func (h *RoomHandler) LeaveRoom(c *fiber.Ctx) error {
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
		return &domain.AppError{Code: "ROOM_CODE_REQUIRED", Message: "Room code is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get room to provide context-specific error messages
	room, err := h.svc.GetRoom(c.UserContext(), roomCode)
	if err != nil {
		return err
	}

	// Check if user is the owner
	if room.CreatedBy == userObjID {
		return domain.ErrOwnerCannotLeaveRoom
	}

	// Leave room via service
	if err := h.svc.LeaveRoom(c.UserContext(), roomCode, userObjID); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    nil,
		"message": "you have left the room successfully",
		"status":  fiber.StatusOK,
	})
}

// RemoveMemberFromRoom removes a member from the room (only owner can remove)
func (h *RoomHandler) RemoveMemberFromRoom(c *fiber.Ctx) error {
	// Extract user ID from context (owner/requester)
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return domain.ErrUnauthorized
	}

	// Convert user ID from string to ObjectID
	ownerObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_USER_ID", Message: "Invalid user ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get room code from URL parameter
	roomCode := c.Params("code")
	if roomCode == "" {
		return &domain.AppError{Code: "ROOM_CODE_REQUIRED", Message: "Room code is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Parse request body for member ID to remove
	var req RemoveMemberFromRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.ErrInvalidInput
	}

	// Convert member ID from string to ObjectID
	memberObjID, err := primitive.ObjectIDFromHex(req.MemberID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_MEMBER_ID", Message: "Invalid member ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Remove member from room via service
	if err := h.svc.RemoveMemberFromRoom(c.UserContext(), roomCode, ownerObjID, memberObjID); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    nil,
		"message": "member removed from room successfully",
		"status":  fiber.StatusOK,
	})
}

// HandleRoomDelete handles both delete room (owner) and leave room (member)
func (h *RoomHandler) HandleRoomDelete(c *fiber.Ctx) error {
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
		return &domain.AppError{Code: "ROOM_CODE_REQUIRED", Message: "Room code is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get room to check if user is owner
	room, err := h.svc.GetRoom(c.Context(), roomCode)
	if err != nil {
		return err
	}

	// If user is the owner, delete the room; otherwise, leave the room
	if room.CreatedBy == userObjID {
		if err := h.svc.DeleteRoom(c.Context(), roomCode, userObjID); err != nil {
			return err
		}
		return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
			"data":    nil,
			"message": "room deleted successfully",
			"status":  fiber.StatusOK,
		})
	}

	// User is a member, so leave the room
	if err := h.svc.LeaveRoom(c.Context(), roomCode, userObjID); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    nil,
		"message": "you have left the room successfully",
		"status":  fiber.StatusOK,
	})
}

// handleError is kept for call sites that wrap non-domain errors.
// It delegates to the global error handler by returning the error.
func (h *RoomHandler) handleError(c *fiber.Ctx, err error) error {
	return err
}

// AddUserToRoomByUserCode adds a user to a room using the user's customer code
func (h *RoomHandler) AddUserToRoomByUserCode(c *fiber.Ctx) error {
	// Parse request body
	var req AddUserToRoomByUserCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.ErrInvalidInput
	}

	// Validate request fields
	if req.RoomCode == "" || req.UserCode == "" {
		return &domain.AppError{Code: "INVALID_INPUT", Message: "room_code and user_code are required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Add user to room via service
	room, err := h.svc.AddUserToRoomByUserCode(c.Context(), req.RoomCode, req.UserCode)
	if err != nil {
		return err
	}

	// Convert members to string slice for response
	members := make([]string, len(room.Members))
	for i, m := range room.Members {
		members[i] = m.Hex()
	}

	// Convert domain Room to RoomResponse
	response := &RoomResponse{
		ID:        room.ID.Hex(),
		Name:      room.Name,
		Code:      room.Code,
		CreatedBy: room.CreatedBy.Hex(),
		Members:   members,
		CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: room.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    response,
		"message": "user added to room successfully",
		"status":  fiber.StatusOK,
	})
}

// RemoveUserFromRoomByUserCode removes a user from a room using the user's customer code (only owner can remove)
func (h *RoomHandler) RemoveUserFromRoomByUserCode(c *fiber.Ctx) error {
	// Extract user ID from context (owner/requester)
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return domain.ErrUnauthorized
	}

	// Convert user ID from string to ObjectID
	ownerObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &domain.AppError{Code: "INVALID_USER_ID", Message: "Invalid user ID", HTTPStatus: fiber.StatusBadRequest}
	}

	// Parse request body
	var req RemoveUserFromRoomByUserCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.ErrInvalidInput
	}

	// Validate request fields
	if req.RoomCode == "" || req.UserCode == "" {
		return &domain.AppError{Code: "INVALID_INPUT", Message: "room_code and user_code are required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Remove user from room via service
	if err := h.svc.RemoveUserFromRoomByUserCode(c.Context(), req.RoomCode, req.UserCode, ownerObjID); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    nil,
		"message": "user removed from room successfully",
		"status":  fiber.StatusOK,
	})
}

// GetRoomUsers retrieves all user details for members of a room - only owner and members can access
func (h *RoomHandler) GetRoomUsers(c *fiber.Ctx) error {
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
		return &domain.AppError{Code: "ROOM_CODE_REQUIRED", Message: "Room code is required", HTTPStatus: fiber.StatusBadRequest}
	}

	// Get room to verify access
	room, err := h.svc.GetRoom(c.Context(), roomCode)
	if err != nil {
		return err
	}

	// Check authorization: user must be owner or member
	isOwner := room.CreatedBy == userObjID
	isMember := false
	for _, m := range room.Members {
		if m == userObjID {
			isMember = true
			break
		}
	}

	if !isOwner && !isMember {
		return domain.ErrForbidden
	}

	// Get all users in the room
	users, err := h.svc.GetRoomUsers(c.Context(), roomCode)
	if err != nil {
		return err
	}

	// Convert domain User objects to UserDetailResponse objects
	responses := make([]UserDetailResponse, len(users))
	for i, user := range users {
		responses[i] = UserDetailResponse{
			ID:            user.ID.Hex(),
			Name:          user.Name,
			Email:         user.Email,
			IsAgeVerified: user.IsAgeVerified,
			Creator:       user.ID == room.CreatedBy,
			CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    responses,
		"count":   len(responses),
		"message": "room users retrieved successfully",
		"status":  fiber.StatusOK,
	})
}

// ListUserRooms retrieves all rooms where the authenticated user is the creator or a member
func (h *RoomHandler) ListUserRooms(c *fiber.Ctx) error {
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

	// Get user's rooms via service
	rooms, err := h.svc.ListUserRooms(c.Context(), userObjID)
	if err != nil {
		return err
	}

	// Convert domain Room objects to RoomResponse objects
	responses := make([]RoomResponse, len(rooms))
	for i, room := range rooms {
		// Convert member ObjectIDs to strings
		memberStrings := make([]string, len(room.Members))
		for j, member := range room.Members {
			memberStrings[j] = member.Hex()
		}

		responses[i] = RoomResponse{
			ID:        room.ID.Hex(),
			Name:      room.Name,
			Code:      room.Code,
			CreatedBy: room.CreatedBy.Hex(),
			Members:   memberStrings,
			CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: room.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":   responses,
		"count":  len(responses),
		"status": fiber.StatusOK,
	})
}
