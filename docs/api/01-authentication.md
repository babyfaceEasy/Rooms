# Authentication API Documentation

## Overview

The Authentication API provides endpoints for user registration, login, token refresh, logout, and profile management. All endpoints (except registration and login) require JWT authentication using Bearer tokens.

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication
Protected endpoints require the `Authorization` header with Bearer token:
```
Authorization: Bearer {access_token}
```

### Response Format
All responses are JSON with the following structure:

**Success Response:**
```json
{
  "data": {},
  "message": "optional message",
  "status": 200
}
```

**Error Response:**
```json
{
  "error": "error description",
  "status": 400
}
```

---

## 1. User Registration

### Endpoint
```
POST /auth/register
```

### Description
Register a new user account with name, email, password, and age verification. A verification email is sent to the provided email address — the user must verify their email before they can log in. The verification token expires after 24 hours.

### Request Headers
```
Content-Type: application/json
```

### Request Body
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePass123!",
  "age_verified": true
}
```

### Request Body Schema
| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | Yes | 2-100 characters |
| `email` | string | Yes | Valid email format |
| `password` | string | Yes | Min 8 chars, 1 uppercase, 1 lowercase, 1 digit, 1 special char |
| `age_verified` | boolean | Yes | Must be `true` (confirms user is 13+) |

### Response (201 Created)
```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "code": "87654321",
    "name": "John Doe",
    "email": "john@example.com",
    "age_verified": true,
    "created_at": "2024-06-28T15:30:00Z"
  },
  "message": "user registered successfully",
  "status": 201
}
```

### Response Fields
| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique user identifier (MongoDB ObjectID in hex format) |
| `code` | string | Unique 8-digit customer code (assigned automatically) |
| `name` | string | User's full name |
| `email` | string | User's email address |
| `age_verified` | boolean | Age verification status |
| `created_at` | string | ISO 8601 timestamp of account creation |

### Error Responses

**400 Bad Request - Missing Fields**
```json
{
  "error": "invalid input",
  "status": 400
}
```

**400 Bad Request - Invalid Email**
```json
{
  "error": "invalid email",
  "status": 400
}
```

**400 Bad Request - Weak Password**
```json
{
  "error": "invalid password",
  "status": 400
}
```

**400 Bad Request - Age Verification Missing**
```json
{
  "error": "age verification required",
  "status": 400
}
```

**409 Conflict - Email Already Exists**
```json
{
  "error": "email already exists",
  "status": 409
}
```

### Example cURL
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

---

---

## 2. Login

### Endpoint
```
POST /auth/login
```

### Description
Authenticate a user with email and password. Returns access and refresh tokens. Users must have verified their email address before they can log in — a 403 error is returned if the email is not yet verified.

### Request Headers
```
Content-Type: application/json
```

### Request Body
```json
{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

### Request Body Schema
| Field | Type | Required |
|-------|------|----------|
| `email` | string | Yes |
| `password` | string | Yes |

### Response (200 OK)
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "token_type": "Bearer"
  },
  "message": "login successful",
  "status": 200
}
```

### Response Schema
| Field | Type | Description |
|-------|------|-------------|
| `access_token` | string | JWT access token (expires in 1 hour) |
| `refresh_token` | string | JWT refresh token (expires in 7 days) |
| `expires_in` | integer | Access token TTL in seconds (3600) |
| `token_type` | string | Always "Bearer" |

### Error Responses

**400 Bad Request - Missing Credentials**
```json
{
  "error": "invalid input",
  "status": 400
}
```

**401 Unauthorized - Invalid Email or Password**
```json
{
  "error": "invalid credentials",
  "status": 401
}
```

**403 Forbidden - Email Not Verified**
```json
{
  "error": "please verify your email address before logging in",
  "status": 403
}
```

**404 Not Found - User Doesn't Exist**
```json
{
  "error": "invalid credentials",
  "status": 401
}
```

### Example cURL
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

---

## 3. Refresh Access Token

### Endpoint
```
POST /auth/refresh
```

### Description
Generate a new access token using a valid refresh token. Refresh tokens are rotated on each refresh.

### Request Headers
```
Content-Type: application/json
```

### Request Body
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Request Body Schema
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `refresh_token` | string | Yes | Valid refresh token from login or previous refresh |

### Response (200 OK)
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "token_type": "Bearer"
  },
  "message": "access token refreshed",
  "status": 200
}
```

