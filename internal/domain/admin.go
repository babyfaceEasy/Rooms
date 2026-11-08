package domain

import (
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminRole represents the role of an admin user
type AdminRole string

const (
	AdminRoleSuperAdmin AdminRole = "super_admin"
	AdminRoleAdmin      AdminRole = "admin"
	AdminRoleModerator  AdminRole = "moderator"
)

// AdminPermission represents a specific permission an admin can have
type AdminPermission string

const (
	// Admin management
	PermissionManageAdmins AdminPermission = "manage_admins"

	// User management
	PermissionManageUsers       AdminPermission = "manage_users"
	PermissionSuspendUsers      AdminPermission = "suspend_users"
	PermissionDeleteUsers       AdminPermission = "delete_users"
	PermissionViewUserAnalytics AdminPermission = "view_user_analytics"

	// Post management
	PermissionManagePosts       AdminPermission = "manage_posts"
	PermissionDeletePosts       AdminPermission = "delete_posts"
	PermissionViewPostAnalytics AdminPermission = "view_post_analytics"

	// Patch/Room management
	PermissionManagePatches       AdminPermission = "manage_patches"
	PermissionDeletePatches       AdminPermission = "delete_patches"
	PermissionViewPatchAnalytics  AdminPermission = "view_patch_analytics"

	// Logs & Analytics
	PermissionViewLogs      AdminPermission = "view_logs"
	PermissionViewAnalytics AdminPermission = "view_analytics"
)

// Admin represents an administrator user
type Admin struct {
	ID                    primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Email                 string             `json:"email" bson:"email"`
	PasswordHash          string             `json:"-" bson:"password_hash"`
	Name                  string             `json:"name" bson:"name"`
	Role                  AdminRole          `json:"role" bson:"role"`
	Permissions           []AdminPermission  `json:"permissions" bson:"permissions"`
	IsActive              bool               `json:"is_active" bson:"is_active"`
	LastLogin             *time.Time         `json:"last_login,omitempty" bson:"last_login,omitempty"`
	FailedLoginAttempts   int                `json:"failed_login_attempts,omitempty" bson:"failed_login_attempts,omitempty"`
	LockedUntil           *time.Time         `json:"locked_until,omitempty" bson:"locked_until,omitempty"`
	CreatedAt             time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt             time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt             *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`

	// Initial setup flow
	RequiresPasswordSetup bool       `json:"requires_password_setup,omitempty" bson:"requires_password_setup,omitempty"`
	SetupToken           string     `json:"-" bson:"setup_token,omitempty"`
	SetupTokenExpiresAt  *time.Time `json:"-" bson:"setup_token_expires_at,omitempty"`

	// Password reset flow
	PasswordResetToken    string     `json:"-" bson:"password_reset_token,omitempty"`
	PasswordResetExpiresAt *time.Time `json:"-" bson:"password_reset_expires_at,omitempty"`
}

// ValidateEmail validates the email format
func (a *Admin) ValidateEmail() error {
	if a.Email == "" {
		return ErrInvalidEmail
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(a.Email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidateName validates the admin name
func (a *Admin) ValidateName() error {
	if len(a.Name) < 2 {
		return ErrInvalidInput
	}
	if len(a.Name) > 100 {
		return ErrInvalidInput
	}
	return nil
}

// ValidateRole validates the admin role
func (a *Admin) ValidateRole() error {
	if a.Role != AdminRoleSuperAdmin && a.Role != AdminRoleAdmin && a.Role != AdminRoleModerator {
		return ErrInvalidAdminRole
	}
	return nil
}

// Validate validates the entire admin
func (a *Admin) Validate() error {
	if err := a.ValidateEmail(); err != nil {
		return err
	}
	if err := a.ValidateName(); err != nil {
		return err
	}
	if err := a.ValidateRole(); err != nil {
		return err
	}
	return nil
}

// HasPermission checks if admin has a specific permission
func (a *Admin) HasPermission(permission AdminPermission) bool {
	// Super admin has all permissions
	if a.Role == AdminRoleSuperAdmin {
		return true
	}

	for _, p := range a.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// DefaultPermissionsForRole returns default permissions based on role
func DefaultPermissionsForRole(role AdminRole) []AdminPermission {
	switch role {
	case AdminRoleSuperAdmin:
		return []AdminPermission{
			PermissionManageAdmins,
			PermissionManageUsers,
			PermissionSuspendUsers,
			PermissionDeleteUsers,
			PermissionViewUserAnalytics,
			PermissionManagePosts,
			PermissionDeletePosts,
			PermissionViewPostAnalytics,
			PermissionManagePatches,
			PermissionDeletePatches,
			PermissionViewPatchAnalytics,
			PermissionViewLogs,
			PermissionViewAnalytics,
		}
	case AdminRoleAdmin:
		return []AdminPermission{
			PermissionManageUsers,
			PermissionSuspendUsers,
			PermissionDeleteUsers,
			PermissionViewUserAnalytics,
			PermissionManagePosts,
			PermissionDeletePosts,
			PermissionViewPostAnalytics,
			PermissionManagePatches,
			PermissionDeletePatches,
			PermissionViewPatchAnalytics,
			PermissionViewLogs,
			PermissionViewAnalytics,
		}
	case AdminRoleModerator:
		return []AdminPermission{
			PermissionManageUsers,
			PermissionSuspendUsers,
			PermissionManagePosts,
			PermissionDeletePosts,
			PermissionManagePatches,
			PermissionDeletePatches,
			PermissionViewLogs,
		}
	default:
		return []AdminPermission{}
	}
}
