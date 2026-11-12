# Admin Endpoint Testing Setup - Complete Reference

## 🎯 Overview

You now have a complete, production-ready local testing setup for the admin endpoints. This includes:
- ✅ Implemented endpoint: `GET /api/v1/admin/profile`
- ✅ Comprehensive documentation (3 guides)
- ✅ Automated setup and testing scripts
- ✅ Docker services for local development
- ✅ **Option to run without S3/MinIO** if it keeps failing

## 📚 Documentation Files

### 0. **[NEW] RUNNING_WITHOUT_S3.md** (If S3 is Failing)
- **Time:** 2 minutes
- **Contains:** Simple commands to disable S3, what works and what doesn't
- **Best for:** Testing without file storage (bypasses S3/MinIO issues)
- **Quick Start:**
  ```bash
  S3_ENABLED=false docker-compose up -d mongodb
  go run cmd/api/main.go
  ```

### 1. **QUICK_START_ADMIN_TESTING.md** (Recommended First Read)
- **Time:** 5 minutes
- **Contains:** Copy-paste commands, quick reference, common issues
- **Best for:** Getting started quickly

### 2. **LOCAL_ADMIN_TESTING_GUIDE.md** (Complete Reference)
- **Time:** 15-20 minutes
- **Contains:** Step-by-step setup, Postman integration, troubleshooting, full test examples
- **Best for:** Understanding all details and advanced setup

## 🚀 Fastest Way to Test (5 Minutes)

### Open 3 Terminal Windows

**Terminal 1: Start Services**
```bash
cd /Users/killacam/Documents/work/bouk/temp_backend
docker-compose up -d mongodb minio createbuckets
```

**Terminal 2: Run Backend**
```bash
cd /Users/killacam/Documents/work/bouk/temp_backend
go run cmd/api/main.go
```

**Terminal 3: Create Admin & Test**
```bash
cd /Users/killacam/Documents/work/bouk/temp_backend

# Connect to MongoDB
mongosh "mongodb://admin:secret@localhost:27017/temp_backend?authSource=admin"

# In MongoDB shell, paste this:
use temp_backend
db.admins.insertOne({
  "_id": ObjectId("507f1f77bcf86cd799439011"),
  "email": "admin@example.com",
  "password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86xVXAaxe9K",
  "name": "Test Admin",
  "role": "super_admin",
  "permissions": ["manage_admins", "manage_users", "suspend_users", "delete_users", "view_user_analytics", "manage_posts", "delete_posts", "view_post_analytics", "manage_patches", "delete_patches", "view_patch_analytics", "view_logs", "view_analytics"],
  "is_active": true,
  "failed_login_attempts": 0,
  "created_at": new Date(),
  "updated_at": new Date()
})
exit

# Run automated tests
./scripts/test_admin_endpoints.sh
```

### Alternative: Run Without S3 (If MinIO Keeps Failing)

If the S3/MinIO container keeps crashing, run without it:

```bash
# Terminal 2: Run Backend without S3
S3_ENABLED=false go run cmd/api/main.go

# Or use the convenience script:
./scripts/run_without_s3.sh
```

See [RUNNING_WITHOUT_S3.md](docs/RUNNING_WITHOUT_S3.md) for details.

## 🧪 Manual Testing with curl

### Step 1: Login
```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'
```

### Step 2: Copy the access_token from response, then test the new endpoint:
```bash
curl -X GET http://localhost:8080/api/v1/admin/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN_HERE" \
  -H "Content-Type: application/json" | jq '.'
```

## 📋 What Gets Tested

| Test | Endpoint | Expected |
|------|----------|----------|
| Admin Login | POST /api/v1/admin/auth/login | 200 OK with tokens |
| Get Profile ✨ | GET /api/v1/admin/profile | 200 OK with admin data |
| Sensitive Fields | Check response | No password_hash or tokens |
| Invalid Token | GET with bad token | 401 Unauthorized |
| Missing Auth | GET without header | 401 Unauthorized |

