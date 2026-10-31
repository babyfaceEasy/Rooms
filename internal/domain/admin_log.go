package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminLogAction represents an action logged in the admin log
type AdminLogAction string

const (
	// Auth actions
	AdminLogActionLogin AdminLogAction = "LOGIN"
	AdminLogActionLogout AdminLogAction = "LOGOUT"

	// User management actions
	AdminLogActionUserSuspend   AdminLogAction = "USER_SUSPEND"
	AdminLogActionUserUnsuspend AdminLogAction = "USER_UNSUSPEND"
	AdminLogActionUserDelete    AdminLogAction = "USER_DELETE"
	AdminLogActionUserUpdate    AdminLogAction = "USER_UPDATE"

	// Post management actions
	AdminLogActionPostDelete AdminLogAction = "POST_DELETE"
	AdminLogActionPostUpdate AdminLogAction = "POST_UPDATE"

	// Patch/Room management actions
	AdminLogActionPatchDelete AdminLogAction = "PATCH_DELETE"
	AdminLogActionPatchUpdate AdminLogAction = "PATCH_UPDATE"

	// Admin management actions
	AdminLogActionAdminCreate AdminLogAction = "ADMIN_CREATE"
	AdminLogActionAdminDelete AdminLogAction = "ADMIN_DELETE"
	AdminLogActionAdminUpdate AdminLogAction = "ADMIN_UPDATE"
)

// AdminLogResourceType represents the type of resource being acted upon
type AdminLogResourceType string

const (
	AdminLogResourceTypeUser  AdminLogResourceType = "USER"
	AdminLogResourceTypePost  AdminLogResourceType = "POST"
	AdminLogResourceTypePatch AdminLogResourceType = "PATCH"
	AdminLogResourceTypeAdmin AdminLogResourceType = "ADMIN"
)

// AdminLogStatus represents the status of a logged action
type AdminLogStatus string

const (
	AdminLogStatusSuccess AdminLogStatus = "SUCCESS"
	AdminLogStatusFailed  AdminLogStatus = "FAILED"
)

// AdminLog represents an audit log entry for admin actions
type AdminLog struct {
	ID             primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	AdminID        primitive.ObjectID   `json:"admin_id" bson:"admin_id"`
	AdminEmail     string               `json:"admin_email" bson:"admin_email"`
	Action         AdminLogAction       `json:"action" bson:"action"`
	ResourceType   AdminLogResourceType `json:"resource_type" bson:"resource_type"`
	ResourceID     *primitive.ObjectID  `json:"resource_id,omitempty" bson:"resource_id,omitempty"`
	Changes        map[string]interface{} `json:"changes,omitempty" bson:"changes,omitempty"` // before/after values
	Status         AdminLogStatus       `json:"status" bson:"status"`
	ErrorMessage   string               `json:"error_message,omitempty" bson:"error_message,omitempty"`
	IPAddress      string               `json:"ip_address,omitempty" bson:"ip_address,omitempty"`
	UserAgent      string               `json:"user_agent,omitempty" bson:"user_agent,omitempty"`
	CreatedAt      time.Time            `json:"created_at" bson:"created_at"`
}
