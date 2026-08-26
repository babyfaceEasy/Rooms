# Posts Module API Documentation

This document provides comprehensive API documentation for the Posts module, including all endpoints, request/response formats, and examples.

## Base URL

```
http://localhost:3000/api/v1
```

## Authentication

All endpoints in this module require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <access_token>
```

The JWT token is obtained from the login endpoint (`POST /api/v1/auth/login`).

---

## Overview

The Posts module allows users to create, retrieve, and delete posts. Posts support optional image, video, and audio attachments that are stored in S3-compatible object storage (MinIO for development, AWS S3 for production).

### Supported File Types

**Images:** jpg, jpeg, png, gif, webp

**Videos:** mp4, webm, mov, avi

**Audio:** mp3, wav, m4a, aac, flac, ogg

---

## Endpoints

### 1. Create Post

Creates a new post with optional image, video, and audio attachments.

**Endpoint:** `POST /api/v1/posts`

**Authentication:** Required

**Request Format:** `multipart/form-data`

**Form Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| text | string | Yes | Post text content (1-5000 characters) |
| image | file | No | Image file (max size limited by S3 bucket config) |
| video | file | No | Video file (max size limited by S3 bucket config) |
| audio | file | No | Audio file (max size limited by S3 bucket config) |

**Example Request (using curl):**

```bash
# Create post with text only
curl -X POST http://localhost:3000/api/v1/posts \
  -H "Authorization: Bearer <access_token>" \
  -F "text=This is my first post!"

# Create post with image
curl -X POST http://localhost:3000/api/v1/posts \
  -H "Authorization: Bearer <access_token>" \
  -F "text=Check out this photo!" \
  -F "image=@/path/to/image.jpg"

# Create post with video
curl -X POST http://localhost:3000/api/v1/posts \
  -H "Authorization: Bearer <access_token>" \
  -F "text=Watch this video!" \
  -F "video=@/path/to/video.mp4"

# Create post with audio
curl -X POST http://localhost:3000/api/v1/posts \
  -H "Authorization: Bearer <access_token>" \
  -F "text=Listen to this podcast episode!" \
  -F "audio=@/path/to/episode.mp3"

# Create post with image, video, and audio
curl -X POST http://localhost:3000/api/v1/posts \
  -H "Authorization: Bearer <access_token>" \
  -F "text=Full media package" \
  -F "image=@/path/to/image.png" \
  -F "video=@/path/to/video.webm" \
  -F "audio=@/path/to/audio.wav"
```

**Success Response (201 Created):**

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439013",
    "user_id": "507f1f77bcf86cd799439012",
    "text": "This is my first post!",
    "image": null,
    "video": null,
    "created_at": "2024-06-28T21:30:00Z",
    "updated_at": "2024-06-28T21:30:00Z"
  },
  "message": "post created successfully",
  "status": 201
}
```

