package handler

import (
	"errors"
	"fmt"

	"temp_backend/internal/domain"
	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeviceHandler exposes HTTP endpoints for managing device tokens and notifications.
type DeviceHandler struct {
	deviceTokenSvc service.DeviceTokenService
	notificationSvc service.NotificationService
}

// NewDeviceHandler creates a new DeviceHandler.
func NewDeviceHandler(
	deviceTokenSvc service.DeviceTokenService,
	notificationSvc service.NotificationService,
) *DeviceHandler {
	return &DeviceHandler{
		deviceTokenSvc: deviceTokenSvc,
		notificationSvc: notificationSvc,
	}
}

// RegisterDeviceRequest represents the request payload for registering a device token.
type RegisterDeviceRequest struct {
	Token    string `json:"token" form:"token"`
	Platform string `json:"platform" form:"platform"` // "ios", "android", "web"
}

// RegisterDeviceResponse represents the response after registering a device.
type RegisterDeviceResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Token     string `json:"token"`
	Platform  string `json:"platform"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// RegisterDevice handles POST /api/v1/devices/register requests.
// Registers a new device token for the authenticated user.
func (h *DeviceHandler) RegisterDevice(c *fiber.Ctx) error {
	var req RegisterDeviceRequest

	if err := c.BodyParser(&req); err != nil {
		return fmt.Errorf("parse register device request: %w", err)
	}

	// Get user ID from middleware
	userID := c.Locals("user_id").(string)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	// Validate platform
	if req.Platform != "ios" && req.Platform != "android" && req.Platform != "web" {
		return domain.ErrInvalidInput
	}

	// Validate token
	if req.Token == "" {
		return domain.ErrInvalidInput
	}

	// Register device token
	deviceToken, err := h.deviceTokenSvc.RegisterDevice(
		c.UserContext(),
		uid,
		req.Token,
		domain.DeviceTokenPlatform(req.Platform),
	)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(RegisterDeviceResponse{
		ID:        deviceToken.ID.Hex(),
		UserID:    deviceToken.UserID.Hex(),
		Token:     deviceToken.Token,
		Platform:  string(deviceToken.Platform),
		IsActive:  deviceToken.IsActive,
		CreatedAt: deviceToken.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UnregisterDevice handles DELETE /api/v1/devices/:tokenId requests.
// Removes a device token for the authenticated user.
func (h *DeviceHandler) UnregisterDevice(c *fiber.Ctx) error {
	tokenID := c.Params("tokenId")
	tid, err := primitive.ObjectIDFromHex(tokenID)
	if err != nil {
		return fmt.Errorf("invalid token id: %w", err)
	}

	userID := c.Locals("user_id").(string)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	// Verify ownership
	deviceToken, err := h.deviceTokenSvc.GetDeviceToken(c.UserContext(), tid)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return err
		}
		return err
	}

	if deviceToken.UserID != uid {
		return domain.ErrForbidden
	}

	// Delete the device token
	if err := h.deviceTokenSvc.UnregisterDevice(c.UserContext(), tid); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetDevicesResponse represents a list of devices for the user.
type GetDevicesResponse struct {
	Devices []RegisterDeviceResponse `json:"devices"`
}

// GetDevices handles GET /api/v1/devices requests.
// Lists all active device tokens for the authenticated user.
func (h *DeviceHandler) GetDevices(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	devices, err := h.deviceTokenSvc.GetActiveDeviceTokens(c.UserContext(), uid)
	if err != nil {
		return err
	}

	response := GetDevicesResponse{
		Devices: make([]RegisterDeviceResponse, len(devices)),
	}

	for i, device := range devices {
		response.Devices[i] = RegisterDeviceResponse{
			ID:        device.ID.Hex(),
			UserID:    device.UserID.Hex(),
			Token:     device.Token,
			Platform:  string(device.Platform),
			IsActive:  device.IsActive,
			CreatedAt: device.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return c.JSON(response)
}

// GetNotificationHistoryResponse represents a single notification in history.
type GetNotificationHistoryResponse struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	PatchID        string `json:"patch_id"`
	Type           string `json:"type"` // "added", "joined", "activity"
	Message        string `json:"message"`
	Status         string `json:"status"` // "pending", "sent", "failed"
	SentAt         *string `json:"sent_at,omitempty"`
	ErrorMessage   *string `json:"error_message,omitempty"`
	CreatedAt      string `json:"created_at"`
}

// NotificationHistoryResponse represents a list of notifications.
type NotificationHistoryResponse struct {
	Notifications []GetNotificationHistoryResponse `json:"notifications"`
}

// GetNotificationHistory handles GET /api/v1/notifications/history requests.
// Lists notification delivery history for the authenticated user.
func (h *DeviceHandler) GetNotificationHistory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	notifications, err := h.notificationSvc.GetUserNotifications(c.UserContext(), uid)
	if err != nil {
		return err
	}

	response := NotificationHistoryResponse{
		Notifications: make([]GetNotificationHistoryResponse, len(notifications)),
	}

	for i, notif := range notifications {
		var sentAt *string
		if notif.SentAt != nil {
			t := notif.SentAt.Format("2006-01-02T15:04:05Z07:00")
			sentAt = &t
		}

		var errorMsg *string
		if notif.ErrorMessage != "" {
			errorMsg = &notif.ErrorMessage
		}

		patchID := ""
		if notif.PatchID != primitive.NilObjectID {
			patchID = notif.PatchID.Hex()
		}

		response.Notifications[i] = GetNotificationHistoryResponse{
			ID:           notif.ID.Hex(),
			UserID:       notif.UserID.Hex(),
			PatchID:      patchID,
			Type:         string(notif.NotificationType),
			Message:      notif.Message,
			Status:       string(notif.Status),
			SentAt:       sentAt,
			ErrorMessage: errorMsg,
			CreatedAt:    notif.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return c.JSON(response)
}
