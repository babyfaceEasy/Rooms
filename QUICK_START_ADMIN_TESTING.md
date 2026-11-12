# Quick Start - Admin Endpoint Testing

## ⚡ Quick Setup (5 minutes)

### 1. Start Docker Services
```bash
cd /Users/killacam/Documents/work/bouk/temp_backend
docker-compose up -d mongodb minio createbuckets
```

### 2. Copy Environment File
```bash
cp .env.example .env
```

### 3. Run the Backend
```bash
go run cmd/api/main.go
```

You should see:
```
2024/11/12 12:00:00 Starting server on :8080
```

### 4. Create Admin User (in another terminal)

```bash
# Connect to MongoDB
mongosh "mongodb://admin:secret@localhost:27017/temp_backend?authSource=admin"
```

In the MongoDB shell:

```javascript
use temp_backend

db.admins.insertOne({
  "_id": ObjectId("507f1f77bcf86cd799439011"),
  "email": "admin@example.com",
  "password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86xVXAaxe9K",
  "name": "Test Admin",
  "role": "super_admin",
  "permissions": [
    "manage_admins", "manage_users", "suspend_users", "delete_users",
    "view_user_analytics", "manage_posts", "delete_posts",
    "view_post_analytics", "manage_patches", "delete_patches",
    "view_patch_analytics", "view_logs", "view_analytics"
  ],
  "is_active": true,
  "failed_login_attempts": 0,
  "created_at": new Date(),
  "updated_at": new Date()
})

exit
```

---

## 🧪 Quick Tests

### Test 1: Login
```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }'
```

Save the `access_token` from the response.

### Test 2: Get Admin Profile (NEW!)
```bash
curl -X GET http://localhost:8080/api/v1/admin/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN_HERE" \
  -H "Content-Type: application/json"
```

### Test 3: Run Full Test Suite
```bash
./scripts/test_admin_endpoints.sh
```

---

## 📋 Service URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| API | http://localhost:8080 | - |
| MongoDB | localhost:27017 | admin / secret |
| MinIO API | http://localhost:9000 | minioadmin / minioadmin |
| MinIO Console | http://localhost:9001 | minioadmin / minioadmin |

---

## 🛠️ Useful Commands

### Check Services Status
```bash
docker-compose ps
```

### View MongoDB Data
```bash
mongosh "mongodb://admin:secret@localhost:27017/temp_backend?authSource=admin"
# Then: db.admins.find().pretty()
```

### View API Logs
```bash
# If running with go run, check console output
# If using docker-compose:
docker-compose logs -f api
```

### Stop All Services
```bash
docker-compose down
```

### Clean Everything (⚠️ Deletes data)
```bash
docker-compose down -v
```

---

## 🐛 Common Issues & Solutions

### ❌ "Cannot connect to MongoDB"
```bash
# Check if MongoDB is running
docker-compose ps mongodb

# If not, restart it
docker-compose up -d mongodb
```

### ❌ "Invalid token" / Login Failed
- Ensure admin user exists in MongoDB
- Check password hash is correct
- Try with credentials: `admin@example.com` / `password123`

### ❌ "Connection refused on 8080"
- Check if API is running: `curl http://localhost:8080/health`
- Restart: `go run cmd/api/main.go`

### ❌ Port 8080 already in use
- Change PORT in `.env` to another port (e.g., 8081)
- Restart the API

---

## 📖 Full Documentation

For detailed information, see:
- `LOCAL_ADMIN_TESTING_GUIDE.md` - Complete testing guide
- `docs/ADMIN_API.md` - Full API documentation
- `README.md` - Project overview

---

## ✅ Expected Response

### Login Response
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer"
}
```

### Get Profile Response
```json
{
  "id": "507f1f77bcf86cd799439011",
  "email": "admin@example.com",
  "name": "Test Admin",
  "role": "super_admin",
  "permissions": [
    "manage_admins",
    "manage_users",
    ...
  ],
  "is_active": true,
  "last_login": null,
  "failed_login_attempts": 0,
  "created_at": "2024-11-12T12:00:00Z",
  "updated_at": "2024-11-12T12:00:00Z"
}
```

---

**Need help?** Check the full guide or the repository README! 🚀
