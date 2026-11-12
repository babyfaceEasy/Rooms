# Admin Login & Profile Endpoint Testing

This guide documents how to test the admin authentication endpoints and the admin profile retrieval endpoint that were recently fixed.

## Issues Fixed

### 1. Route Registration Order Issue
**Problem**: The public admin auth routes (`/api/v1/admin/auth/*`) were being caught by the protected routes middleware, requiring authentication for public endpoints.

**Root Cause**: Fiber's route matching was prioritizing the parent `/api/v1` group with middleware over the more specific `/api/v1/admin/auth` public group, because the protected group was created with an empty path and the middleware was applied to all sub-routes.

**Solution**: Moved the admin auth public routes to be registered in the main `registerRoutes()` function before the protected routes group is created. This ensures they are registered directly on the app root and not subject to the protected routes middleware.

### 2. Admin User Password Hash
**Problem**: The pre-created admin user had an invalid password hash that didn't match "password123".

**Solution**: Updated the admin user's password_hash in MongoDB to the correct bcrypt hash of "password123".

## Testing Endpoints

### Prerequisites
- Ensure the application is running: `docker-compose up -d`
- MongoDB should have the admin user created with email `admin@example.com` and password `password123`

### 1. Admin Login (Public)

```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'
```

**Expected Response (200 OK)**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer"
}
```

### 2. Admin Profile (Protected)

Use the `access_token` from the login response:

```bash
TOKEN="<access_token_from_login>"

curl -X GET http://localhost:8080/api/v1/admin/profile \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK)**:
```json
{
  "id": "507f1f77bcf86cd799439011",
  "email": "admin@example.com",
  "name": "System Admin",
  "role": "super_admin",
  "permissions": [
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
  "is_active": true,
  "last_login": "2026-11-12T15:48:28.062Z",
  "created_at": "2026-11-12T15:10:26.205Z",
  "updated_at": "2026-11-12T15:48:28.062Z"
}
```

### 3. Testing Without Token (Protected Endpoint)

```bash
curl -X GET http://localhost:8080/api/v1/admin/profile
```

**Expected Response (401 Unauthorized)**:
```json
{
  "error": "missing authorization header"
}
```

## Creating Additional Admin Users

To create a new admin user in MongoDB:

```bash
docker-compose exec -T mongodb mongosh mongodb://admin:secret@localhost:27017/temp_backend?authSource=admin --eval "
db.admins.insertOne({
  email: 'newadmin@example.com',
  password_hash: '\$2a\$10\$l7jPnFGq0ZUTTdsKAPgvv.Nn9n9aP3p.w8T41egx/H1SkWLNKRshW',
  name: 'New Admin',
  role: 'admin',
  permissions: ['view_analytics', 'view_logs'],
  is_active: true,
  failed_login_attempts: 0,
  created_at: new Date(),
  updated_at: new Date(),
  locked_until: null
})
"
```

**Note**: The password_hash shown above corresponds to password "password123". To generate a hash for a different password, use:

```go
package main
import ("fmt"; "golang.org/x/crypto/bcrypt")
func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte("your_password"), bcrypt.DefaultCost)
    fmt.Println(string(hash))
}
```

## Files Modified

1. **internal/api/routes.go**
   - Moved admin auth public routes registration to `registerRoutes()` function
   - Registered admin auth routes before protected routes to prevent middleware conflicts

2. **internal/api/server.go**
   - Updated `registerRoutes()` function signature to include `adminAuthHandler` parameter
   - Passed `adminAuthHandler` when calling `registerRoutes()`

3. **Database** (MongoDB)
   - Updated admin user password_hash to correct bcrypt hash of "password123"

## Architecture Notes

- **Route Registration**: Public routes must be registered before protected routes that use the same path prefix. Fiber evaluates routes in registration order.
- **Middleware Isolation**: Protected routes use a separate group with middleware applied only to that group, ensuring public routes remain unaffected.
- **Clean Architecture**: Error handling properly propagates `AppError` types from service layer to handler, which return appropriate HTTP status codes via middleware.

