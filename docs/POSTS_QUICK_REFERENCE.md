# Posts Module - Quick Reference

## Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/posts` | Create post (text, optional image/video) | Required |
| GET | `/api/v1/posts/:id` | Get post by ID | Required |
| DELETE | `/api/v1/posts/:id` | Delete post (soft delete, owner only) | Required |
| POST | `/api/v1/posts/:id/comments` | Create comment on post | Required |
| GET | `/api/v1/posts/:id/comments` | Get all comments for post | Required |
| DELETE | `/api/v1/comments/:id` | Delete comment (author only) | Required |

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

### Create Comment
```bash
curl -X POST http://localhost:3000/api/v1/posts/507f1f77bcf86cd799439013/comments \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"post_id":"507f1f77bcf86cd799439013","text":"Great post!"}'
```

### Get Comments for Post
```bash
curl -X GET http://localhost:3000/api/v1/posts/507f1f77bcf86cd799439013/comments \
  -H "Authorization: Bearer TOKEN"
```

### Delete Comment
```bash
curl -X DELETE http://localhost:3000/api/v1/comments/507f1f77bcf86cd799439020 \
  -H "Authorization: Bearer TOKEN"
```

## Validation Rules

| Field | Rule |
|-------|------|
| text (post) | Required, 1-5000 characters |
| image | Optional, jpg/jpeg/png/gif/webp |
| video | Optional, mp4/webm/mov/avi |
| text (comment) | Required, 1-1000 characters |

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
- Comments are text-only (1-1000 characters)
- Only comment author can delete their own comment
- Comments linked to posts via `post_id`
