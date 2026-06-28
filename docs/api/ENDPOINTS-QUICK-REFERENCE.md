# Authentication API - Quick Reference Guide

## All Endpoints Summary

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/auth/register` | ❌ | Register new user |
| `POST` | `/auth/login` | ❌ | Login and get tokens |
| `POST` | `/auth/refresh` | ❌ | Get new access token |
| `POST` | `/auth/logout` | ✅ | Logout and invalidate tokens |
| `GET` | `/profile` | ✅ | View current user profile |
| `PATCH` | `/profile` | ✅ | Update profile (name) |
| `POST` | `/profile/change-password` | ✅ | Change password |
| `DELETE` | `/profile` | ✅ | Delete account (soft delete) |

> ✅ = Requires JWT Bearer token in `Authorization` header
> ❌ = No authentication required

---

## Typical Flow

### New User
```
1. POST /auth/register
2. POST /auth/login
3. Use access_token for protected endpoints
```

### Token Expiration
```
1. Access token expires (after 1 hour)
2. POST /auth/refresh with refresh_token
3. Get new access_token, continue using
```

### Password Reset
```
1. POST /profile/change-password (with old + new password)
2. All sessions logged out automatically
3. POST /auth/login again with new password
```

### Account Deletion
```
1. DELETE /profile (with access_token)
2. All sessions logged out
3. Account marked as deleted
```

---

## Status Codes

| Code | Meaning |
|------|---------|
| `200` | ✅ Success |
| `201` | ✅ Created (registration) |
| `204` | ✅ No Content (logout, delete) |
| `400` | ❌ Bad Request (validation error) |
| `401` | ❌ Unauthorized (missing/invalid token) |
| `404` | ❌ Not Found (user doesn't exist) |
| `409` | ❌ Conflict (email already exists) |
| `500` | ❌ Internal Server Error |

---

## Error Messages

| Error | Code | Endpoint(s) |
|-------|------|------------|
| `invalid input` | 400 | Most endpoints |
| `invalid email` | 400 | Register |
| `invalid password` | 400 | Register, Change Password |
| `age verification required` | 400 | Register |
| `email already exists` | 409 | Register |
| `invalid credentials` | 401 | Login |
| `invalid or expired refresh token` | 401 | Refresh |
| `unauthorized` | 401 | Protected endpoints |
| `user not found` | 404 | Profile, Update, Change Password, Delete |
| `invalid user id` | 400 | Logout, Delete |

---

## Token Info

### Access Token
- **TTL:** 1 hour
- **Used For:** Authorization header in protected requests
- **Format:** `Authorization: Bearer {access_token}`

### Refresh Token
- **TTL:** 7 days
- **Used For:** Getting new access tokens
- **Storage:** Hashed in database, never exposed in responses after login

---

## Password Requirements

✅ Valid: `SecurePass123!`
- At least 8 characters
- 1+ uppercase letter (A-Z)
- 1+ lowercase letter (a-z)
- 1+ digit (0-9)
- 1+ special character (!@#$%^&*)

❌ Invalid Examples:
- `pass123` (no uppercase, no special char)
- `Pass123` (no special char)
- `Pass!` (too short, no digit)
- `PASSWORD123!` (no lowercase)

---

## Quick cURL Examples

### Register
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "SecurePass123!",
    "age_verified": true
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

### View Profile (using access token)
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer {access_token_here}"
```

### Update Profile
```bash
curl -X PATCH http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer {access_token_here}" \
  -H "Content-Type: application/json" \
  -d '{"name": "Jane Doe"}'
```

### Change Password
```bash
curl -X POST http://localhost:8080/api/v1/profile/change-password \
  -H "Authorization: Bearer {access_token_here}" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "OldPass123!",
    "new_password": "NewPass456!"
  }'
```

### Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "{refresh_token_here}"}'
```

### Logout
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer {access_token_here}"
```

### Delete Account
```bash
curl -X DELETE http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer {access_token_here}"
```

---

## Response Examples

### Successful Login Response
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNTA3ZjFmNzdiY2Y4NmNkNzk5NDM5MDExIiwiZW1haWwiOiJqb2huQGV4YW1wbGUuY29tIiwiZXhwIjoxNzE5NjY0NzIwLCJpYXQiOjE3MTk2NjExMjB9.abc123...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNTA3ZjFmNzdiY2Y4NmNkNzk5NDM5MDExIiwiZXhwIjoxNzI3NDQwNTIwLCJpYXQiOjE3MTk2NjExMjB9.xyz789...",
    "expires_in": 3600,
    "token_type": "Bearer"
  },
  "message": "login successful",
  "status": 200
}
```

### Error Response
```json
{
  "error": "invalid credentials",
  "status": 401
}
```

---

## Implementation Notes

### For Frontend Developers
1. Store `access_token` in memory or session storage
2. Store `refresh_token` in secure HTTP-only cookie
3. When access token expires (401 response), refresh automatically
4. Clear all tokens on logout
5. Redirect to login when 401 received

### For Mobile Developers
1. Store tokens securely (Keychain on iOS, Keystore on Android)
2. Implement token refresh before expiration
3. Handle 401 responses gracefully
4. Clear tokens on logout

### For API Testing
1. Use Postman/Insomnia collections
2. Save tokens in environment variables after login
3. Use `{{access_token}}` in protected request headers
4. Test error cases thoroughly

---

## Common Issues & Solutions

### Issue: `invalid credentials` on login
**Solution:** Check email and password are correct

### Issue: `unauthorized` on protected endpoint
**Solution:** Include `Authorization: Bearer {token}` header

### Issue: Token expired (401)
**Solution:** Use refresh token to get new access token

### Issue: `invalid or expired refresh token`
**Solution:** Login again to get new tokens

### Issue: Password change failed but showing success
**Solution:** Verify old password is correct

### Issue: Can't login after password change
**Solution:** Normal - all tokens invalidated, must login with new password

---

## Next Steps

For more detailed information, see:
- Full documentation: `docs/api/01-authentication.md`
- Testing guide: `docs/api/TESTING.md` (coming soon)
- Integration guide: `docs/api/INTEGRATION.md` (coming soon)
