# Admin System Architecture

This document describes the architecture and design of the admin section.

## Overview

The admin system follows the **Clean/Hexagonal Architecture** pattern used throughout the application:

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Handlers (fiber)                     │
│  AdminAuthHandler | AdminUserHandler | AdminAnalyticsHandler │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                     Middleware Layer                         │
│  AdminAuthMiddleware | AdminPermissionMiddleware (RBAC)      │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                    Service Layer                             │
│ AdminAuthService | AdminUserService | AdminUserMgmtService │
│            AdminLogService (Audit Trail)                    │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                   Repository Layer                           │
│  AdminRepository | AdminLogRepository | AdminRefreshToken    │
│           UserRepository (extended)                         │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│              Database (MongoDB)                             │
│  Collections: admins | admin_logs | admin_refresh_tokens   │
│             users (with admin fields)                       │
└─────────────────────────────────────────────────────────────┘
```

## Layer Descriptions

### Domain Layer (`internal/domain/`)

**Files:**
- `admin.go` - Admin entity with roles, permissions, and validation
- `admin_log.go` - Audit log entity
- `admin_errors.go` - Admin-specific errors

**Key Concepts:**
- **Roles**: `super_admin`, `admin`, `moderator`
- **Permissions**: 13 permission constants (manage_admins, manage_users, etc.)
- **Role-Based Permissions**: Each role has default permissions via `DefaultPermissionsForRole()`

```go
type Admin struct {
    ID                     primitive.ObjectID
    Email                  string
    Name                   string
    Role                   AdminRole
    Permissions            []AdminPermission  // Specific permissions for this admin
    IsActive               bool
    RequiresPasswordSetup  bool
    PasswordHash           string
    FailedLoginAttempts    int
    LockedUntil            *time.Time
    SetupToken             string
    SetupTokenExpiresAt    *time.Time
    PasswordResetToken     string
    PasswordResetTokenExpiresAt *time.Time
    CreatedAt              time.Time
    UpdatedAt              time.Time
    DeletedAt              *time.Time
}
```

### Repository Layer (`internal/repository/`)

**Files:**
- `admin_repository.go` - Interface definition
- `mongo_admin_repository.go` - MongoDB implementation (400+ lines)
- `admin_log_repository.go` - Interface for audit logs
- `mongo_admin_log_repository.go` - MongoDB implementation with TTL
- `admin_refresh_token_repository.go` - Admin token storage
- `mongo_admin_refresh_token_repository.go` - Implementation
- `user_repository.go` - Extended with 8 new methods for admin operations

**Key Methods:**

AdminRepository:
```go
Create(ctx, admin) error
GetByID(ctx, id) (*Admin, error)
GetByEmail(ctx, email) (*Admin, error)
List(ctx, limit, offset) ([]*Admin, error)
Update(ctx, admin) error
Delete(ctx, id) error                    // Soft delete
GetSuperAdminCount(ctx) (int, error)
StoreSetupToken(ctx, id, token, expiry) error
VerifySetupToken(ctx, email, token) error
ClearSetupToken(ctx, id) error
ClearFailedAttempts(ctx, id) error
SetFailedAttempts(ctx, id, count, lockedUntil) error
GetBySetupToken(ctx, token) (*Admin, error)
```

AdminLogRepository:
```go
Create(ctx, log) error
GetByAdminID(ctx, adminID, limit, offset) ([]*AdminLog, error)
GetAll(ctx, limit, offset) ([]*AdminLog, error)
// TTL index auto-deletes logs older than 90 days
```

**Database Indexes:**
```javascript
// admins collection
db.admins.createIndex({ email: 1 }, { unique: true })
db.admins.createIndex({ deleted_at: 1 })
db.admins.createIndex({ role: 1, is_active: 1 })

// admin_logs collection
db.admin_logs.createIndex({ admin_id: 1 })
db.admin_logs.createIndex({ created_at: 1 }, { expireAfterSeconds: 7776000 }) // 90 days

// admin_refresh_tokens collection
db.admin_refresh_tokens.createIndex({ admin_id: 1 })
```

### Service Layer (`internal/service/`)

**Files:**
- `admin_auth_service.go` - Authentication, token management, password reset
- `admin_user_service.go` - Admin CRUD operations
- `admin_user_mgmt_service.go` - Customer management (suspend, delete, analytics)
- `admin_log_service.go` - Audit logging

**AdminAuthService (1000+ lines)**

Key Methods:
```go
// Authentication
Login(ctx, email, password) (accessToken, refreshToken string, error)
ValidateToken(ctx, token string) (*AdminJWTClaims, error)
RefreshAccessToken(ctx, refreshToken string) (newAccessToken string, error)
Logout(ctx, adminID ObjectID) error

