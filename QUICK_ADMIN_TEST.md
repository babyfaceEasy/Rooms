# Quick Admin Testing Guide

## One-Liner Admin Test

```bash
# Login and get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}' | jq -r '.access_token')

# Get admin profile with token
curl http://localhost:8080/api/v1/admin/profile -H "Authorization: Bearer $TOKEN" | jq .
```

## Default Test Account

| Field | Value |
|-------|-------|
| Email | admin@example.com |
| Password | password123 |
| Role | super_admin |

## Common Endpoints

### Login (Public)
```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'
```

### Get Profile (Requires Token)
```bash
curl http://localhost:8080/api/v1/admin/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Logout (Requires Token)
```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/admin/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"YOUR_REFRESH_TOKEN"}'
```

## Running Full Test Suite

```bash
bash scripts/test_admin_flow.sh
```

## Docker Commands

```bash
# View logs
docker-compose logs -f api

# Restart API
docker-compose restart api

# Check if API is running
docker-compose ps api

# Access MongoDB directly
docker-compose exec -T mongodb mongosh mongodb://admin:secret@localhost:27017/temp_backend?authSource=admin

# Query admin users
docker-compose exec -T mongodb mongosh mongodb://admin:secret@localhost:27017/temp_backend?authSource=admin --eval "db.admins.find().pretty()"
```

## Troubleshooting

### "missing authorization header"
- The endpoint requires a token
- Add: `-H "Authorization: Bearer <token>"`

### "Admin authentication is required" (401)
- Password is incorrect or admin doesn't exist
- Check MongoDB for admin user
- Verify password hash is correct

### "invalid token"
- Token has expired
- Use refresh token to get new access token
- Or login again

## Creating New Admin User

1. Generate password hash:
```go
package main
import ("fmt"; "golang.org/x/crypto/bcrypt")
func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte("your_password"), bcrypt.DefaultCost)
    fmt.Println(string(hash))
}
```

2. Insert into MongoDB:
```javascript
db.admins.insertOne({
  email: "newadmin@example.com",
  password_hash: "YOUR_HASH_HERE",
  name: "New Admin",
  role: "admin",
  permissions: ["view_analytics"],
  is_active: true,
  failed_login_attempts: 0,
  created_at: new Date(),
  updated_at: new Date(),
  locked_until: null
})
```

## Documentation

- Full testing guide: `docs/ADMIN_LOGIN_TESTING.md`
- API documentation: `docs/ADMIN_API.md`
- Architecture: `docs/ADMIN_ARCHITECTURE.md`