**Success Response with Image (201 Created):**

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439014",
    "user_id": "507f1f77bcf86cd799439012",
    "text": "Check out this photo!",
    "image": "posts/images/507f1f77bcf86cd799439012/sunset.jpg",
    "video": null,
    "audio": null,
    "created_at": "2024-06-28T21:32:00Z",
    "updated_at": "2024-06-28T21:32:00Z"
  },
  "message": "post created successfully",
  "status": 201
}
```

**Success Response with Audio (201 Created):**

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439015",
    "user_id": "507f1f77bcf86cd799439012",
    "text": "Listen to this podcast episode!",
    "image": null,
    "video": null,
    "audio": "posts/audio/507f1f77bcf86cd799439012/episode.mp3",
    "created_at": "2024-06-28T21:35:00Z",
    "updated_at": "2024-06-28T21:35:00Z"
  },
  "message": "post created successfully",
  "status": 201
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid input | Missing text field or invalid request format |
| 400 | invalid image type | Image file has unsupported extension |
| 400 | invalid video type | Video file has unsupported extension |
| 400 | invalid audio type | Audio file has unsupported extension |
| 400 | text is required | Text field is empty |
| 400 | text must be at least 1 character | Text too short |
| 400 | text must not exceed 5000 characters | Text too long |
| 401 | unauthorized | Missing or invalid JWT token |
| 500 | failed to upload image | S3 upload error for image |
| 500 | failed to upload video | S3 upload error for video |
| 500 | failed to upload audio | S3 upload error for audio |

**Example Error Response (400 - Invalid Image Type):**

```json
{
  "error": "invalid image type",
  "status": 400
}
```

**Example Error Response (400 - Text Validation):**

```json
{
  "error": "text must be at least 1 character",
  "status": 400
}
```

---

### 2. Get Post

Retrieves a post by ID.

**Endpoint:** `GET /api/v1/posts/:id`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Post ID (MongoDB ObjectID) |
| sort | string | No | Sort order: `desc` (newest first, default) or `asc` (oldest first) |

**Example Request:**

```bash
curl -X GET http://localhost:3000/api/v1/posts/507f1f77bcf86cd799439013 \
  -H "Authorization: Bearer <access_token>"
```

**Success Response (200 OK):**

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439013",
    "room_id": "507f1f77bcf86cd799439011",
    "room_code": "MY_ROOM",
    "room_name": "My Room",
    "user_id": "507f1f77bcf86cd799439012",
    "user_name": "John Doe",
    "text": "This is my first post!",
    "image": null,
    "video": null,
    "audio": null,
    "validations_count": 3,
    "respects_count": 7,
    "created_at": "2024-06-28T21:30:00Z",
    "updated_at": "2024-06-28T21:30:00Z"
  },
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid post id | Invalid MongoDB ObjectID format |
| 401 | unauthorized | Missing or invalid JWT token |
| 404 | post not found | Post does not exist or is soft-deleted |
| 500 | internal server error | Server error |

**Example Error Response (404):**

```json
{
  "error": "post not found",
  "status": 404
}
```

---

### 3. Delete Post

Deletes a post (soft delete). Only the post creator can delete their posts.

**Endpoint:** `DELETE /api/v1/posts/:id`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Post ID (MongoDB ObjectID) |
| sort | string | No | Sort order: `desc` (newest first, default) or `asc` (oldest first) |

**Example Request:**

```bash
curl -X DELETE http://localhost:3000/api/v1/posts/507f1f77bcf86cd799439013 \
  -H "Authorization: Bearer <access_token>"