## 📊 Service Information

| Service | URL | Credentials |
|---------|-----|-------------|
| **API** | http://localhost:8080 | - |
| **MongoDB** | localhost:27017 | admin / secret |
| **MinIO API** | http://localhost:9000 | minioadmin / minioadmin |
| **MinIO Console** | http://localhost:9001 | minioadmin / minioadmin |

## 🔑 Test Credentials

```
Email:    admin@example.com
Password: password123
Role:     super_admin
```

## 🛠️ Automated Scripts

### `scripts/start_dev.sh`
Automated setup that:
- Checks Docker installation
- Starts MongoDB and MinIO
- Builds the application
- Shows service information

Usage:
```bash
./scripts/start_dev.sh
```

### `scripts/test_admin_endpoints.sh`
Automated test suite that:
- Tests admin login
- Tests get admin profile (NEW!)
- Verifies sensitive fields removed
- Tests error cases
- Displays results

Usage:
```bash
./scripts/test_admin_endpoints.sh
```

## 📁 File Locations

Project root: `/Users/killacam/Documents/work/bouk/temp_backend/`

```
Documentation
├─ QUICK_START_ADMIN_TESTING.md
├─ LOCAL_ADMIN_TESTING_GUIDE.md
└─ TESTING_SETUP_README.md (this file)

Scripts
├─ scripts/start_dev.sh
└─ scripts/test_admin_endpoints.sh

Implementation
├─ internal/handler/admin_auth_handler.go (39 lines added)
├─ internal/api/routes.go (1 line added)
└─ docs/ADMIN_API.md (50 lines added)
```

## ✅ Expected Responses

### Login Success (200)
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer"
}
```

### Get Profile Success (200)
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
  "last_login": null,
  "failed_login_attempts": 0,
  "created_at": "2024-11-12T11:27:00Z",
  "updated_at": "2024-11-12T11:27:00Z"
}
```

### Error - Missing Authorization (401)
```json
{
  "error": "missing authorization header"
}
```

### Error - Invalid Token (401)
```json
{
  "error": "invalid token"
}
```

## 🔧 Common Commands

```bash
# Check Docker services status
docker-compose ps

# View logs
docker-compose logs -f mongodb
docker-compose logs -f api

# Stop services
docker-compose down

# Clean everything (DELETES DATA)
docker-compose down -v

# Build application
go build -o temp_backend cmd/api/main.go

# Run with debug logging
LOG_LEVEL=debug go run cmd/api/main.go
```

## 🆘 Troubleshooting Quick Links

| Issue | Solution |
|-------|----------|
| Services won't start | See LOCAL_ADMIN_TESTING_GUIDE.md → Troubleshooting |
| Login fails | Check admin user exists in MongoDB |
| Token expired | Get new token by logging in again |
| Port 8080 in use | Change PORT in .env |
| Can't connect to MongoDB | Check MongoDB container is running |

See **LOCAL_ADMIN_TESTING_GUIDE.md** for detailed troubleshooting.

## 📖 Next Steps

1. **Quick Start:** Read `QUICK_START_ADMIN_TESTING.md`
2. **Setup Services:** Run `docker-compose up -d mongodb minio createbuckets`
3. **Start Backend:** Run `go run cmd/api/main.go`
4. **Create Admin:** Connect to MongoDB and insert test admin
5. **Run Tests:** Execute `./scripts/test_admin_endpoints.sh`
6. **Verify:** See all tests pass ✅

## 🎯 Endpoint Details

### GET /api/v1/admin/profile

**Description:** Returns the profile of the currently logged-in admin

**Authentication:** Required (Bearer token)

**Returns:**
- Admin ID, email, name
- Role and permissions
- Account status and metadata
- Sensitive fields removed

**Errors:**
- `401 Unauthorized` - Missing or invalid token
- `500 Internal Server Error` - Failed to retrieve profile

---

**For complete details, see the Quick Start or Local Testing guides!**
