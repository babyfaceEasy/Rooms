# Authentication API - Testing Guide

## Manual Testing with cURL

### Setup
```bash
# Set base URL
BASE_URL="http://localhost:8080/api/v1"

# Or for production
BASE_URL="https://your-api.com/api/v1"
```

---

## Full User Lifecycle Test

### Step 1: Register a New User
```bash
curl -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "TestPass123!",
    "age_verified": true
  }'
```

**Expected Response (201 Created):**
```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "code": "87654321",
    "name": "Test User",
    "email": "test@example.com",
    "age_verified": true,
    "created_at": "2024-06-28T15:30:00Z"
  },
  "message": "user registered successfully",
  "status": 201
}
```

### Step 2: Login
```bash
curl -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123!"
  }'
```

**Expected Response (200 OK):**
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

**Save tokens for next steps:**
```bash
ACCESS_TOKEN="your_access_token_here"
REFRESH_TOKEN="your_refresh_token_here"
```

### Step 3: View Profile
```bash
curl -X GET $BASE_URL/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "code": "87654321",
    "name": "Test User",
    "email": "test@example.com"
  },
  "message": "profile retrieved successfully",
  "status": 200
}
```

### Step 4: Update Profile
```bash
curl -X PATCH $BASE_URL/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Test User"
  }'
```

**Expected Response (200 OK):**
```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "code": "87654321",
    "name": "Updated Test User",
    "email": "test@example.com"
  },
  "message": "profile updated successfully",
  "status": 200
}
```

### Step 5: Change Password
```bash
curl -X POST $BASE_URL/profile/change-password \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "TestPass123!",
    "new_password": "NewTestPass456!"
  }'
```

**Expected Response (200 OK):**
```json
{
  "data": {
    "message": "password changed successfully, please log in again"
  },
  "message": "password changed successfully",
  "status": 200
}
```

**⚠️ Important:** All refresh tokens are now invalid. Must login with new password.

### Step 6: Login with New Password
```bash
curl -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "NewTestPass456!"
  }'
```

**Expected Response (200 OK):** New tokens generated

### Step 7: Refresh Access Token
```bash
curl -X POST $BASE_URL/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "'$REFRESH_TOKEN'"
  }'
```

**Expected Response (200 OK):**
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

### Step 8: Logout
```bash
curl -X POST $BASE_URL/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Expected Response (204 No Content):** Empty body

### Step 9: Verify Can't Use Refresh Token
```bash
curl -X POST $BASE_URL/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "'$REFRESH_TOKEN'"
  }'
```

**Expected Response (401 Unauthorized):**
```json
{
  "error": "invalid or expired refresh token",
  "status": 401
}
```

### Step 10: Delete Account
First, login again:
```bash
curl -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "NewTestPass456!"
  }'
```

Then delete:
```bash
curl -X DELETE $BASE_URL/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Expected Response (204 No Content):** Empty body

### Step 11: Verify Can't Login After Deletion
```bash
curl -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "NewTestPass456!"
  }'
```

**Expected Response (401 Unauthorized):**
```json
{
  "error": "invalid credentials",
  "status": 401
}
```

---

## Error Testing

### Test 1: Register with Invalid Email
```bash
curl -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "invalid-email",
    "password": "TestPass123!",
    "age_verified": true
  }'
```

**Expected (400):** `invalid email`

### Test 2: Register with Weak Password
```bash
curl -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "weak",
    "age_verified": true
  }'
```

**Expected (400):** `invalid password`

### Test 3: Register Without Age Verification
```bash
curl -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "TestPass123!",
    "age_verified": false
  }'
```

**Expected (400):** `age verification required`

### Test 4: Login with Wrong Password
```bash
curl -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "WrongPassword123!"
  }'
```

**Expected (401):** `invalid credentials`

### Test 5: Login with Non-existent Email
```bash
curl -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "nonexistent@example.com",
    "password": "TestPass123!"
  }'
```

**Expected (401):** `invalid credentials`

### Test 6: Access Protected Endpoint Without Token
```bash
curl -X GET $BASE_URL/profile
```

**Expected (401):** `unauthorized`

### Test 7: Access Protected Endpoint with Invalid Token
```bash
curl -X GET $BASE_URL/profile \
  -H "Authorization: Bearer invalid_token_here"
```

**Expected (401):** `unauthorized`

### Test 8: Update Profile with Invalid Name
```bash
curl -X PATCH $BASE_URL/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "A"
  }'
```

**Expected (400):** `invalid input`

### Test 9: Change Password with Wrong Current Password
```bash
curl -X POST $BASE_URL/profile/change-password \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "WrongPassword123!",
    "new_password": "NewPass456!"
  }'
```

**Expected (400):** `invalid password`

### Test 10: Change Password with Weak New Password
```bash
curl -X POST $BASE_URL/profile/change-password \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "CurrentPass123!",
    "new_password": "weak"
  }'
```

**Expected (400):** `invalid password`

---

## Postman Collection