### Error Responses

**400 Bad Request - Missing Refresh Token**
```json
{
  "error": "invalid input",
  "status": 400
}
```

**401 Unauthorized - Invalid Refresh Token**
```json
{
  "error": "invalid or expired refresh token",
  "status": 401
}
```

**401 Unauthorized - Expired Refresh Token**
```json
{
  "error": "invalid or expired refresh token",
  "status": 401
}
```

### Example cURL
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

---

## 4. Verify Email

### Endpoint
```
POST /auth/verify-email
```

### Description
Verify a user's email address using the verification token sent during registration. Users must verify their email before they can log in.

### Request Headers
```
Content-Type: application/json
```

### Request Body
```json
{
  "token": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
}
```

### Request Body Schema
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `token` | string | Yes | Verification token received in the registration email |

### Response (200 OK)
```json
{
  "message": "Email verified successfully"
}
```

### Error Responses

**400 Bad Request - Missing Token**
```json
{
  "error": "invalid input",
  "status": 400
}
```

**400 Bad Request - Token Expired**
```json
{
  "error": "invalid input",
  "status": 400
}
```

**400 Bad Request - Token Already Used**
```json
{
  "error": "invalid input",
  "status": 400
}
```

**400 Bad Request - Invalid Token**
```json
{
  "error": "invalid input",
  "status": 400
}
```

### Example cURL
```bash
curl -X POST http://localhost:8080/api/v1/auth/verify-email \
  -H "Content-Type: application/json" \
  -d '{
    "token": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  }'
```

### Notes
- Verification token expires after 24 hours
- Token is one-time use only
- If the token expires, the user must re-register or request support

---

## 5. Forgot Password

### Endpoint
```
POST /auth/forgot-password
```

### Description
Request a password reset email. If the email exists on an account, a reset token is generated and sent via email. Returns the same response regardless of whether the email exists (prevents email enumeration).

### Request Headers
```
Content-Type: application/json
```

### Request Body
```json
{
  "email": "john@example.com"
}
```

### Request Body Schema
| Field | Type | Required |
|-------|------|----------|
| `email` | string | Yes |

### Response (200 OK)
```json
{
  "message": "If an account with this email exists, a password reset link has been sent."
}
```

### Error Responses

**400 Bad Request - Missing Email**
```json
{
  "error": "invalid input",
  "status": 400
}
```

### Example cURL
```bash
curl -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com"
  }'
```

### Notes
- Always returns 200 to prevent email enumeration
- Reset token expires after 1 hour
- Sends email with `reset_url` and `reset_token` in dynamic template data

---

## 6. Reset Password

### Endpoint
```
POST /auth/reset-password
```

### Description
Reset a user's password using a valid reset token received via email. Invalidates all refresh tokens, logging out all devices.

### Request Headers
```
Content-Type: application/json
```

### Request Body
```json
{
  "token": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "new_password": "NewSecurePass456!"
}
```

### Request Body Schema
| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `token` | string | Yes | Reset token from email |
| `new_password` | string | Yes | Min 8 chars, 1 uppercase, 1 lowercase, 1 digit, 1 special char |

### Response (200 OK)
```json
{
  "message": "Password has been reset successfully. Please log in with your new password."
}
```

### Error Responses

**400 Bad Request - Missing Fields**
```json
{
  "error": "invalid input",
  "status": 400
}
```

**400 Bad Request - Token Expired**
```json
{
  "error": "invalid input",
  "status": 400
}
```

**400 Bad Request - Token Already Used**
```json
{
  "error": "invalid input",
  "status": 400
}
```

**400 Bad Request - Invalid Token**
```json
{
  "error": "invalid input",
  "status": 400
}
```

**400 Bad Request - Weak Password**
```json
{
  "error": "invalid password",
  "status": 400
}
```

### Important Notes
⚠️ **All refresh tokens are immediately invalidated.** User must login again with the new password on all devices.

### Example cURL
```bash
curl -X POST http://localhost:8080/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
    "new_password": "NewSecurePass456!"
  }'
```

