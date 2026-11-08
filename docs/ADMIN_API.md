# Admin API Documentation

## Overview

The Admin API provides endpoints for administrators to manage users, content, and view analytics. The API uses JWT-based authentication with role-based access control (RBAC).

## Authentication

### Admin Roles

- **`super_admin`**: Full access to all admin features
- **`admin`**: Can manage users, content, and view analytics (cannot manage other admins)
- **`moderator`**: Limited access (can suspend users, delete posts/patches, view logs)

### Token Types

- **Access Token**: Short-lived (30 minutes by default), used for API requests
- **Refresh Token**: Long-lived (7 days by default), used to obtain new access tokens

### Authorization Header

All protected endpoints require:
```
Authorization: Bearer <access_token>
```

---

## Public Endpoints (No Authentication)

### Admin Login
```
POST /api/v1/admin/auth/login
Content-Type: application/json

{
  "email": "admin@example.com",
  "password": "securepassword"
}

Response (200 OK):
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer"
}
```

### Initial Password Setup
First-time admins use this to set up their password using the setup token provided during account creation.

```
POST /api/v1/admin/auth/initial-setup
Content-Type: application/json

{
  "email": "admin@example.com",
  "setup_token": "generated-token-from-creation",
  "password": "newSecurePassword"
}

Response (200 OK):
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "message": "Password setup completed successfully"
}
```

### Forgot Password
Initiates password reset flow.

```
POST /api/v1/admin/auth/forgot-password
Content-Type: application/json

{
  "email": "admin@example.com"
}

Response (200 OK):
{
  "message": "Password reset instructions sent to email",
  "reset_token": "optional-token-for-dev" // Only in development
}
```

### Reset Password
Completes the password reset using the reset token.

```
POST /api/v1/admin/auth/reset-password
Content-Type: application/json

{
  "reset_token": "provided-reset-token",
  "password": "newSecurePassword"
}

Response (200 OK):
{
  "message": "Password reset successfully"
}
```

### Refresh Access Token
```
POST /api/v1/admin/auth/refresh
Content-Type: application/json

{
  "refresh_token": "your-refresh-token"
}

Response (200 OK):
{
  "access_token": "new-access-token",
  "token_type": "Bearer"
}
```

---

## Protected Endpoints (Require Authentication)

All endpoints below require `Authorization: Bearer <access_token>` header.

### Admin Logout
Invalidates all refresh tokens for the authenticated admin.

```
POST /api/v1/admin/auth/logout
Authorization: Bearer <access_token>

Response (200 OK):
{
  "message": "Logged out successfully"
}
```

---

## Admin Management Endpoints

### View Admin
```
GET /api/v1/admin/admins/:id
Authorization: Bearer <access_token>

Response (200 OK):
{
  "id": "507f1f77bcf86cd799439011",
  "email": "admin@example.com",
  "name": "John Admin",
  "role": "admin",
  "permissions": ["manage_users", "manage_posts", "view_analytics"],
  "is_active": true,
  "last_login": "2024-10-31T10:30:00Z",
  "created_at": "2024-10-01T08:00:00Z",
  "updated_at": "2024-10-31T10:30:00Z"
}
```

### List All Admins
```
GET /api/v1/admin/admins?limit=10&offset=0
Authorization: Bearer <access_token>

Response (200 OK):
{
  "data": [
    {
      "id": "507f1f77bcf86cd799439011",
      "email": "admin@example.com",
      "name": "John Admin",
      "role": "admin",
      "is_active": true,
      ...
    }
  ],
  "limit": 10,
  "offset": 0
}
```

### Create New Admin
**Requires Permission**: `manage_admins` (super_admin only)

```
POST /api/v1/admin/admins
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "email": "newadmin@example.com",
  "name": "New Admin",
  "role": "admin"
}

Response (201 Created):
{
  "admin": {
    "id": "507f1f77bcf86cd799439012",
    "email": "newadmin@example.com",
    "name": "New Admin",
    "role": "admin",
    "permissions": [...],
    "is_active": true,
    "requires_password_setup": true,
    "created_at": "2024-10-31T12:00:00Z"
  },
  "setup_token": "setup-token-xyz123",
  "message": "Admin created. Send setup_token to admin via secure channel for password setup."
}
```

**Important**: Send the `setup_token` to the new admin securely (email, secure message, etc.). They will use this with the `/initial-setup` endpoint.

### Update Admin
**Requires Permission**: `manage_admins`

```
PATCH /api/v1/admin/admins/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Updated Name",
  "is_active": false
}

Response (200 OK):
{
  "message": "Admin updated successfully"
}
```

### Delete Admin
**Requires Permission**: `manage_admins` (super_admin only)

```
DELETE /api/v1/admin/admins/:id
Authorization: Bearer <access_token>

Response (200 OK):
{
  "message": "Admin deleted successfully"
}
```

**Constraints**:
- Cannot delete your own account
- Cannot delete the last super_admin account
- Admins are soft-deleted (records preserved for audit)

---

## Customer User Management Endpoints

