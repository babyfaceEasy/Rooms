# Local Admin Endpoint Testing Guide

This guide will help you set up and test the admin endpoints locally.

## Prerequisites

- Docker & Docker Compose installed
- Go 1.25+ (for running without Docker)
- curl or Postman for testing API endpoints
- MongoDB and MinIO will run in Docker containers

---

## Step 1: Start Docker Services

### Start MongoDB and MinIO

```bash
cd /Users/killacam/Documents/work/bouk/temp_backend

# Start only MongoDB and MinIO (without the API)
docker-compose up -d mongodb minio createbuckets

# Verify services are running
docker-compose ps
```

**Expected Output:**
```
CONTAINER ID   IMAGE                  STATUS              NAMES
xxx            mongo:7                Up (healthy)        temp_backend_mongo
xxx            minio/minio:latest     Up (healthy)        temp_backend_minio
xxx            minio/mc:latest        Exited (0)          temp_backend_createbuckets
```

### Access Information

- **MongoDB**: `mongodb://localhost:27017`
  - Username: `admin`
  - Password: `secret`
  - Database: `temp_backend`

- **MinIO Console**: http://localhost:9001
  - Username: `minioadmin`
  - Password: `minioadmin`

- **MinIO API**: http://localhost:9000

---

## Step 2: Setup Environment File

Copy the example environment file:

```bash
cp .env.example .env
```

Update `.env` with these key values (for local development):

```env
APP_NAME=temp_backend
PORT=8080
APP_ENV=development
LOG_LEVEL=debug

MONGO_URI=mongodb://admin:secret@localhost:27017/temp_backend?authSource=admin
MONGO_DATABASE=temp_backend
MONGO_CONNECT_TIMEOUT=10s

S3_ENDPOINT_URL=http://localhost:9000
S3_REGION=us-east-1
S3_BUCKET=temp-bucket
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_USE_SSL=false
S3_PRESIGNED_EXPIRY=15m
S3_PUBLIC_URL=http://localhost:9000

JWT_SECRET=LKBSSPsjj3jZSQ3TuYuGVVAmiyu1CAB4OxouAW2fZRU
ACCESS_TOKEN_TTL=1h
REFRESH_TOKEN_TTL=168h

ADMIN_ACCESS_TOKEN_TTL=30m
ADMIN_REFRESH_TOKEN_TTL=7d
ADMIN_MAX_LOGIN_ATTEMPTS=5
ADMIN_LOCKOUT_DURATION=15m

SENDGRID_ENABLED=false
FIREBASE_ENABLED=false
APPSYNC_ENABLED=false
```

---

## Step 3: Run the Backend Locally

### Option A: Using `go run` (Recommended for Development)

```bash
cd /Users/killacam/Documents/work/bouk/temp_backend

# Build (optional, if not already built)
go build -o temp_backend cmd/api/main.go

# Run the application
go run cmd/api/main.go
```

**Expected Output:**
```
2024/11/12 11:30:00 Starting server on :8080
2024/11/12 11:30:00 Connected to MongoDB
2024/11/12 11:30:00 Connected to S3
```

### Option B: Using Docker Compose (Entire Stack)

```bash
# Start all services (MongoDB, MinIO, and API)
docker-compose up -d

# View logs
docker-compose logs -f api
```

---

## Step 4: Test Admin Endpoints

### 4a. Create an Admin User in MongoDB

First, you need an admin user in the database. Connect to MongoDB and insert an admin:

```bash
# Connect to MongoDB
mongosh "mongodb://admin:secret@localhost:27017/temp_backend?authSource=admin"
```

Then in the MongoDB shell:

```javascript
use temp_backend

db.admins.insertOne({
  "_id": ObjectId("507f1f77bcf86cd799439011"),
  "email": "admin@example.com",
  "password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86xVXAaxe9K", // bcrypt hash of "password123"
  "name": "Test Admin",
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
  "failed_login_attempts": 0,
  "created_at": new Date(),
  "updated_at": new Date()
})

exit
```

