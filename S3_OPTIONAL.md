# S3 is Now Optional! 🎉

Your application can now run **without S3/MinIO**. This solves the problem of MinIO container failing.

## Quick Solution

If S3/MinIO keeps failing, just disable it:

```bash
# Set environment variable
export S3_ENABLED=false

# Start MongoDB only
docker-compose up -d mongodb

# Run the app
go run cmd/api/main.go
```

**Expected Output:**
```
s3 storage disabled - using no-op storage
server started on port 8080
```

## What Works Without S3

✅ All admin endpoints (login, profile, etc.)
✅ All user endpoints (register, profile, etc.)
✅ Room management
✅ Post creation (text-only)
✅ Item metadata management
✅ Everything except file uploads/downloads

## What Doesn't Work Without S3

❌ File uploads (items, posts, profile avatars)
❌ File downloads
❌ Image/video storage

## How It Works

1. **S3_ENABLED=false** disables S3 initialization
2. Application uses `NoOpStorage` instead of S3
3. File operations return: `"file storage is disabled"` error
4. Everything else works normally

## Files Changed

### New Files
- `internal/repository/noop_storage_repository.go` — No-op storage implementation
- `docs/RUNNING_WITHOUT_S3.md` — Complete guide
- `scripts/run_without_s3.sh` — Convenience script

### Modified Files
- `config/config.go` — Added `S3.Enabled` field, optional S3 validation
- `cmd/api/main.go` — Conditional S3 initialization
- `.env.example` — Added `S3_ENABLED=true` flag and docs

## Environment Variable

| Variable | Default | Purpose |
|----------|---------|---------|
| `S3_ENABLED` | `true` | Enable/disable S3 storage |

When `S3_ENABLED=false`:
- `S3_ENDPOINT_URL`, `S3_BUCKET` are not required
- All S3 config can be left empty
- Application starts with no-op storage

## Testing

```bash
# Run without S3
S3_ENABLED=false go run cmd/api/main.go

# Test health
curl http://localhost:8080/health

# File upload will fail gracefully
curl -F "file=@test.txt" http://localhost:8080/api/v1/items
# Returns: {"error": "file storage is disabled"}
```

## Switching Back to S3

When you're ready to use S3 again:

```bash
export S3_ENABLED=true
export S3_ENDPOINT_URL=http://localhost:9000
export S3_BUCKET=temp-bucket

docker-compose up -d minio

go run cmd/api/main.go
```

## No Code Changes Required

The application automatically detects `S3_ENABLED` and:
- Initializes S3 only if enabled
- Uses no-op storage if disabled
- Works seamlessly either way

See [RUNNING_WITHOUT_S3.md](docs/RUNNING_WITHOUT_S3.md) for full details.