// Setup & Password Reset
InitialSetup(ctx, email, setupToken, password) (accessToken, refreshToken, error)
ForgotPassword(ctx, email) (resetToken string, error)
ResetPassword(ctx, email, resetToken, newPassword) error

// Token Management
GenerateTokenPair(ctx, admin) (accessToken, refreshToken, error)
GenerateSetupToken(ctx, admin) (token string, expiry time.Time, error)
```

**Features:**
- Failed login attempt tracking with 15-minute lockout after 5 attempts
- Secure token hashing (SHA256) for refresh tokens
- Setup token flow for new admins (one-time use, 24-hour expiry)
- Password reset flow with secure tokens (one-time use, 1-hour expiry)

**AdminUserService**

```go
ListAdmins(ctx, limit, offset) ([]*Admin, error)
CreateAdmin(ctx, email, name, role) (*Admin, setupToken, error)
GetAdmin(ctx, id) (*Admin, error)
UpdateAdmin(ctx, id, updates) error
DeleteAdmin(ctx, id) error
```

**AdminUserManagementService**

```go
// Customer management
GetCustomer(ctx, userID) (*User, error)
ListCustomers(ctx, limit, offset) ([]*User, error)
SuspendUser(ctx, userID, reason, adminID) error
UnsuspendUser(ctx, userID, adminID) error

// Analytics
GetUserAnalytics(ctx) (*UserAnalytics, error)
GetPostAnalytics(ctx) (*PostAnalytics, error)
GetPatchAnalytics(ctx) (*PatchAnalytics, error)
```

**AdminLogService**

```go
LogAction(ctx, adminID, email, action, resourceType, resourceID, changes, ipAddress, userAgent) error
GetLogs(ctx, limit, offset) ([]*AdminLog, error)
GetAdminLogs(ctx, adminID, limit, offset) ([]*AdminLog, error)
```

### Middleware Layer (`internal/middleware/`)

**Files:**
- `admin_auth.go` - JWT validation and claims extraction
- `admin_permission.go` - Role-Based Access Control (RBAC)

**AdminAuthMiddleware**

```go
// Extracts and validates JWT from Authorization header
// Sets in c.Locals():
// - admin_id (ObjectID)
// - admin_email (string)
// - admin_role (AdminRole)
// - admin_permissions ([]AdminPermission)
// - admin_claims (*AdminJWTClaims)
```

**AdminPermissionMiddleware**

```go
// Factory function that creates middleware for specific permissions
func RequirePermission(requiredPermissions ...AdminPermission) func(*fiber.Ctx) error

// Check if admin has ANY of the required permissions
// Super admins bypass all permission checks
```

### Handler Layer (`internal/handler/`)

**Files:**
- `admin_auth_handler.go` - 6 auth endpoints
- `admin_user_handler.go` - 5 admin management endpoints
- `admin_customer_handler.go` - 5 customer management endpoints
- `admin_analytics_handler.go` - 3 analytics endpoints
- `admin_logs_handler.go` - 2 log access endpoints

**AdminAuthHandler**

```go
POST   /api/v1/admin/auth/login           - Login with email/password
POST   /api/v1/admin/auth/initial-setup   - Set initial password with setup token
POST   /api/v1/admin/auth/forgot-password - Request password reset
POST   /api/v1/admin/auth/reset-password  - Reset password with token
POST   /api/v1/admin/auth/refresh         - Refresh access token
POST   /api/v1/admin/auth/logout          - Logout (invalidate refresh tokens)
```

**AdminUserHandler**

```go
GET    /api/v1/admin/admins               - List all admins (paginated)
POST   /api/v1/admin/admins               - Create new admin (super admin only)
GET    /api/v1/admin/admins/:id           - Get admin details
PATCH  /api/v1/admin/admins/:id           - Update admin details
DELETE /api/v1/admin/admins/:id           - Delete admin (soft delete)
```

**AdminCustomerHandler**

```go
GET    /api/v1/admin/customers            - List all customers (paginated)
GET    /api/v1/admin/customers/:id        - Get customer details
POST   /api/v1/admin/customers/:id/suspend    - Suspend customer
DELETE /api/v1/admin/customers/:id/suspend    - Unsuspend customer
DELETE /api/v1/admin/customers/:id        - Delete customer (soft delete)
```

**AdminAnalyticsHandler**

```go
GET    /api/v1/admin/analytics/users      - User statistics
GET    /api/v1/admin/analytics/posts      - Post statistics
GET    /api/v1/admin/analytics/patches    - Patch/room statistics
```

**AdminLogsHandler**

```go
GET    /api/v1/admin/logs                 - List all admin action logs (paginated)
GET    /api/v1/admin/logs/admin/:id       - Get logs for specific admin
```

## Authentication Flow

### Login Flow

```
1. Admin enters credentials (email/password)
   ↓