### View Customer
```
GET /api/v1/admin/customers/:id
Authorization: Bearer <access_token>

Response (200 OK):
{
  "id": "507f191e810c19729de860ea",
  "code": "USER_001",
  "name": "John Doe",
  "email": "john@example.com",
  "is_age_verified": true,
  "is_email_verified": true,
  "is_suspended": false,
  "profile_picture": "https://...",
  "created_at": "2024-10-01T08:00:00Z",
  "updated_at": "2024-10-31T10:30:00Z",
  "deleted_at": null
}
```

### List All Customers
```
GET /api/v1/admin/customers?limit=20&offset=0
Authorization: Bearer <access_token>

Response (200 OK):
{
  "data": [
    {
      "id": "507f191e810c19729de860ea",
      "email": "john@example.com",
      ...
    }
  ],
  "limit": 20,
  "offset": 0
}
```

### Suspend Customer
**Requires Permission**: `suspend_users`

```
POST /api/v1/admin/customers/:id/suspend
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "reason": "Violated community guidelines"
}

Response (200 OK):
{
  "message": "User suspended successfully"
}
```

The user will be recorded as `is_suspended: true` with:
- `suspension_reason`: The reason provided
- `suspended_at`: Timestamp of suspension
- `suspended_by`: Admin ID who performed the suspension

### Unsuspend Customer
**Requires Permission**: `suspend_users`

```
DELETE /api/v1/admin/customers/:id/suspend
Authorization: Bearer <access_token>

Response (200 OK):
{
  "message": "User unsuspended successfully"
}
```

### Delete Customer
**Requires Permission**: `delete_users`

Soft-deletes the customer account (preserves data for audit).

```
DELETE /api/v1/admin/customers/:id
Authorization: Bearer <access_token>

Response (200 OK):
{
  "message": "User deleted successfully"
}
```

---

## Post Management Endpoints

### View Post
```
GET /api/v1/admin/posts/:id
Authorization: Bearer <access_token>

Response (200 OK):
{
  "id": "507f191e810c19729de860eb",
  "room_id": "507f191e810c19729de860ec",
  "user_id": "507f191e810c19729de860ed",
  "text": "This is a post",
  "image": "https://s3.../image.jpg",
  "video": null,
  "validations": ["507f191e810c19729de860ee"],
  "respects": ["507f191e810c19729de860ef"],
  "report_count": 2,
  "created_at": "2024-10-30T08:00:00Z",
  "updated_at": "2024-10-30T08:00:00Z",
  "deleted_at": null
}
```

### List All Posts
```
GET /api/v1/admin/posts?limit=20&offset=0
Authorization: Bearer <access_token>

Response (200 OK):
{
  "data": [...],
  "limit": 20,
  "offset": 0
}
```

### Delete Post
**Requires Permission**: `delete_posts`

```
DELETE /api/v1/admin/posts/:id
Authorization: Bearer <access_token>

Response (200 OK):
{
  "message": "Post deleted successfully"
}
```

### Update Post
**Requires Permission**: `manage_posts`

```
PATCH /api/v1/admin/posts/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "text": "Updated post content"
}

Response (200 OK):
{
  "message": "Post updated successfully"
}
```

---

## Patch/Room Management Endpoints

### View Patch
```
GET /api/v1/admin/patches/:id
Authorization: Bearer <access_token>

Response (200 OK):
{
  "id": "507f191e810c19729de860ec",
  "name": "Community Group",
  "code": "COMM_001",
  "created_by": "507f191e810c19729de860ed",
  "members": ["507f191e810c19729de860ee", "507f191e810c19729de860ef"],
  "created_at": "2024-10-01T08:00:00Z",
  "updated_at": "2024-10-31T10:30:00Z",
  "deleted_at": null
}
```

### List All Patches
```
GET /api/v1/admin/patches?limit=20&offset=0
Authorization: Bearer <access_token>

Response (200 OK):
{
  "data": [...],
  "limit": 20,
  "offset": 0
}
```

### Delete Patch
**Requires Permission**: `delete_patches`

```
DELETE /api/v1/admin/patches/:id
Authorization: Bearer <access_token>

Response (200 OK):
{
  "message": "Patch deleted successfully"
}
```

### Update Patch
**Requires Permission**: `manage_patches`

```
PATCH /api/v1/admin/patches/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Updated Group Name"
}

Response (200 OK):
{
  "message": "Patch updated successfully"
}
```

---

## Analytics Endpoints

### User Analytics
**Requires Permission**: `view_analytics` or `view_user_analytics`

```
GET /api/v1/admin/analytics/users
Authorization: Bearer <access_token>

Response (200 OK):
{
  "data": {
    "total_users": 1250,
    "deleted_users": 15,
    "suspended_users": 8,
    "verified_users": 1100,
    "unverified_users": 150
  }
}
```

### Post Analytics
**Requires Permission**: `view_analytics` or `view_post_analytics`

```
GET /api/v1/admin/analytics/posts
Authorization: Bearer <access_token>

Response (200 OK):
{
  "data": {
    "total_posts": 5600,
    "posts_created_today": 45,
    "deleted_posts": 120,
    "avg_posts_per_room": 3.2
  }
}
```

