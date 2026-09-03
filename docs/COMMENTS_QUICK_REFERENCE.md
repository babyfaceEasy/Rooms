# Comments Module - Quick Reference

## Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/posts/:id/comments` | Create comment on post | Required |
| GET | `/api/v1/posts/:id/comments` | Get paginated comments (?page=&limit=&sort=) | Required |
| GET | `/api/v1/posts/:id/stream/comments` | Stream new comments (SSE) for a post | Required |
| DELETE | `/api/v1/comments/:id` | Delete comment (author only) | Required |

## Quick Examples

### Create Comment
```bash
curl -X POST http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439013/comments \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": "507f1f77bcf86cd799439013",
    "text": "Great post! I totally agree."
  }'
```

### Get All Comments (Newest First)
```bash
curl -X GET http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439013/comments \
  -H "Authorization: Bearer TOKEN"
```

### Get Comments with Pagination (Oldest First)
```bash
curl -X GET "http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439013/comments?page=2&limit=10&sort=asc" \
  -H "Authorization: Bearer TOKEN"
```

### Delete Comment
```bash
curl -X DELETE http://localhost:8080/api/v1/comments/507f1f77bcf86cd799439020 \
  -H "Authorization: Bearer TOKEN"
```

### Stream New Comments (Real-Time)
```bash
# Terminal 1: Start streaming
curl -N http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439013/stream/comments \
  -H "Authorization: Bearer TOKEN"

# Terminal 2: Create a comment
curl -X POST http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439013/comments \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": "507f1f77bcf86cd799439013",
    "text": "This appears in Terminal 1 in real-time!"
  }'
```

### Stream Comments (JavaScript)
```javascript
const postId = '507f1f77bcf86cd799439013';
const token = localStorage.getItem('access_token');

const eventSource = new EventSource(
  `http://localhost:8080/api/v1/posts/${postId}/stream/comments`,
  { headers: { 'Authorization': `Bearer ${token}` } }
);

eventSource.addEventListener('new_comment', (event) => {
  const data = JSON.parse(event.data);
  console.log('New comment:', data.data);
  // Update UI dynamically
});

eventSource.onerror = () => {
  console.error('Connection lost');
  eventSource.close();
};
```

## Validation Rules

| Field | Rule |
|-------|------|
| post_id | Required, valid MongoDB ObjectID |
| text | Required, 1-1000 characters |

## Response Codes

| Code | Meaning |
|------|---------|
| 200 | Success (GET, DELETE) |
| 201 | Comment created successfully |
| 400 | Bad request (validation error) |
| 401 | Unauthorized (invalid token) |
| 403 | Forbidden (not comment author) |
| 404 | Post or comment not found |
| 500 | Server error |

## Response Format

### Comment Response
```json
{
  "id": "string (MongoDB ObjectID)",
  "post_id": "string",
  "user_id": "string",
  "user_name": "string",
  "text": "string",
  "created_at": "string (RFC 3339)",
  "updated_at": "string (RFC 3339)"
}
```

### SSE Event Format
```json
{
  "type": "new_comment",
  "post_id": "comment_id",
  "data": {
    "id": "comment_id",
    "post_id": "post_id",
    "user_id": "user_id",
    "user_name": "username",
    "text": "comment text",
    "created_at": "2024-07-15T14:30:00Z",
    "updated_at": "2024-07-15T14:30:00Z"
  }
}
```

### Paginated List Response
```json
{
  "data": [ /* comment objects */ ],
  "count": 2,
  "page": 1,
  "limit": 20,
  "total": 42,
  "message": "comments retrieved successfully",
  "status": 200
}
```

## Real-Time Features

✅ **SSE Streaming**: `GET /api/v1/posts/:id/stream/comments`
- Real-time comment notifications
- Works in browsers with EventSource API
- Automatic reconnection on disconnect
- Multiple concurrent clients supported

## Sorting Options

| Sort | Behavior |
|------|----------|
| `desc` (default) | Newest comments first |
| `asc` | Oldest comments first |

## Common Use Cases

1. **Display latest comments**: GET with `?sort=desc&limit=20` (default)
2. **Load more comments**: GET with `?page=2&limit=20`
3. **Real-time updates**: Subscribe with `stream/comments` endpoint
4. **Combined approach**: Fetch historical + stream new for seamless UX

## Architecture

Comments use a **pub/sub event system** for real-time delivery:

- **SSEManager**: Central event broker
- **CommentHandler**: Publishes events when comments are created
- **Client EventSource**: Subscribes to events

See [COMMENTS_API.md](COMMENTS_API.md) for full documentation.
