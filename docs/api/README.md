# API Documentation

This folder contains comprehensive API documentation for the backend services.

## Contents

### 📖 [01-authentication.md](./01-authentication.md)
Complete authentication API reference with all endpoints, request/response examples, error handling, and security details.

**Includes:**
- User registration with validation
- JWT-based login and token refresh
- Logout and token invalidation
- Profile management (view, update, delete)
- Password changes with multi-device logout
- Account deletion with soft delete
- Token details and claims
- Security considerations
- Common use cases

**All 8 endpoints documented:**
1. `POST /auth/register` — Create new account
2. `POST /auth/login` — Authenticate and get tokens
3. `POST /auth/refresh` — Get new access token
4. `POST /auth/logout` — Invalidate all tokens
5. `GET /profile` — View current user profile
6. `PATCH /profile` — Update profile (name)
7. `POST /profile/change-password` — Change password
8. `DELETE /profile` — Delete account (soft delete)

### 🚀 [ENDPOINTS-QUICK-REFERENCE.md](./ENDPOINTS-QUICK-REFERENCE.md)
Quick lookup guide for all endpoints with summary tables, cURL examples, and common issues.

**Includes:**
- All endpoints summary table
- Typical user flows
- Status codes reference
- Error messages lookup
- Token information
- Password requirements
- Quick cURL examples
- Response examples

### 🧪 [TESTING.md](./TESTING.md)
Complete testing guide for manual and automated testing of all endpoints.

**Includes:**
- Full user lifecycle test with cURL
- Error testing scenarios
- Postman collection setup
- Automated test commands
- Performance testing examples
- Debugging tips
- Common test scenarios
- Test data and edge cases

---

## Quick Start

### 1. Register a User
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

### 2. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

### 3. Use Access Token
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer {access_token}"
```

---

## Which Document Should I Read?

| Question | Document |
|----------|----------|
| "What are all the endpoints?" | [ENDPOINTS-QUICK-REFERENCE.md](./ENDPOINTS-QUICK-REFERENCE.md) |
| "How do I use endpoint X?" | [01-authentication.md](./01-authentication.md) |
| "What's the full request/response for X?" | [01-authentication.md](./01-authentication.md) |
| "How do I test the API?" | [TESTING.md](./TESTING.md) |
| "How do I use Postman?" | [TESTING.md](./TESTING.md) |
| "What error codes exist?" | [ENDPOINTS-QUICK-REFERENCE.md](./ENDPOINTS-QUICK-REFERENCE.md) or [01-authentication.md](./01-authentication.md) |
| "How do tokens work?" | [01-authentication.md](./01-authentication.md) |
| "What are password requirements?" | [ENDPOINTS-QUICK-REFERENCE.md](./ENDPOINTS-QUICK-REFERENCE.md) or [01-authentication.md](./01-authentication.md) |

---

## Key Features

### 🔐 Security
- **JWT Authentication** — HS256 algorithm, Bearer tokens
- **Bcrypt Password Hashing** — Industry-standard password security
- **Token Expiration** — 1 hour access tokens, 7 day refresh tokens
- **Soft Delete** — Account deletion preserves data for compliance
- **Multi-Device Logout** — Password change invalidates all sessions

### 📊 Complete Validation
- Email format validation
- Strong password requirements (min 8 chars, uppercase, lowercase, digit, special char)
- Age verification (13+ years old)
- Name length validation (2-100 characters)

### 🚀 Production Ready
- Comprehensive error handling with specific error codes
- Full test coverage (83+ tests)
- Proper HTTP status codes
- No sensitive data in responses
- Refresh token rotation on each refresh

### 📱 Multi-Device Support
- Multiple simultaneous login sessions
- Per-device token management
- Mass logout on password change or account deletion
- Independent token refresh per device

---

## API Response Format

### Success Response
```json
{
  "data": {
    // Response payload
  },
  "message": "human-readable message",
  "status": 200
}
```

### Error Response
```json
{
  "error": "error description",
  "status": 400
}
```

---

## Authentication

All protected endpoints require:
```
Authorization: Bearer {access_token}
```

**Token Details:**
- **Type:** JWT (JSON Web Token)
- **Algorithm:** HS256
- **TTL:** 1 hour (3600 seconds)
- **Claims:** user_id, email, exp, iat

---

## Environment Variables

```bash
# JWT Configuration
JWT_SECRET=your-super-secret-key-min-32-characters-long
ACCESS_TOKEN_TTL=1h
REFRESH_TOKEN_TTL=168h

# Database
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=rooms
```

---

## Status Codes

| Code | Meaning | When |
|------|---------|------|
| 200 | OK | Successful request |
| 201 | Created | Successful registration |
| 204 | No Content | Logout, delete (no body response) |
| 400 | Bad Request | Validation error |
| 401 | Unauthorized | Missing/invalid token |
| 404 | Not Found | User doesn't exist |
| 409 | Conflict | Email already exists |
| 500 | Server Error | Unexpected error |

---

## Endpoints by Category

### 🔑 Authentication
- `POST /auth/register` — Create account
- `POST /auth/login` — Login
- `POST /auth/refresh` — Refresh token
- `POST /auth/logout` — Logout

### 👤 Profile Management
- `GET /profile` — View profile
- `PATCH /profile` — Update profile
- `POST /profile/change-password` — Change password
- `DELETE /profile` — Delete account

---

## Common Workflows

### New User Registration and First Login
```
1. POST /auth/register → User created
2. POST /auth/login → Get tokens
3. GET /profile → Verify profile
```

### Token Refresh (Access Token Expired)
```
1. POST /auth/refresh → New access token
2. Continue using new token
```

### Password Change
```
1. POST /profile/change-password → Password updated
2. All tokens invalidated
3. POST /auth/login → New tokens required
```

### Account Deletion
```
1. DELETE /profile → Account soft-deleted
2. All tokens invalidated
3. Cannot login anymore
```

---

## Contact & Support

- 📧 Email: support@example.com
- 🐛 Issues: GitHub Issues
- 📚 Docs: This folder
- 🧪 Tests: `go test ./...`

---

## Version

**API Version:** v1
**Last Updated:** June 28, 2024
**Status:** Production Ready ✅

---

## License

[Your License Here]