2. POST /admin/auth/login
   ↓
3. AdminAuthHandler validates input
   ↓
4. AdminAuthService.Login():
   - Find admin by email
   - Check if account locked (locked_until > now)
   - Compare password hash
   - On success: reset failed_login_attempts
   - On failure: increment failed_login_attempts, lock if > 5
   ↓
5. Generate token pair (access + refresh)
   ↓
6. Store refresh token hash in database
   ↓
7. Return access_token + refresh_token to client
```

### Authorization Flow (Protected Requests)

```
1. Client sends request with Authorization: Bearer <access_token>
   ↓
2. AdminAuthMiddleware:
   - Extract token from header
   - Validate JWT signature
   - Check expiry
   - Extract claims (admin_id, role, permissions)
   - Store in c.Locals()
   ↓
3. Handler receives request
   ↓
4. AdminPermissionMiddleware (if applicable):
   - Check if admin has required permission
   - Super admins bypass checks
   ↓
5. Handler processes request using admin_id from c.Locals()
```

### Setup Token Flow (New Admin)

```
1. Super admin creates new admin
   ↓
2. AdminUserService.CreateAdmin():
   - Validate email/name/role
   - Generate setup_token
   - Set requires_password_setup = true
   - Save to database with expiry (24 hours)
   ↓
3. Return setup_token to super admin
   ↓
4. Super admin sends token to new admin securely
   ↓
5. New admin calls /admin/auth/initial-setup
   ↓
6. AdminAuthService.InitialSetup():
   - Verify setup_token matches email and hasn't expired
   - Hash new password
   - Save password_hash
   - Clear setup_token
   - Set requires_password_setup = false
   ↓
7. Issue access + refresh tokens
   ↓
8. Admin can now login normally
```

## JWT Token Structure

### Access Token Claims (AdminJWTClaims)

```go
type AdminJWTClaims struct {
    AdminID     primitive.ObjectID
    Email       string
    Role        AdminRole
    AdminPermissions []AdminPermission
    ExpiresAt   jwt.NumericDate
}
```

Default TTL: 30 minutes (can be changed via ADMIN_ACCESS_TOKEN_TTL)

### Refresh Token

Also a JWT, stored as SHA256 hash in database for security.
Default TTL: 7 days (can be changed via ADMIN_REFRESH_TOKEN_TTL)

## Audit Logging

All sensitive admin actions are logged in the `admin_logs` collection:

```go
type AdminLog struct {
    ID           primitive.ObjectID
    AdminID      primitive.ObjectID
    AdminEmail   string
    Action       AdminLogAction        // CREATED, UPDATED, DELETED, LOGIN, etc.
    ResourceType AdminLogResourceType  // ADMIN, USER, POST, PATCH
    ResourceID   *primitive.ObjectID
    Changes      map[string]interface{} // Before/after values
    Success      bool
    IPAddress    string
    UserAgent    string
    CreatedAt    time.Time
}
```

**Actions Logged:**
- `ADMIN_LOGIN` - Admin login attempt
- `ADMIN_SETUP` - Admin initial setup
- `ADMIN_PASSWORD_RESET` - Password reset
- `ADMIN_CREATED` - New admin created
- `ADMIN_UPDATED` - Admin details updated
- `ADMIN_DELETED` - Admin deleted
- `USER_SUSPENDED` - User suspended
- `USER_UNSUSPENDED` - User unsuspended
- `USER_DELETED` - User deleted
- `POST_DELETED` - Post deleted
- `PATCH_DELETED` - Patch deleted

**Auto-Cleanup:** Logs older than 90 days are automatically deleted via TTL index.

## Configuration

Admin system uses these environment variables:

```bash
# Token TTLs (in seconds)
ADMIN_ACCESS_TOKEN_TTL=1800              # 30 minutes
ADMIN_REFRESH_TOKEN_TTL=604800           # 7 days

# Security settings
ADMIN_MAX_LOGIN_ATTEMPTS=5
ADMIN_LOCKOUT_DURATION=900               # 15 minutes in seconds
ADMIN_SETUP_TOKEN_EXPIRY=86400           # 24 hours

# JWT signing (must be at least 32 characters)
JWT_SECRET=your_secret_key_minimum_32_chars

