# Local Development Setup

Choose your setup based on your needs:

## Option 1: Without S3 (Recommended if MinIO Fails) ⭐

**Best for:** Testing admin endpoints, auth, and core functionality without file storage.

```bash
# 1. Start MongoDB only
docker-compose up -d mongodb

# 2. Run the app
S3_ENABLED=false go run cmd/api/main.go

# Or use the convenience script
./scripts/run_without_s3.sh
```

**What works:**
- ✅ Admin authentication & profile
- ✅ User registration & login
- ✅ All endpoints except file uploads/downloads

**What doesn't work:**
- ❌ File uploads
- ❌ File downloads

See [docs/RUNNING_WITHOUT_S3.md](docs/RUNNING_WITHOUT_S3.md) for details.

---

## Option 2: Full Stack (With S3/MinIO)

**Best for:** Testing file uploads and full functionality.

```bash
# 1. Start all services
docker-compose up -d

# 2. Run the app (S3_ENABLED defaults to true)
go run cmd/api/main.go
```

**What works:**
- ✅ All features
- ✅ File uploads
- ✅ File downloads
- ✅ Full admin endpoints

**If MinIO fails:**
- Check logs: `docker-compose logs minio`
- Try resetting: `docker-compose down && docker-compose up -d minio`
- Or use Option 1 (without S3)

---

## Testing Admin Endpoints

See [TESTING_SETUP_README.md](TESTING_SETUP_README.md) for complete testing setup.

**Quick test (after starting the app):**
```bash
# Health check
curl http://localhost:8080/health

# Admin login (default credentials in docs/ADMIN_API.md)
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'
```

---

## Environment Variables

### Required (both options)
- `JWT_SECRET` - At least 32 characters

### S3-Specific (Option 2 only)
- `S3_ENABLED=true` (default)
- `S3_ENDPOINT_URL=http://localhost:9000`
- `S3_BUCKET=temp-bucket`

### S3-Optional (Option 1 only)
- `S3_ENABLED=false`
- Other S3 vars can be empty

See `.env.example` for all variables.

---

## Common Issues

| Issue | Solution |
|-------|----------|
| MinIO won't start | Use Option 1 (without S3) |
| "file storage is disabled" | You're using S3_ENABLED=false (expected) |
| MongoDB auth failed | Check MONGO_URI includes `?authSource=admin` |
| Port 8080 in use | Change `PORT` env var |

---

## Documentation

- **Quick Start:** This file
- **Admin Testing:** [TESTING_SETUP_README.md](TESTING_SETUP_README.md)
- **Running Without S3:** [docs/RUNNING_WITHOUT_S3.md](docs/RUNNING_WITHOUT_S3.md)
- **S3 Feature Details:** [S3_OPTIONAL.md](S3_OPTIONAL.md)
- **API Documentation:** [docs/ADMIN_API.md](docs/ADMIN_API.md)
- **Changes Made:** [CHANGELOG_S3_OPTIONAL.md](CHANGELOG_S3_OPTIONAL.md)

---

## Database

MongoDB is always required.

**Connection:** `mongodb://admin:secret@localhost:27017/temp_backend?authSource=admin`

**Start MongoDB:**
```bash
docker-compose up -d mongodb
```

**Connect to MongoDB:**
```bash
mongosh "mongodb://admin:secret@localhost:27017/temp_backend?authSource=admin"
```

---

## Need Help?

1. Check the relevant documentation file above
2. See troubleshooting sections in each guide
3. Verify environment variables in `.env`
4. Check service logs: `docker-compose logs -f [service]`
