# Admin Setup Guide

This guide walks through setting up the admin system for the first time and managing admins.

## First Time Setup

### Step 1: Create the First Super Admin

The first super admin must be created directly in the MongoDB database since there are no existing admins yet.

#### Using MongoDB Shell

```bash
# Connect to MongoDB
mongosh --uri "mongodb://admin:password@localhost:27017/temp_backend?authSource=admin"

# Create the first super admin with a temporary hashed password
# You'll need to hash a temporary password using bcrypt

db.admins.insertOne({
  email: "super.admin@example.com",
  name: "Super Admin",
  role: "super_admin",
  permissions: [
    "manage_admins",
    "manage_users",
    "suspend_users",
    "delete_users",
    "view_user_analytics",
    "manage_posts",
    "delete_posts",
    "view_post_analytics",
    "manage_patches",
    "delete_patches",
    "view_patch_analytics",
    "view_logs",
    "view_analytics"
  ],
  is_active: true,
  requires_password_setup: true,
  setup_token: "temp_setup_token_here",
  setup_token_expires_at: new Date(Date.now() + 24 * 60 * 60 * 1000),
  password_hash: "$2a$10$...", // Leave empty or use a bcrypt hash
  password_reset_token: null,
  password_reset_token_expires_at: null,
  failed_login_attempts: 0,
  locked_until: null,
  created_at: new Date(),
  updated_at: new Date(),
  deleted_at: null
})
```

#### Or Using a Seed Script (Recommended)

Create `scripts/seed_admin.go`:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"temp_backend/config"
	"temp_backend/internal/domain"
	"temp_backend/pkg/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()
	client, err := mongodb.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database(cfg.Mongo.Database)
	collection := db.Collection("admins")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if super admin already exists
	var existing domain.Admin
	err = collection.FindOne(ctx, bson.M{"role": "super_admin"}).Decode(&existing)
	if err == nil {
		fmt.Println("Super admin already exists")
		return
	}

	// Create temporary password hash
	tempPassword := "TempPassword123!"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create admin
	admin := domain.Admin{
		Email:                 "super.admin@example.com",
		Name:                  "Super Admin",
		Role:                  domain.AdminRoleSuperAdmin,
		Permissions:           domain.DefaultPermissionsForRole(domain.AdminRoleSuperAdmin),
		IsActive:              true,
		RequiresPasswordSetup: false,
		PasswordHash:          string(passwordHash),
		CreatedAt:             time.Now().UTC(),
		UpdatedAt:             time.Now().UTC(),
	}

	result, err := collection.InsertOne(ctx, admin)
	if err != nil {
		log.Fatalf("Failed to insert admin: %v", err)
	}

	fmt.Printf("Super admin created successfully! ID: %v\n", result.InsertedID)
	fmt.Printf("Email: %s\n", admin.Email)
	fmt.Printf("Temporary Password: %s\n", tempPassword)
	fmt.Println("\nIMPORTANT: Change this password after first login!")
}
```

Run with:
```bash
go run scripts/seed_admin.go
```

### Step 2: First Login

```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "super.admin@example.com",
    "password": "TempPassword123!"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "admin": {
    "id": "507f1f77bcf86cd799439011",
    "email": "super.admin@example.com",
    "name": "Super Admin",
    "role": "super_admin"
  }
}
```

### Step 3: Change Initial Password

The super admin should immediately change their password. Use the profile update endpoint (when implemented) or call:

```bash
curl -X POST http://localhost:8080/api/v1/admin/profile/change-password \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "TempPassword123!",
    "new_password": "SecureNewPassword123!"
  }'
```

## Managing Admin Users

### Creating Additional Admins

Only super admins can create new admins. When creating an admin:

1. The admin is created with `requires_password_setup: true`
2. A `setup_token` is generated (valid for 24 hours)
3. The setup token must be sent to the admin via a secure channel (email, Slack, etc.)
4. The admin uses the setup token to set their initial password

```bash
# Super admin creates a new admin
curl -X POST http://localhost:8080/api/v1/admin/admins \
  -H "Authorization: Bearer SUPER_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "name": "John Admin",
    "role": "admin"
  }'
```

Response:
```json
{
  "admin": {
    "id": "507f1f77bcf86cd799439012",
    "email": "admin@example.com",
    "name": "John Admin",
    "role": "admin",
    "permissions": [...]
  },
  "setup_token": "5f6f8c4d9e2a7b1c3a0f5e8d9b2c4f6a"
}
```

**IMPORTANT:** Send the `setup_token` to the admin securely. The token is only returned once and cannot be retrieved later.

### Admin Setup Process

The new admin receives the setup token and uses it to set their initial password:

```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/initial-setup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "setup_token": "5f6f8c4d9e2a7b1c3a0f5e8d9b2c4f6a",
    "password": "AdminPassword123!"
  }'
```

After this, the admin can log in normally:

```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "AdminPassword123!"
  }'
```

### Listing Admins

```bash
curl -X GET "http://localhost:8080/api/v1/admin/admins?limit=20&offset=0" \
  -H "Authorization: Bearer SUPER_ADMIN_TOKEN"
