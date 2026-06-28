# Posts Module - Quick Reference

## Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/posts` | Create post (text, optional image/video) | Required |
| GET | `/api/v1/posts/:id` | Get post by ID | Required |
| DELETE | `/api/v1/posts/:id` | Delete post (soft delete, owner only) | Required |

## Quick Examples

### Create Text Post
```bash
curl -X POST http://localhost:3000/api/v1/posts \
  -H "Authorization: Bearer TOKEN" \
  -F "text=Hello world!"
```

### Create Post with Image
```bash
curl -X POST http://localhost:3000/api/v1/posts \
  -H "Authorization: Bearer TOKEN" \
  -F "text=My photo" \
  -F "image=@photo.jpg"
```

### Create Post with Video
```bash
curl -X POST http://localhost:3000/api/v1/posts \
  -H "Authorization: Bearer TOKEN" \
  -F "text=My video" \
  -F "video=@video.mp4"
```

### Get Post
```bash
curl -X GET http://localhost:3000/api/v1/posts/507f1f77bcf86cd799439013 \
  -H "Authorization: Bearer TOKEN"
```

### Delete Post
```bash
curl -X DELETE http://localhost:3000/api/v1/posts/507f1f77bcf86cd799439013 \
  -H "Authorization: Bearer TOKEN"
```

## Validation Rules

| Field | Rule |
|-------|------|
| text | Required, 1-5000 characters |
| image | Optional, jpg/jpeg/png/gif/webp |
| video | Optional, mp4/webm/mov/avi |

## Response Codes

| Code | Meaning |
|------|---------|
| 201 | Post created successfully |
| 200 | Success (get/delete) |
| 400 | Bad request (validation error) |
| 401 | Unauthorized (invalid token) |
| 403 | Forbidden (not post owner) |
| 404 | Post not found |
| 500 | Server error |

## S3 Storage Paths

- Images: `posts/images/{user_id}/{filename}`
- Videos: `posts/videos/{user_id}/{filename}`

## Notes

- Posts use **soft delete** (marked as deleted, not removed)
- Only post creator can delete posts
- One image + one video per post (both optional)
- Files validated by extension before upload
- JWT token required for all endpoints