### Patch Analytics
**Requires Permission**: `view_analytics` or `view_patch_analytics`

```
GET /api/v1/admin/analytics/patches
Authorization: Bearer <access_token>

Response (200 OK):
{
  "data": {
    "total_patches": 450,
    "active_patches": 420,
    "avg_members_per_patch": 4.5
  }
}
```

---

## Admin Logs/Audit Trail Endpoints

### Get All Logs
**Requires Permission**: `view_logs`

```
GET /api/v1/admin/logs?limit=50&offset=0
Authorization: Bearer <access_token>

Response (200 OK):
{
  "data": [
    {
      "id": "507f191e810c19729de860fa",
      "admin_id": "507f191e810c19729de860fb",
      "admin_email": "admin@example.com",
      "action": "USER_SUSPEND",
      "resource_type": "USER",
      "resource_id": "507f191e810c19729de860fc",
      "changes": {
        "reason": "Violated community guidelines"
      },
      "status": "SUCCESS",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2024-10-31T10:30:00Z"
    }
  ],
  "limit": 50,
  "offset": 0
}
```

### Get Admin's Logs
**Requires Permission**: `view_logs`

```
GET /api/v1/admin/logs/admin/:admin_id?limit=50&offset=0
Authorization: Bearer <access_token>

Response (200 OK):
{
  "data": [...],
  "limit": 50,
  "offset": 0
}
```

---

## Error Responses

### Unauthorized (401)
```json
{
  "error": "Admin authentication is required",
  "code": "UNAUTHORIZED_ADMIN"
}
```

### Forbidden (403)
```json
{
  "error": "You do not have permission to perform this action",
  "code": "INSUFFICIENT_PERMISSIONS"
}
```

### Not Found (404)
```json
{
  "error": "Admin not found",
  "code": "ADMIN_NOT_FOUND"
}
```

### Conflict (409)
```json
{
  "error": "An admin with this email already exists",
  "code": "ADMIN_ALREADY_EXISTS"
}
```

### Invalid Input (400)
```json
{
  "error": "The provided input is invalid",
  "code": "INVALID_INPUT"
}
```

---

## Admin Permissions Reference

| Permission | Description | Default Roles |
|-----------|-------------|----------------|
| `manage_admins` | Create, update, delete admin accounts | super_admin |
| `manage_users` | View and edit user accounts | super_admin, admin |
| `suspend_users` | Suspend/unsuspend users | super_admin, admin, moderator |
| `delete_users` | Delete user accounts | super_admin, admin |
| `view_user_analytics` | View user statistics | super_admin, admin |
| `manage_posts` | Edit post content | super_admin, admin |
| `delete_posts` | Delete posts | super_admin, admin, moderator |
| `view_post_analytics` | View post statistics | super_admin, admin |
| `manage_patches` | Edit patch/room details | super_admin, admin |
| `delete_patches` | Delete patches/rooms | super_admin, admin, moderator |
| `view_patch_analytics` | View patch statistics | super_admin, admin |
| `view_logs` | Access audit logs | super_admin, admin, moderator |
| `view_analytics` | Access all analytics | super_admin, admin |

---

## Rate Limiting

Admin endpoints are NOT rate limited differently from regular endpoints. Standard rate limits apply.

---

## Security Best Practices

1. **Setup Tokens**: Always send setup tokens via secure channels (not email, preferably in-person or secure messaging)
2. **Passwords**: Enforce strong password requirements (minimum 8 characters)
3. **Token Storage**: Store tokens securely on the client (HttpOnly cookies preferred for web)
4. **Audit Logs**: Review audit logs regularly for suspicious activity
5. **Least Privilege**: Assign admin roles with minimum necessary permissions
6. **Token Refresh**: Implement automatic token refresh before expiration
7. **Session Timeout**: Log out admins after inactivity periods

---

## Examples for Frontend

### React/Fetch Admin Login
```javascript
async function adminLogin(email, password) {
  const response = await fetch('/api/v1/admin/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  
  if (!response.ok) throw new Error('Login failed');
  
  const { access_token, refresh_token } = await response.json();
  
  // Store tokens securely
  localStorage.setItem('adminAccessToken', access_token);
  localStorage.setItem('adminRefreshToken', refresh_token);
  
  return { access_token, refresh_token };
}
```

### React/Fetch Admin Request with Token
```javascript
async function fetchAdmin(adminId) {
  const token = localStorage.getItem('adminAccessToken');
  
  const response = await fetch(`/api/v1/admin/admins/${adminId}`, {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  if (response.status === 401) {
    // Token expired, refresh it
    await refreshAdminToken();
    return fetchAdmin(adminId); // Retry
  }
  
  return response.json();
}
```

### React/Fetch Suspend User
```javascript
async function suspendUser(userId, reason) {
  const token = localStorage.getItem('adminAccessToken');
  
  const response = await fetch(`/api/v1/admin/customers/${userId}/suspend`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ reason })
  });
  
  if (!response.ok) throw new Error('Suspend failed');
  return response.json();
}
```