```

Response:
```json
{
  "data": [
    {
      "id": "507f1f77bcf86cd799439011",
      "email": "super.admin@example.com",
      "name": "Super Admin",
      "role": "super_admin",
      "is_active": true,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "limit": 20,
  "offset": 0
}
```

### Getting Admin Details

```bash
curl -X GET "http://localhost:8080/api/v1/admin/admins/507f1f77bcf86cd799439012" \
  -H "Authorization: Bearer SUPER_ADMIN_TOKEN"
```

### Updating Admin Details

```bash
curl -X PATCH "http://localhost:8080/api/v1/admin/admins/507f1f77bcf86cd799439012" \
  -H "Authorization: Bearer SUPER_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Updated Admin",
    "is_active": true,
    "role": "admin"
  }'
```

### Deleting an Admin

Only super admins can delete admins. Admins are soft-deleted (marked as deleted but not removed from database).

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/admins/507f1f77bcf86cd799439012" \
  -H "Authorization: Bearer SUPER_ADMIN_TOKEN"
```

## Password Management

### Forgotten Password

If an admin forgets their password, they can request a password reset:

```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com"}'
```

Response:
```json
{
  "message": "Password reset token has been sent to your email",
  "reset_token": "4d7e9f2c8a1b5e6f3d0c9a2b7e4f1d8c"
}
```

The admin then uses the reset token to set a new password:

```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "reset_token": "4d7e9f2c8a1b5e6f3d0c9a2b7e4f1d8c",
    "new_password": "NewPassword123!"
  }'
```

## Admin Permissions

### Role-Based Access Control (RBAC)

Three roles are available:

#### Super Admin
- Full access to all features
- Can manage other admins
- Can assign permissions to regular admins

Permissions:
- `manage_admins`
- `manage_users`
- `suspend_users`
- `delete_users`
- `view_user_analytics`
- `manage_posts`
- `delete_posts`
- `view_post_analytics`
- `manage_patches`
- `delete_patches`
- `view_patch_analytics`
- `view_logs`
- `view_analytics`

#### Admin
- User management (view, suspend, delete)
- Content management (delete posts, patches)
- View analytics

Default Permissions:
- `manage_users`
- `suspend_users`
- `delete_users`
- `view_user_analytics`
- `delete_posts`
- `view_post_analytics`
- `delete_patches`
- `view_patch_analytics`

#### Moderator
- Basic content moderation
- View analytics

Default Permissions:
- `delete_posts`
- `view_post_analytics`
- `delete_patches`
- `view_patch_analytics`

### Checking Permissions

Permissions are encoded in the JWT access token. After decoding the token, you can check the `admin_permissions` claim:

```javascript
// In frontend
const token = localStorage.getItem('adminAccessToken');
const decoded = jwt_decode(token);
console.log(decoded.admin_permissions); // Array of permissions

// Check if admin has a specific permission
const hasPermission = (permission) => {
  const decoded = jwt_decode(localStorage.getItem('adminAccessToken'));
  return decoded.role === 'super_admin' || 
         decoded.admin_permissions.includes(permission);
};
```

## Account Lock Out

After 5 failed login attempts, an admin account is locked for 15 minutes. The account automatically unlocks after the lockout period expires.

To unlock an account immediately, a super admin can update the admin:

```bash
curl -X PATCH "http://localhost:8080/api/v1/admin/admins/507f1f77bcf86cd799439012" \
  -H "Authorization: Bearer SUPER_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "failed_login_attempts": 0,
    "locked_until": null
  }'
```

## Security Best Practices

1. **Strong Passwords**: Enforce minimum 8 characters with mixed case, numbers, and symbols
2. **Token Storage**: Store tokens in HTTP-only, secure cookies (not localStorage if possible)
3. **Token Rotation**: Refresh tokens before expiry to maintain session
4. **Audit Logging**: Monitor all admin actions via the logs endpoint
5. **Regular Reviews**: Periodically review admin access and permissions
6. **Deactivate Unused Admins**: Set `is_active: false` for admins no longer needed
7. **Change Default Credentials**: Always change initial super admin password
8. **Secure Setup Tokens**: Transmit setup tokens through secure channels only

## Troubleshooting

### Super admin doesn't exist

If you get "admin not found" errors during setup, you need to create the first super admin in MongoDB as shown in Step 1.

### Setup token expired

Setup tokens expire after 24 hours. If an admin doesn't set their password in time, a super admin must create a new admin with a new setup token.

### Locked account

If an admin account is locked due to failed login attempts, a super admin can unlock it by updating the admin with `locked_until: null`.

### Permission denied errors

Ensure the admin has the required permission for the operation. Check the `admin_permissions` in the JWT token or review the permissions assigned to the admin's role.

## Environment Configuration

The admin system uses these environment variables:

```bash
# Admin token settings (seconds)
ADMIN_ACCESS_TOKEN_TTL=1800              # 30 minutes
ADMIN_REFRESH_TOKEN_TTL=604800           # 7 days

# Admin security
ADMIN_MAX_LOGIN_ATTEMPTS=5               # Lockout after 5 attempts
ADMIN_LOCKOUT_DURATION=900               # 15 minutes lockout
ADMIN_SETUP_TOKEN_EXPIRY=86400           # 24 hours for setup token

# JWT secret (must be at least 32 characters)
JWT_SECRET=your_super_secret_key_here_minimum_32_chars
```

Make sure these are set in your `.env` file before starting the application.
