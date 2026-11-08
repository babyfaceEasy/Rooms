# Admin API Quick Reference

## Authentication

### Login
```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "password123"}'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "admin": {"id": "...", "email": "...", "name": "...", "role": "..."}
}
```

### Setup Initial Password (New Admin)
```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/initial-setup \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "setup_token": "token123", "password": "newpass123"}'
```

### Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "token123"}'
```

### Logout
```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Admin Management (Super Admin Only)

### Create Admin
```bash
curl -X POST http://localhost:8080/api/v1/admin/admins \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email": "newadmin@example.com", "name": "New Admin", "role": "admin"}'
```

### List Admins
```bash
curl -X GET "http://localhost:8080/api/v1/admin/admins?limit=20&offset=0" \
  -H "Authorization: Bearer TOKEN"
```

### Get Admin Details
```bash
curl -X GET http://localhost:8080/api/v1/admin/admins/ADMIN_ID \
  -H "Authorization: Bearer TOKEN"
```

### Update Admin
```bash
curl -X PATCH http://localhost:8080/api/v1/admin/admins/ADMIN_ID \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Name", "is_active": true}'
```

### Delete Admin (Soft Delete)
```bash
curl -X DELETE http://localhost:8080/api/v1/admin/admins/ADMIN_ID \
  -H "Authorization: Bearer TOKEN"
```

## Customer Management

### List Customers
```bash
curl -X GET "http://localhost:8080/api/v1/admin/customers?limit=20&offset=0" \
  -H "Authorization: Bearer TOKEN"
```

### Get Customer Details
```bash
curl -X GET http://localhost:8080/api/v1/admin/customers/USER_ID \
  -H "Authorization: Bearer TOKEN"
```

### Suspend Customer
```bash
curl -X POST http://localhost:8080/api/v1/admin/customers/USER_ID/suspend \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Violation of community guidelines"}'
```

### Unsuspend Customer
```bash
curl -X DELETE http://localhost:8080/api/v1/admin/customers/USER_ID/suspend \
  -H "Authorization: Bearer TOKEN"
```

### Delete Customer (Soft Delete)
```bash
curl -X DELETE http://localhost:8080/api/v1/admin/customers/USER_ID \
  -H "Authorization: Bearer TOKEN"
```

## Analytics

### User Analytics
```bash
curl -X GET http://localhost:8080/api/v1/admin/analytics/users \
  -H "Authorization: Bearer TOKEN"
```

### Post Analytics
```bash
curl -X GET http://localhost:8080/api/v1/admin/analytics/posts \
  -H "Authorization: Bearer TOKEN"
```

### Patch Analytics
```bash
curl -X GET http://localhost:8080/api/v1/admin/analytics/patches \
  -H "Authorization: Bearer TOKEN"
```

## Audit Logs

### List All Logs
```bash
curl -X GET "http://localhost:8080/api/v1/admin/logs?limit=50&offset=0" \
  -H "Authorization: Bearer TOKEN"
```

### Get Logs for Specific Admin
```bash
curl -X GET "http://localhost:8080/api/v1/admin/logs/admin/ADMIN_ID?limit=50&offset=0" \
  -H "Authorization: Bearer TOKEN"
```

## Roles & Permissions

### Super Admin
All 13 permissions (full access)

### Admin
- manage_users
- suspend_users
- delete_users
- view_user_analytics
- delete_posts
- view_post_analytics
- delete_patches
- view_patch_analytics

### Moderator
- delete_posts
- view_post_analytics
- delete_patches
- view_patch_analytics

## Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Invalid input |
| 401 | Unauthorized (missing/invalid token) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Not found |
| 409 | Conflict (e.g., email already exists) |
| 500 | Server error |

## Common Response Structure

Success:
```json
{
  "data": {...},
  "limit": 20,
  "offset": 0
}
```

Error:
```json
{
  "error": "Error message describing the issue"
}
```

## Headers

Required for protected endpoints:
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

## Environment Variables

```bash
ADMIN_ACCESS_TOKEN_TTL=1800              # seconds
ADMIN_REFRESH_TOKEN_TTL=604800           # seconds
ADMIN_MAX_LOGIN_ATTEMPTS=5
ADMIN_LOCKOUT_DURATION=900               # seconds
ADMIN_SETUP_TOKEN_EXPIRY=86400           # seconds
JWT_SECRET=your_secret_minimum_32_chars
```

## Documentation Files

| File | Purpose |
|------|---------|
| ADMIN_API.md | Complete API reference with examples |
| ADMIN_SETUP_GUIDE.md | Setup, deployment, and management |
| ADMIN_FRONTEND_INTEGRATION.md | Frontend implementation guide |
| ADMIN_ARCHITECTURE.md | System design and architecture |
| ADMIN_README.md | Master documentation index |

## Troubleshooting

**"Insufficient permissions"**
- Check your admin role and required permissions
- Super admins have all permissions

**"Token expired"**
- Use refresh endpoint to get new access token
- If refresh token also expired, login again

**"Admin not found"**
- Create first super admin in MongoDB
- See ADMIN_SETUP_GUIDE.md

**"Email already exists"**
- Admin with this email already exists
- Delete existing admin first or use different email

**Account locked**
- 15-minute lockout after 5 failed login attempts
- Wait or have super admin reset failed_login_attempts

## Important Notes

1. **Setup tokens** are only returned once during admin creation
2. **Reset tokens** are valid for 1 hour only
3. **Setup tokens** are valid for 24 hours only
4. **Access tokens** expire after 30 minutes (configurable)
5. **Refresh tokens** expire after 7 days (configurable)
6. All admins/users with `deleted_at` are considered deleted (soft delete)
7. Suspension is different from deletion (users can be unsuspended)

## First Admin Setup

```bash
# 1. Create in MongoDB (see ADMIN_SETUP_GUIDE.md)
# 2. Login with temporary password
# 3. Change password if needed
# 4. Use token for API calls
# 5. Create additional admins as needed
```

---

For more details, see the complete documentation in the `docs/` directory.