---

## 7. Logout

### Endpoint
```
POST /auth/logout
```

### Description
Invalidate all refresh tokens for the current user. The access token remains valid until expiration but cannot be refreshed. **All user sessions across all devices are terminated.**

### Request Headers
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

### Request Body
No request body required (empty JSON object acceptable).

### Response (204 No Content)
```
(empty response body)
```

### Error Responses

**401 Unauthorized - Missing or Invalid Token**
```json
{
  "error": "unauthorized",
  "status": 401
}
```

**400 Bad Request - Invalid User ID**
```json
{
  "error": "invalid user id",
  "status": 400
}
```

### Example cURL
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

---

## 8. View Profile

### Endpoint
```
GET /profile
```

### Description
Retrieve the current user's profile information (name and email).

### Request Headers
```
Authorization: Bearer {access_token}
```

### Request Body
No request body.

### Response (200 OK)
```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "code": "87654321",
    "name": "John Doe",
    "email": "john@example.com"
  },
  "message": "profile retrieved successfully",
  "status": 200
}
```

### Response Fields
| Field | Type | Description |
|-------|------|-------------|
| `id` | string | User's unique identifier |
| `code` | string | User's unique 8-digit customer code |
| `name` | string | User's full name |
| `email` | string | User's email address |

### Error Responses

**401 Unauthorized - Missing or Invalid Token**
```json
{
  "error": "unauthorized",
  "status": 401
}
```

**404 Not Found - User Doesn't Exist**
```json
{
  "error": "user not found",
  "status": 404
}
```

### Example cURL
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## 9. Update Profile

### Endpoint
```
PATCH /profile
```

### Description
Update the current user's profile. Currently supports updating the name only.

### Request Headers
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

### Request Body
```json
{
  "name": "John Smith"
}
```

### Request Body Schema
| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `name` | string | Yes | 2-100 characters |

### Response (200 OK)
```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "code": "87654321",
    "name": "John Smith",
    "email": "john@example.com"
  },
  "message": "profile updated successfully",
  "status": 200
}
```

### Error Responses

**400 Bad Request - Invalid Name**
```json
{
  "error": "invalid input",
  "status": 400
}
```

**401 Unauthorized - Missing or Invalid Token**
```json
{
  "error": "unauthorized",
  "status": 401
}
```

**404 Not Found - User Doesn't Exist**
```json
{
  "error": "user not found",
  "status": 404
}
```

### Example cURL
```bash
curl -X PATCH http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Smith"
  }'
```

---

## 10. Change Password

### Endpoint
```
POST /profile/change-password
```

### Description
Change the current user's password. **All refresh tokens are invalidated, logging out all devices.** User must login again with the new password.

### Request Headers
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

### Request Body
```json
{
  "current_password": "OldSecurePass123!",
  "new_password": "NewSecurePass456!"
}
```

### Request Body Schema
| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `current_password` | string | Yes | User's current password |
| `new_password` | string | Yes | Min 8 chars, 1 uppercase, 1 lowercase, 1 digit, 1 special char |

### Response (200 OK)
```json
{
  "data": {
    "message": "password changed successfully, please log in again"
  },
  "message": "password changed successfully",
  "status": 200
}
```

### Error Responses

**400 Bad Request - Missing Fields**
```json
{
  "error": "invalid input",
  "status": 400
}
```

**400 Bad Request - Invalid Current Password**
```json
{
  "error": "invalid password",
  "status": 400
}
```

**400 Bad Request - Weak New Password**
```json
{
  "error": "invalid password",
  "status": 400
}
```

**400 Bad Request - Same as Current Password**
```json
{
  "error": "invalid password",
  "status": 400
}
```

**401 Unauthorized - Missing or Invalid Token**
```json
{
  "error": "unauthorized",
  "status": 401
}
```

**404 Not Found - User Doesn't Exist**
```json
{
  "error": "user not found",
  "status": 404
}
```

### Important Notes
⚠️ **All refresh tokens are immediately invalidated.** Users must login again with the new password on all devices.

### Example cURL
```bash
curl -X POST http://localhost:8080/api/v1/profile/change-password \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "OldSecurePass123!",
    "new_password": "NewSecurePass456!"
  }'
```