```

**Success Response (200 OK):**

```json
{
  "message": "post deleted successfully",
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid post id | Invalid MongoDB ObjectID format |
| 401 | unauthorized | Missing or invalid JWT token |
| 403 | forbidden | User is not the post creator |
| 404 | post not found | Post does not exist |
| 500 | internal server error | Server error |

**Example Error Response (403 - Unauthorized Delete):**

```json
{
  "error": "forbidden",
  "status": 403
}
```

---

### 4. Report Post

Reports a post for inappropriate content. Users can report posts for various reasons including bullying, harmful content, spam, etc. Each user can report a post at most once per day per post.

**Endpoint:** `POST /api/v1/posts/:id/report`

**Authentication:** Required

**Content-Type:** `application/json`

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Post ID (MongoDB ObjectID) |

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| reason | string | Yes | Report reason (must be one of valid reasons below) |
| comment | string | No | Optional additional context (max 500 characters) |

**Valid Report Reasons:**

- `under_18` — Post involves minors inappropriately
- `bullying_harassment` — Post contains bullying or harassment
- `suicide_self_harm` — Post discusses suicide or self-harm
- `violent_hateful` — Post contains violence or hate speech
- `selling_restricted` — Post attempts to sell restricted items
- `adult_content` — Post contains explicit adult content
- `scam_fraud` — Post is spam, scam, or fraud
- `intellectual_property` — Post violates intellectual property rights
- `dont_want_to_see` — Generic dislike/don't want to see

**Example Request:**

```bash
# Report post for bullying
curl -X POST http://localhost:3000/api/v1/posts/507f1f77bcf86cd799439013/report \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "bullying_harassment",
    "comment": "This post contains personal attacks"
  }'

# Report post without comment
curl -X POST http://localhost:3000/api/v1/posts/507f1f77bcf86cd799439013/report \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "dont_want_to_see"
  }'
```

**Success Response (200 OK):**

```json
{
  "message": "post reported successfully",
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid input | Missing reason field or invalid request format |
| 400 | invalid report reason | Reason is not a valid report type |
| 400 | cannot report own post | User cannot report their own posts |
| 401 | unauthorized | Missing or invalid JWT token |
| 404 | post not found | Post does not exist or is soft-deleted |
| 409 | report already exists | User has already reported this post |
| 429 | report limit exceeded | User has exceeded their daily report limit (default: 10 per day) |
| 500 | internal server error | Server error |

**Example Error Response (409 - Already Reported):**

```json
{
  "error": "report already exists",
  "status": 409
}
```

**Example Error Response (429 - Rate Limited):**

```json
{
  "error": "report limit exceeded",
  "status": 429
}
```

**Auto-Moderation:**

Posts that receive a configurable number of reports (default: 15) are automatically soft-deleted from the platform. The post creator is notified when their post receives a report via the notifications system.

**Note on Privacy:**

- Reporter identity is kept confidential and not shared with the post creator
- Post creator is notified that their post was reported along with the reason, but not who reported it
- Reports are stored permanently for moderation team review

---

## Data Model

### Post

```json
{
  "id": "string (MongoDB ObjectID)",
  "user_id": "string (MongoDB ObjectID of creator)",
  "text": "string (1-5000 characters)",
  "image": "string or null (S3 object key)",
  "video": "string or null (S3 object key)",
  "audio": "string or null (S3 object key)",
  "created_at": "string (ISO 8601 timestamp)",
  "updated_at": "string (ISO 8601 timestamp)",
  "deleted_at": "string (ISO 8601 timestamp) or null"
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique post identifier |
| user_id | string | Creator's user ID |
| text | string | Post content |
| image | string\|null | S3 path to image (nullable) |
| video | string\|null | S3 path to video (nullable) |
| audio | string\|null | S3 path to audio file (nullable) |
| created_at | string | Creation timestamp |
| updated_at | string | Last update timestamp |
| deleted_at | string\|null | Soft delete timestamp (null if active) |

---

## Common Use Cases

### 1. Create a Simple Text Post

```bash
curl -X POST http://localhost:3000/api/v1/posts \
  -H "Authorization: Bearer eyJhbGc..." \
  -F "text=Hello world!"
```

### 2. Create a Post with Image

```bash
curl -X POST http://localhost:3000/api/v1/posts \
  -H "Authorization: Bearer eyJhbGc..." \
  -F "text=Beautiful sunset" \
  -F "image=@sunset.jpg"
```

### 3. Create a Post with Video

```bash
curl -X POST http://localhost:3000/api/v1/posts \
  -H "Authorization: Bearer eyJhbGc..." \
  -F "text=Check this out" \
  -F "video=@video.mp4"
```

### 4. Create a Post with Audio

```bash
curl -X POST http://localhost:3000/api/v1/posts \
  -H "Authorization: Bearer eyJhbGc..." \
  -F "text=Listen to this podcast episode" \
  -F "audio=@episode.mp3"
```

### 5. Create a Post with Multiple Media Types

```bash
curl -X POST http://localhost:3000/api/v1/posts \
  -H "Authorization: Bearer eyJhbGc..." \
  -F "text=Full multimedia post" \
  -F "video=@tutorial.mp4"
```

### 4. View Your Own Post

```bash
curl -X GET http://localhost:3000/api/v1/posts/507f1f77bcf86cd799439013 \
  -H "Authorization: Bearer eyJhbGc..."
```

### 5. Delete a Post

```bash
curl -X DELETE http://localhost:3000/api/v1/posts/507f1f77bcf86cd799439013 \
  -H "Authorization: Bearer eyJhbGc..."
```

---

## S3 File Storage

Files are stored in S3 with the following path structure:

```
posts/images/{user_id}/{filename}
posts/videos/{user_id}/{filename}
posts/audio/{user_id}/{filename}
```

**Example Paths:**
- `posts/images/507f1f77bcf86cd799439012/vacation-photo.jpg`
- `posts/videos/507f1f77bcf86cd799439012/birthday-video.mp4`
- `posts/audio/507f1f77bcf86cd799439012/podcast-episode.mp3`

### Local Development (MinIO)

For local development, MinIO is used as an S3-compatible object store. Files can be accessed via the MinIO console at:

```
http://localhost:9001
```

### Production (AWS S3)

In production, files are stored in AWS S3 and accessible via CloudFront or direct S3 URLs.

---

## Notes

- **Soft Delete:** Posts are never permanently deleted from the database. Instead, they are marked with a `deleted_at` timestamp and excluded from queries.
- **File Ownership:** Image and video files are namespaced by user ID to ensure isolation and prevent accidental overwrites.
- **Single Upload:** The API accepts one image and one video per request. If multiple files of the same type are provided, only the first one is used.
- **File Validation:** File type is validated by file extension before upload. Ensure files have correct extensions.
- **File Size:** Maximum file size is determined by S3 bucket configuration. Consult deployment documentation for limits.

---

## Integration with Other Modules

### User Authentication

All post endpoints require valid JWT tokens from the authentication module. Tokens are issued via:

- `POST /api/v1/auth/register` — Register new user
- `POST /api/v1/auth/login` — Login and get access token
- `POST /api/v1/auth/refresh` — Refresh access token

### Room Integration (Future)

Posts can be extended to be associated with rooms. Future endpoints may include:

- `POST /api/v1/rooms/:code/posts` — Post to a room
- `GET /api/v1/rooms/:code/posts` — List room posts

---

## Testing

For testing the Posts API, see [TESTING.md](TESTING.md) for instructions on setting up the test environment and running tests.

### Unit Tests

- Domain validation tests: `internal/domain/post_test.go`
- Service layer tests: `internal/service/post_service_test.go`
- Handler layer tests: `internal/handler/post_handler_test.go`

### Running Tests

```bash
# Run all tests
go test ./...

# Run only post tests
go test ./internal/domain ./internal/service ./internal/handler -v -k Post

# Run tests with coverage
go test ./internal/service -cover
```

---

## Comments Module

The Comments module allows users to add text comments to posts. Comments are stored as separate documents in MongoDB and linked to posts via `post_id`.

### Comment Limits

- **Text Length:** 1-1000 characters
- **Delete Permission:** Only the comment author can delete their own comment

### Comment Endpoints

### 1. Create Comment

Creates a new comment on a post.

**Endpoint:** `POST /api/v1/posts/:id/comments`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Post ID (MongoDB ObjectID) |
| sort | string | No | Sort order: `desc` (newest first, default) or `asc` (oldest first) |

**Request Body:**

```json
{
  "post_id": "507f1f77bcf86cd799439013",
  "text": "Great post! I really enjoyed this."
}
```

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| post_id | string | Yes | Post ID (must match URL parameter) |
| text | string | Yes | Comment text (1-1000 characters) |

**Success Response (201 Created):**

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439020",
    "post_id": "507f1f77bcf86cd799439013",
    "user_id": "507f1f77bcf86cd799439012",
    "text": "Great post! I really enjoyed this.",
    "created_at": "2024-06-28T22:15:00Z",
    "updated_at": "2024-06-28T22:15:00Z"
  },
  "message": "comment created successfully",
  "status": 201
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid request body | Missing or invalid JSON |
| 400 | post_id is required | Missing post_id field |
| 400 | comment text is required | Text field is empty |
| 400 | comment text exceeds maximum length of 1000 characters | Text too long |
| 400 | invalid post id | Invalid ObjectID format |
| 401 | unauthorized | Missing or invalid JWT token |
| 404 | post not found | Post doesn't exist |
| 500 | internal server error | Server error |

**Example Error Response (404 - Post Not Found):**

```json
{
  "error": "post not found",
  "status": 404
}
```

**Example Error Response (400 - Text Required):**

```json
{
  "error": "comment text is required",
  "status": 400
}
```

**Example Request (curl):**

```bash
curl -X POST http://localhost:3000/api/v1/posts/507f1f77bcf86cd799439013/comments \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": "507f1f77bcf86cd799439013",
    "text": "Great post! I really enjoyed this."
  }'
```

---

### 2. Get Comments for Post

Retrieves all comments for a specific post.

**Endpoint:** `GET /api/v1/posts/:id/comments`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Post ID (MongoDB ObjectID) |
| sort | string | No | Sort order: `desc` (newest first, default) or `asc` (oldest first) |

**Success Response (200 OK):**

```json
{
  "data": [
    {
      "id": "507f1f77bcf86cd799439020",
      "post_id": "507f1f77bcf86cd799439013",
      "user_id": "507f1f77bcf86cd799439012",
      "text": "Great post! I really enjoyed this.",
      "created_at": "2024-06-28T22:15:00Z",
      "updated_at": "2024-06-28T22:15:00Z"
    },
    {
      "id": "507f1f77bcf86cd799439021",
      "post_id": "507f1f77bcf86cd799439013",
      "user_id": "507f1f77bcf86cd799439014",
      "text": "Thanks for sharing this!",
      "created_at": "2024-06-28T22:20:00Z",
      "updated_at": "2024-06-28T22:20:00Z"
    }
  ],
  "count": 2,
  "message": "comments retrieved successfully",
  "status": 200
}
```

**Empty Response (200 OK):**

```json
{
  "data": [],
  "count": 0,
  "message": "comments retrieved successfully",
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | post id is required | Missing URL parameter |
| 400 | invalid post id | Invalid ObjectID format |
| 401 | unauthorized | Missing or invalid JWT token |
| 500 | internal server error | Server error |

**Example Request (curl):**

```bash
curl http://localhost:3000/api/v1/posts/507f1f77bcf86cd799439013/comments \
  -H "Authorization: Bearer <access_token>"
```

---

### 3. Delete Comment

Deletes a comment. Only the comment author can delete their own comment.

**Endpoint:** `DELETE /api/v1/comments/:id`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Comment ID (MongoDB ObjectID) |

**Success Response (200 OK):**

```json
{
  "data": null,
  "message": "comment deleted successfully",
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | comment id is required | Missing URL parameter |
| 400 | invalid comment id | Invalid ObjectID format |
| 401 | unauthorized | Missing or invalid JWT token |
| 403 | forbidden | User is not the comment author |
| 404 | comment not found | Comment doesn't exist |
| 500 | internal server error | Server error |

**Example Error Response (403 - Not Author):**

```json
{
  "error": "you are not authorized to perform this action on this comment",
  "status": 403
}
```

**Example Request (curl):**

```bash
curl -X DELETE http://localhost:3000/api/v1/comments/507f1f77bcf86cd799439020 \
  -H "Authorization: Bearer <access_token>"
```

---

## HTTP Status Codes

| Code | Meaning | Usage |
|------|---------|-------|
| 200 | OK | Successful GET, DELETE |
| 201 | Created | Successful POST (create comment) |
| 400 | Bad Request | Invalid input or validation error |
| 401 | Unauthorized | Missing or invalid JWT token |
| 403 | Forbidden | User lacks permission (not comment author) |
| 404 | Not Found | Resource (post, comment) doesn't exist |
| 500 | Internal Server Error | Unexpected server error |