### 4b: Admin Login (Get Access Token)

```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }'
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer"
}
```

**Save the `access_token` for the next tests.**

---

### 4c: Get Admin Profile (NEW ENDPOINT)

```bash
# Replace YOUR_ACCESS_TOKEN with the token from login
curl -X GET http://localhost:8080/api/v1/admin/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json"
```

**Response (200 OK):**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "email": "admin@example.com",
  "name": "Test Admin",
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
  "failed_login_attempts": 0,
  "created_at": "2024-11-12T11:26:43.431Z",
  "updated_at": "2024-11-12T11:26:43.431Z"
}
```

---

### 4d: Test Error Cases

#### Missing Authorization Header

```bash
curl -X GET http://localhost:8080/api/v1/admin/profile \
  -H "Content-Type: application/json"
```

**Response (401 Unauthorized):**
```json
{
  "error": "missing authorization header"
}
```

#### Invalid Token

```bash
curl -X GET http://localhost:8080/api/v1/admin/profile \
  -H "Authorization: Bearer invalid_token_here" \
  -H "Content-Type: application/json"
```

**Response (401 Unauthorized):**
```json
{
  "error": "invalid token"
}
```

---

## Step 5: Full Testing Script

Create a file called `test_admin_endpoints.sh`:

```bash
#!/bin/bash

set -e

API_URL="http://localhost:8080"
ADMIN_EMAIL="admin@example.com"
ADMIN_PASSWORD="password123"

echo "════════════════════════════════════════"
echo "     Admin Endpoints Testing Script"
echo "════════════════════════════════════════"
echo ""

# Step 1: Login
echo "1️⃣  Testing Admin Login..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/admin/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$ADMIN_EMAIL\",
    \"password\": \"$ADMIN_PASSWORD\"
  }")

ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.access_token')
REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.refresh_token')

if [ "$ACCESS_TOKEN" = "null" ] || [ -z "$ACCESS_TOKEN" ]; then
  echo "❌ Login failed!"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo "✅ Login successful!"
echo "Access Token: ${ACCESS_TOKEN:0:50}..."
echo ""

# Step 2: Get Admin Profile
echo "2️⃣  Testing Get Admin Profile..."
PROFILE_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/admin/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json")

ADMIN_ID=$(echo $PROFILE_RESPONSE | jq -r '.id')
ADMIN_NAME=$(echo $PROFILE_RESPONSE | jq -r '.name')
ADMIN_ROLE=$(echo $PROFILE_RESPONSE | jq -r '.role')

if [ "$ADMIN_ID" = "null" ] || [ -z "$ADMIN_ID" ]; then
  echo "❌ Get profile failed!"
  echo "Response: $PROFILE_RESPONSE"
  exit 1
fi

echo "✅ Get admin profile successful!"
echo "Admin ID: $ADMIN_ID"
echo "Admin Name: $ADMIN_NAME"
echo "Admin Role: $ADMIN_ROLE"
echo ""

# Step 3: Verify sensitive fields are removed
echo "3️⃣  Verifying sensitive fields are removed..."
PASSWORD_HASH=$(echo $PROFILE_RESPONSE | jq -r '.password_hash')
SETUP_TOKEN=$(echo $PROFILE_RESPONSE | jq -r '.setup_token')

if [ "$PASSWORD_HASH" != "null" ] && [ -n "$PASSWORD_HASH" ]; then
  echo "❌ Password hash should be removed!"
  exit 1
fi

if [ "$SETUP_TOKEN" != "null" ] && [ -n "$SETUP_TOKEN" ]; then
  echo "❌ Setup token should be removed!"
  exit 1
fi

echo "✅ Sensitive fields properly removed!"
echo ""

# Step 4: Test invalid token
echo "4️⃣  Testing with invalid token..."
INVALID_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/admin/profile" \
  -H "Authorization: Bearer invalid_token" \
  -H "Content-Type: application/json")