# Email (for password reset and setup tokens)
SMTP_HOST=smtp.mailtrap.io
SMTP_PORT=587
SMTP_USER=your_username
SMTP_PASSWORD=your_password
SMTP_FROM=admin@example.com
```

## Error Handling

Admin-specific errors are defined in `internal/domain/admin_errors.go`:

```go
ErrAdminNotFound
ErrAdminAlreadyExists
ErrInvalidAdminRole
ErrAdminAccountLocked
ErrInvalidSetupToken
ErrSetupTokenExpired
ErrInvalidResetToken
ErrResetTokenExpired
ErrInvalidCredentials
ErrInsufficientPermissions
ErrUnauthorizedAdminOperation
ErrCannotDeleteLastSuperAdmin
```

All errors are mapped to HTTP status codes by the global error handler middleware.

## Testing

The admin system includes unit and integration tests:

```bash
# Run all tests
go test ./...

# Run only admin service tests
go test ./internal/service -v -run Admin

# Run only admin handler tests
go test ./internal/handler -v -run Admin
```

## Future Enhancements

1. **Email Integration**: Send setup tokens and password reset links via email
2. **2FA**: Two-factor authentication for admin accounts
3. **Admin Audit Dashboard**: UI for viewing and analyzing admin logs
4. **Permission Management**: UI for dynamically assigning permissions
5. **Bulk Operations**: Suspend multiple users at once
6. **Admin Activity Timeline**: Visual timeline of admin actions
7. **Alerts**: Real-time alerts for critical admin actions
8. **Admin Sessions**: Track active admin sessions and allow remote logout

## Security Considerations

1. **Passwords**: Hashed with bcrypt at cost 10 (same as user passwords)
2. **Tokens**: JWT signed with HS256, refresh tokens stored as SHA256 hashes
3. **Setup Tokens**: One-time use, 24-hour expiry, sent securely
4. **Account Lockout**: Automatic 15-minute lockout after 5 failed attempts
5. **Audit Trail**: All actions logged for compliance
6. **Soft Deletes**: Admins and users are soft-deleted to preserve audit trail
7. **Permission Checks**: Two-layer validation (middleware + service)
8. **RBAC**: Role-based access control with super admin override

## File Organization

```
temp_backend/
├── internal/
│   ├── domain/
│   │   ├── admin.go                      # Admin entity (180 lines)
│   │   ├── admin_log.go                  # Audit log entity (70 lines)
│   │   └── admin_errors.go               # Error definitions (50 lines)
│   │
│   ├── repository/
│   │   ├── admin_repository.go           # Interface (50 lines)
│   │   ├── mongo_admin_repository.go     # MongoDB impl (400+ lines)
│   │   ├── admin_log_repository.go       # Interface (30 lines)
│   │   ├── mongo_admin_log_repository.go # MongoDB impl (120 lines)
│   │   ├── admin_refresh_token_repository.go  # Interface (20 lines)
│   │   └── mongo_admin_refresh_token_repository.go # Impl (100 lines)
│   │
│   ├── service/
│   │   ├── admin_auth_service.go         # Auth logic (1000+ lines)
│   │   ├── admin_user_service.go         # Admin CRUD (200 lines)
│   │   ├── admin_user_mgmt_service.go    # Customer mgmt (250 lines)
│   │   └── admin_log_service.go          # Audit logging (100 lines)
│   │
│   ├── handler/
│   │   ├── admin_auth_handler.go         # Auth endpoints (300 lines)
│   │   ├── admin_user_handler.go         # Admin mgmt (200 lines)
│   │   ├── admin_customer_handler.go     # Customer mgmt (250 lines)
│   │   ├── admin_analytics_handler.go    # Analytics (180 lines)
│   │   └── admin_logs_handler.go         # Log access (100 lines)
│   │
│   ├── middleware/
│   │   ├── admin_auth.go                 # JWT validation (80 lines)
│   │   └── admin_permission.go           # RBAC (60 lines)
│   │
│   └── api/
│       └── routes.go                     # Admin routes registration (150 lines)
│
├── config/
│   └── config.go                         # Admin config section (50 lines)
│
├── cmd/api/
│   └── main.go                           # Dependency wiring (100+ lines)
│
└── docs/
    ├── ADMIN_API.md                      # API reference (400+ lines)
    ├── ADMIN_SETUP_GUIDE.md              # Setup & deployment guide
    ├── ADMIN_FRONTEND_INTEGRATION.md    # Frontend integration guide
    └── ADMIN_ARCHITECTURE.md             # This file
```

**Total Lines of Code:** ~3500+ lines across all layers
