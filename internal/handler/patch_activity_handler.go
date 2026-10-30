package handler

import (
	"fmt"

	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PatchActivityHandler exposes HTTP endpoints for patch activity tracking.
type PatchActivityHandler struct {
	cooldownSvc service.ActivityCooldownService
}

// NewPatchActivityHandler creates a new PatchActivityHandler.
func NewPatchActivityHandler(cooldownSvc service.ActivityCooldownService) *PatchActivityHandler {
	return &PatchActivityHandler{
		cooldownSvc: cooldownSvc,
	}
}

// MarkPatchViewedRequest represents the request payload for marking a patch as viewed.
type MarkPatchViewedRequest struct {
	// No body required, patchId comes from URL
}

// MarkPatchViewedResponse represents the response after marking a patch as viewed.
type MarkPatchViewedResponse struct {
	Message string `json:"message"`
	PatchID string `json:"patch_id"`
	UserID  string `json:"user_id"`
}

// MarkPatchViewed handles POST /api/v1/patches/:patchId/viewed requests.
// Records that a user viewed a patch, resetting the activity notification cooldown.
// This should be called when a user navigates to or opens a patch.
func (h *PatchActivityHandler) MarkPatchViewed(c *fiber.Ctx) error {
	patchID := c.Params("patchId")
	pid, err := primitive.ObjectIDFromHex(patchID)
	if err != nil {
		return fmt.Errorf("invalid patch id: %w", err)
	}

	userID := c.Locals("user_id").(string)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	// Record patch view (updates last_checked_at in cooldown tracking)
	if err := h.cooldownSvc.RecordPatchViewed(c.UserContext(), uid, pid); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(MarkPatchViewedResponse{
		Message: "Patch marked as viewed, activity notifications will reset",
		PatchID: pid.Hex(),
		UserID:  uid.Hex(),
	})
}