ERROR=$(echo $INVALID_RESPONSE | jq -r '.error')

if [ "$ERROR" = "null" ] || [ -z "$ERROR" ]; then
  echo "❌ Should have returned error for invalid token!"
  exit 1
fi

echo "✅ Invalid token properly rejected!"
echo "Error: $ERROR"
echo ""

# Step 5: Test missing auth header
echo "5️⃣  Testing without authorization header..."
NO_AUTH_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/admin/profile" \
  -H "Content-Type: application/json")

ERROR=$(echo $NO_AUTH_RESPONSE | jq -r '.error')

if [ "$ERROR" = "null" ] || [ -z "$ERROR" ]; then
  echo "❌ Should have returned error for missing auth!"
  exit 1
fi

echo "✅ Missing auth header properly rejected!"
echo "Error: $ERROR"
echo ""

echo "════════════════════════════════════════"
echo "✅ All tests passed!"
echo "════════════════════════════════════════"
```

Make it executable and run:

```bash
chmod +x test_admin_endpoints.sh
./test_admin_endpoints.sh
```

---

## Troubleshooting

### MongoDB Connection Refused

```bash
# Check if MongoDB is running
docker-compose ps mongodb

# If not running, start it
docker-compose up -d mongodb
```

### MinIO Connection Issues

```bash
# Check MinIO status
docker-compose ps minio

# View MinIO logs
docker-compose logs minio

# Recreate the bucket
docker-compose restart createbuckets
```

### Token Expired

If you get "invalid token" error, the token might have expired (default: 30 minutes).
Generate a new token by logging in again:

```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'
```

### API Not Starting

Check logs:

```bash
# If running with go run
# Look at console output

# If using Docker Compose
docker-compose logs api
```

Common issues:
- MongoDB not connected: Check MongoDB container is healthy
- Wrong JWT_SECRET: Ensure it's at least 32 characters
- Port 8080 already in use: Change PORT in .env

---

## Using Postman for Testing

### 1. Create a Postman Collection

Create a new request in Postman:

**Request 1: Admin Login**
- Method: `POST`
- URL: `http://localhost:8080/api/v1/admin/auth/login`
- Body (JSON):
  ```json
  {
    "email": "admin@example.com",
    "password": "password123"
  }
  ```
- Tests (add this to automatically save the token):
  ```javascript
  var jsonData = pm.response.json();
  pm.environment.set("admin_access_token", jsonData.access_token);
  pm.environment.set("admin_refresh_token", jsonData.refresh_token);
  ```

**Request 2: Get Admin Profile**
- Method: `GET`
- URL: `http://localhost:8080/api/v1/admin/profile`
- Headers:
  - Key: `Authorization`
  - Value: `Bearer {{admin_access_token}}`

### 2. Run Requests in Sequence

In Postman, use the "Run" feature or trigger requests manually. The token will be automatically extracted and used for subsequent requests.

---

## Summary

1. ✅ Start Docker services: `docker-compose up -d mongodb minio createbuckets`
2. ✅ Setup `.env` file with correct configuration
3. ✅ Run backend: `go run cmd/api/main.go`
4. ✅ Create admin user in MongoDB
5. ✅ Login to get access token
6. ✅ Test the new `GET /api/v1/admin/profile` endpoint
7. ✅ Run error case tests to verify error handling

---

## API Endpoint Reference

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/admin/auth/login` | ❌ | Admin login |
| GET | `/api/v1/admin/profile` | ✅ | **[NEW]** Get current admin profile |
| POST | `/api/v1/admin/auth/logout` | ✅ | Admin logout |
| POST | `/api/v1/admin/auth/refresh` | ❌ | Refresh access token |
| GET | `/api/v1/admin/admins` | ✅ | List all admins |
| GET | `/api/v1/admin/admins/:id` | ✅ | Get specific admin |
| POST | `/api/v1/admin/admins` | ✅ | Create new admin |

---

**Happy Testing! 🚀**
