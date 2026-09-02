# Posts Module - Quick Reference

## Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/posts` | Create post (text, optional image/video/audio) | Required |
| GET | `/api/v1/posts/:id` | Get post by ID | Required |
| DELETE | `/api/v1/posts/:id` | Delete post (soft delete, owner only) | Required |
| POST | `/api/v1/posts/:id/validate` | Validate a post | Required |
| DELETE | `/api/v1/posts/:id/validate` | Remove validation from a post | Required |
| POST | `/api/v1/posts/:id/respect` | Respect a post | Required |
| DELETE | `/api/v1/posts/:id/respect` | Remove respect from a post | Required |
| POST | `/api/v1/posts/:id/comments` | Create comment on post | Required |
| GET | `/api/v1/posts/:id/comments` | Get paginated comments for post (?page=&limit=&sort=) | Required |
| DELETE | `/api/v1/comments/:id` | Delete comment (author only) | Required |
| POST | `/api/v1/posts/:id/report` | Report post for inappropriate content | Required |
| GET | `/api/v1/posts/stream/new` | Stream new posts (SSE) for a room (?room_code=) | Required |

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

### Report Post
```bash
curl -X POST http://localhost:3000/api/v1/posts/507f1f77bcf86cd799439013/report \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "bullying_harassment",
    "comment": "This post contains personal attacks"
  }'
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
| 200 | Success (get/delete/validate/respect) |
| 201 | Post created successfully |
| 400 | Bad request (validation error, NOT_VALIDATED, NOT_RESPECTED, invalid reason) |
| 401 | Unauthorized (invalid token) |
| 403 | Forbidden (not post owner, not room member) |
| 404 | Post not found |
| 409 | Conflict (ALREADY_VALIDATED, ALREADY_RESPECTED, report already exists) |
| 429 | Too Many Requests (daily report limit exceeded) |
| 500 | Server error |

## S3 Storage Paths

- Images: `posts/images/{user_id}/{filename}`
- Videos: `posts/videos/{user_id}/{filename}`

## Response Changes

### Post Response
```json
{
  "id": "string",
  "room_id": "string",
  "room_code": "string",
  "room_name": "string",
  "user_id": "string",
  "user_name": "string",
  "text": "string",
  "image": "string or null (full S3 URL)",
  "video": "string or null (full S3 URL)",
  "audio": "string or null (full S3 URL)",
  "created_at": "string",
  "updated_at": "string"
}
```

### Comment Response
```json
{
  "id": "string",
  "post_id": "string",
  "user_id": "string",
  "user_name": "string",
  "text": "string",
  "created_at": "string",
  "updated_at": "string"
}
```

### Paginated List Response
Comments and room posts return paginated results:
```json
{
  "data": [...],
  "count": 20,
  "page": 1,
  "limit": 20,
  "total": 153,
  "message": "...",
  "status": 200
}
```

Query parameters: `?page=1&limit=20` (default page=1, limit=20, max limit=100)

## Notes

- Posts use **soft delete** (marked as deleted, not removed)
- Only post creator can delete posts
- One image + one video + one audio per post (all optional)
- Files validated by extension before upload
- JWT token required for all endpoints
- Room membership required for all post operations
- **Validate**: Room members can mark a post as valid (POST). Duplicate returns 409. Remove with DELETE.
- **Respect**: Room members can respect a post (POST). Duplicate returns 409. Remove with DELETE.
- `validations_count` and `respects_count` are returned in every post response
- Comments are text-only (1-1000 characters)
- Only comment author can delete their own comment
- Comments linked to posts via `post_id`