---

## 11. Delete Account

### Endpoint
```
DELETE /profile
```

### Description
Permanently delete the current user's account. The account is soft-deleted (marked as deleted but preserved in the database for compliance). **All refresh tokens are invalidated.** The user cannot login after account deletion.

### Request Headers
```
Authorization: Bearer {access_token}
```

### Request Body
No request body required (empty JSON object acceptable).

### Response (204 No Content)
```
(empty response body)
```

### Error Responses

**400 Bad Request - Invalid User ID**
```json
{
  "error": "invalid user id",
  "status": 400
}
```

**401 Unauthorized - Missing or Invalid Token**
```json
{
  "error": "unauthorized",
  "status": 401
}
```

**404 Not Found - User Doesn't Exist**
```json
{
  "error": "user not found",
  "status": 404
}
```

### Important Notes
⚠️ **This action is permanent.** Accounts are soft-deleted (marked deleted but data preserved).

⚠️ **All refresh tokens are immediately invalidated** across all devices.

### Example cURL
```bash
curl -X DELETE http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## JWT Token Details

### Access Token
- **Type:** JWT (JSON Web Token)
- **Algorithm:** HS256
- **TTL:** 1 hour (3600 seconds)
- **Claims:**
  - `user_id` — ObjectID of the user
  - `email` — User's email address
  - `exp` — Expiration time (Unix timestamp)
  - `iat` — Issued at time (Unix timestamp)

### Refresh Token
- **Type:** JWT (JSON Web Token)
- **Algorithm:** HS256
- **TTL:** 7 days (604800 seconds)
- **Storage:** Hashed with SHA256 in database (plaintext never stored)
- **Rotation:** New refresh token generated on each refresh operation
- **Claims:**
  - `user_id` — ObjectID of the user
  - `exp` — Expiration time (Unix timestamp)
  - `iat` — Issued at time (Unix timestamp)

---

## Security Considerations

### Password Requirements
Passwords must meet the following criteria:
- Minimum 8 characters
- At least 1 uppercase letter (A-Z)
- At least 1 lowercase letter (a-z)
- At least 1 digit (0-9)
- At least 1 special character (!@#$%^&*)

**Example valid password:** `SecurePass123!`

### Token Security
- Tokens should be sent over **HTTPS only** in production
- Store access tokens in memory or short-lived session storage (NOT localStorage for web apps)
- Store refresh tokens in secure HTTP-only cookies (NOT localStorage)
- Do not expose tokens in URLs or query parameters

### Multi-Device Logout
- Password change invalidates all refresh tokens across all devices
- Account deletion invalidates all refresh tokens across all devices
- Logout invalidates all refresh tokens for the current user

---

## Common Use Cases

### Registration → Login Flow
```
1. POST /auth/register → Get user ID (verification email sent)
2. POST /auth/verify-email → Verify email with token
3. POST /auth/login → Get access + refresh tokens
4. Use access_token in Authorization header for protected endpoints
```

### Password Reset Flow
```
1. POST /auth/forgot-password with email
2. Check email for reset link (includes token)
3. POST /auth/reset-password with token + new password
4. All refresh tokens invalidated
5. POST /auth/login again with new password
```

### Token Refresh Flow
```
1. Access token is about to expire
2. POST /auth/refresh with refresh_token
3. Get new access_token (refresh_token also rotated)
4. Continue using new access_token
```

### Logout Flow
```
1. POST /auth/logout with access_token
2. All refresh tokens invalidated
3. User must login again to get new tokens
```

### Password Change Flow
```
1. POST /profile/change-password with old + new password
2. All refresh tokens invalidated immediately
3. User must login again with new password on all devices
```

### Account Deletion Flow
```
1. DELETE /profile with access_token
2. Account soft-deleted
3. All refresh tokens invalidated
4. User cannot login
```

---

## Rate Limiting

Currently not implemented. Consider adding in production:
- Login: 5 attempts per 15 minutes per IP
- Registration: 3 per hour per IP
- Password change: 3 per hour per user

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

## Support & Questions

For issues or questions:
1. Check the error messages in the response
2. Verify token validity using `Authorization: Bearer {token}`
3. Ensure Content-Type is `application/json`
4. Contact the development team
