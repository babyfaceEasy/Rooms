# Running Without S3/MinIO

Your application can now run without the S3/MinIO container. This is useful for:
- Testing core functionality without file upload features
- Reducing local development dependencies
- Faster startup time

## Quick Start (No S3)

### 1. Start MongoDB Only

```bash
# Start only MongoDB (without MinIO)
docker-compose up -d mongodb

# Verify it's running
docker-compose ps mongodb
```

### 2. Configure Environment

Create or update your `.env` file:

```bash
# Disable S3/MinIO
S3_ENABLED=false

# Keep other settings as needed
MONGO_URI=mongodb://localhost:27017/temp_backend
JWT_SECRET=your-256-bit-secret-key-min-32-characters-long-here
```

### 3. Run the Application

```bash
go run cmd/api/main.go
```

You should see:
```
s3 storage disabled - using no-op storage
```

## How It Works

When `S3_ENABLED=false`:

- ✅ All endpoints work normally
- ✅ Admin authentication works
- ✅ Database operations work
- ❌ File uploads return: "file storage is disabled"
- ❌ File downloads return: "file storage is disabled"

## Testing Without S3

Most endpoints work fine without S3:

```bash
# Health check
curl http://localhost:8080/health

# Admin login
curl -X POST http://localhost:8080/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'

# Get admin profile (with token)
curl http://localhost:8080/api/v1/admin/profile \
  -H "Authorization: Bearer <token>"
```

## Endpoints Affected by S3 Disabled

The following endpoints will fail if you try to upload/download files:

- `POST /api/v1/items` (with file upload)
- `GET /api/v1/items/:id/download`
- `PUT /api/v1/items/:id` (with file update)
- `POST /api/v1/posts` (with media)
- `PATCH /api/v1/profile` (avatar upload)

They will return: `"file storage is disabled"` error.

## Switching Back to S3

To re-enable S3/MinIO:

```bash
# Update .env
S3_ENABLED=true
S3_ENDPOINT_URL=http://localhost:9000
S3_BUCKET=temp-bucket
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin

# Start MinIO container
docker-compose up -d minio

# Run the application
go run cmd/api/main.go
```

## Troubleshooting

### Application won't start without S3

Check that `S3_ENABLED=false` is set in your `.env` file. The application won't exit if S3 is disabled; only file operations will fail.

### MongoDB connection fails

If you see "mongo connection failed", ensure:
- MongoDB is running: `docker-compose ps mongodb`
- `MONGO_URI` is correct in `.env`
- Default: `mongodb://localhost:27017/temp_backend`

### Want to use S3 but MinIO won't start?

Check MinIO logs:
```bash
docker-compose logs minio
```

Common issues:
- Port 9000 already in use
- Previous MinIO data conflicts
- Solution: `docker-compose down && docker-compose up -d minio`

## Environment Variables Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `S3_ENABLED` | `true` | Enable/disable S3 storage |
| `S3_ENDPOINT_URL` | `http://localhost:9000` | S3 endpoint (ignored if S3_ENABLED=false) |
| `S3_BUCKET` | `temp-bucket` | Bucket name (ignored if S3_ENABLED=false) |
| `S3_REGION` | `us-east-1` | S3 region |
| `S3_ACCESS_KEY` | `minioadmin` | S3 access key |
| `S3_SECRET_KEY` | `minioadmin` | S3 secret key |

When `S3_ENABLED=false`, the S3 configuration variables are not validated and can be left empty.