### Import to Postman
1. Create new collection "Auth API"
2. Add the following requests with variables

### Variables (Set in collection)
```
{{base_url}} = http://localhost:8080/api/v1
{{access_token}} = (set after login)
{{refresh_token}} = (set after login)
{{user_email}} = test@example.com
{{user_password}} = TestPass123!
{{new_password}} = NewTestPass456!
```

### Requests

#### POST /auth/register
```
URL: {{base_url}}/auth/register
Method: POST
Headers:
  Content-Type: application/json

Body (JSON):
{
  "name": "Test User",
  "email": "{{user_email}}",
  "password": "{{user_password}}",
  "age_verified": true
}
```

#### POST /auth/login
```
URL: {{base_url}}/auth/login
Method: POST
Headers:
  Content-Type: application/json

Body (JSON):
{
  "email": "{{user_email}}",
  "password": "{{user_password}}"
}

Post-request Script:
var jsonData = pm.response.json();
pm.environment.set("access_token", jsonData.data.access_token);
pm.environment.set("refresh_token", jsonData.data.refresh_token);
```

#### GET /profile
```
URL: {{base_url}}/profile
Method: GET
Headers:
  Authorization: Bearer {{access_token}}
```

#### PATCH /profile
```
URL: {{base_url}}/profile
Method: PATCH
Headers:
  Authorization: Bearer {{access_token}}
  Content-Type: application/json

Body (JSON):
{
  "name": "Updated Name"
}
```

#### POST /profile/change-password
```
URL: {{base_url}}/profile/change-password
Method: POST
Headers:
  Authorization: Bearer {{access_token}}
  Content-Type: application/json

Body (JSON):
{
  "current_password": "{{user_password}}",
  "new_password": "{{new_password}}"
}
```

#### POST /auth/refresh
```
URL: {{base_url}}/auth/refresh
Method: POST
Headers:
  Content-Type: application/json

Body (JSON):
{
  "refresh_token": "{{refresh_token}}"
}

Post-request Script:
var jsonData = pm.response.json();
pm.environment.set("access_token", jsonData.data.access_token);
```

#### POST /auth/logout
```
URL: {{base_url}}/auth/logout
Method: POST
Headers:
  Authorization: Bearer {{access_token}}
```

#### DELETE /profile
```
URL: {{base_url}}/profile
Method: DELETE
Headers:
  Authorization: Bearer {{access_token}}
```

---

## Automated Testing

### Unit Tests
```bash
cd /path/to/backend
go test -v ./internal/service -run TestDeleteAccount
go test -v ./internal/handler -run TestDeleteAccount
go test -v ./internal/middleware
```

### All Tests
```bash
go test -v ./...
```

### With Coverage
```bash
go test -cover ./...
```

---

## Performance Testing

### Load Test with Apache Bench
```bash
# Register 100 users
ab -n 100 -c 10 -p register.json \
  -T application/json \
  http://localhost:8080/api/v1/auth/register

# Login test
ab -n 1000 -c 50 -p login.json \
  -T application/json \
  http://localhost:8080/api/v1/auth/login
```

### Load Test with Apache JMeter
1. Create test plan with HTTP requests
2. Add samplers for each endpoint
3. Use login response tokens in subsequent requests
4. Run with desired thread count and ramp-up

---

## Debugging Tips

### Check Token Claims
```bash
# Decode JWT (without verification, for inspection only)
# Use https://jwt.io or decode online

# Or with jq
echo $ACCESS_TOKEN | cut -d'.' -f2 | base64 -d | jq .
```

### Check Response Headers
```bash
curl -i -X GET $BASE_URL/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Enable Verbose Output
```bash
curl -v -X GET $BASE_URL/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Test with Different Content-Type
```bash
curl -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json; charset=utf-8" \
  -d '{...}'
```

---

## Common Test Scenarios

### Scenario 1: Multi-Device Login
```
1. Login on device A → Token A
2. Login on device B → Token B
3. Both tokens should work independently
4. Change password on device A
5. Both tokens should be invalidated
6. Both devices must login again
```

### Scenario 2: Token Expiration
```
1. Login → get access token
2. Wait for 1 hour (or mock time)
3. Try to use access token → should fail
4. Refresh with refresh token → should succeed
5. Continue using new access token
```

### Scenario 3: Session Recovery
```
1. Login → tokens
2. Logout
3. All refresh tokens deleted
4. Try to refresh → should fail
5. Login again → new tokens
```

### Scenario 4: Account Lifecycle
```
1. Register
2. Login
3. Use features
4. Change password
5. Delete account
6. Verify can't login
```

---

## Test Data

### Valid Test Credentials
- Email: `test@example.com`
- Password: `TestPass123!`
- Name: `Test User`

### Invalid Test Data
- Email: `invalid-email`
- Password: `weak` (too short, no special char)
- Name: `A` (too short)

### Edge Cases
- Email with special characters: `test+tag@example.com`
- Very long email (255 chars max)
- Name with spaces: `Test User Name`
- Name with special characters: `José García`
