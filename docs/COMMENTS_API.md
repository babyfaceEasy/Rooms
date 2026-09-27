# Comments Module API Documentation

This document provides comprehensive API documentation for the Comments module, including all endpoints, request/response formats, and real-time streaming examples.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

All endpoints in this module require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <access_token>
```

The JWT token is obtained from the login endpoint (`POST /api/v1/auth/login`).

---

## Overview

The Comments module allows users to create, retrieve, and delete comments on posts. Comments are post-scoped and support real-time streaming via Server-Sent Events (SSE) for live comment updates.

---

## Endpoints

### 1. Create Comment

Creates a new comment on a post.

**Endpoint:** `POST /api/v1/posts/:id/comments`

**Authentication:** Required

**Content-Type:** `application/json`

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Post ID (MongoDB ObjectID) |

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| post_id | string | Yes | Post ID (must match URL parameter) |
| text | string | Yes | Comment text (1-1000 characters) |

**Example Request:**

```bash
curl -X POST http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439013/comments \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": "507f1f77bcf86cd799439013",
    "text": "Great post! I totally agree with this."
  }'
```

**Success Response (201 Created):**

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439020",
    "post_id": "507f1f77bcf86cd799439013",
    "user_id": "507f1f77bcf86cd799439012",
    "user_name": "Jane Doe",
    "text": "Great post! I totally agree with this.",
    "created_at": "2024-07-15T14:30:00Z",
    "updated_at": "2024-07-15T14:30:00Z"
  },
  "message": "comment created successfully",
  "status": 201
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid input | Missing required fields or invalid request format |
| 400 | post_id is required | Post ID not provided |
| 400 | invalid post id | Invalid MongoDB ObjectID format |
| 400 | text is required | Text field is empty |
| 400 | text must be at least 1 character | Text too short |
| 400 | text must not exceed 1000 characters | Text too long |
| 401 | unauthorized | Missing or invalid JWT token |
| 404 | post not found | Post does not exist or is soft-deleted |
| 500 | internal server error | Server error |

**Example Error Response (400 - Text Too Long):**

```json
{
  "error": "text must not exceed 1000 characters",
  "status": 400
}
```

**Real-Time Event (SSE):**

When a comment is created, all clients subscribed to the post's comment stream will receive an event:

```json
{
  "type": "new_comment",
  "post_id": "507f1f77bcf86cd799439020",
  "data": {
    "id": "507f1f77bcf86cd799439020",
    "post_id": "507f1f77bcf86cd799439013",
    "user_id": "507f1f77bcf86cd799439012",
    "user_name": "Jane Doe",
    "text": "Great post! I totally agree with this.",
    "created_at": "2024-07-15T14:30:00Z",
    "updated_at": "2024-07-15T14:30:00Z"
  }
}
```

---

### 2. Get Comments for Post

Retrieves paginated comments for a specific post.

**Endpoint:** `GET /api/v1/posts/:id/comments`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Post ID (MongoDB ObjectID) |
| page | integer | No | Page number (default: 1, minimum: 1) |
| limit | integer | No | Items per page (default: 20, min: 1, max: 100) |
| sort | string | No | Sort order: `desc` (newest first, default) or `asc` (oldest first) |

**Example Request:**

```bash
# Get first page of comments (newest first)
curl -X GET "http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439013/comments" \
  -H "Authorization: Bearer <access_token>"

# Get page 2 with custom limit, sorted oldest first
curl -X GET "http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439013/comments?page=2&limit=10&sort=asc" \
  -H "Authorization: Bearer <access_token>"
```

**Success Response (200 OK):**

```json
{
  "data": [
    {
      "id": "507f1f77bcf86cd799439020",
      "post_id": "507f1f77bcf86cd799439013",
      "user_id": "507f1f77bcf86cd799439012",
      "user_name": "Jane Doe",
      "text": "Great post! I totally agree with this.",
      "created_at": "2024-07-15T14:30:00Z",
      "updated_at": "2024-07-15T14:30:00Z"
    },
    {
      "id": "507f1f77bcf86cd799439021",
      "post_id": "507f1f77bcf86cd799439013",
      "user_id": "507f1f77bcf86cd799439014",
      "user_name": "Bob Smith",
      "text": "Thanks for sharing! This is really helpful.",
      "created_at": "2024-07-15T14:35:00Z",
      "updated_at": "2024-07-15T14:35:00Z"
    }
  ],
  "count": 2,
  "page": 1,
  "limit": 20,
  "total": 12,
  "message": "comments retrieved successfully",
  "status": 200
}
```

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| data | array | Array of Comment objects |
| count | integer | Number of comments on this page |
| page | integer | Current page number |
| limit | integer | Items per page |
| total | integer | Total number of comments for the post |
| message | string | Success message |
| status | integer | HTTP status code |

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

### 3. Stream Comment Events (SSE)

Streams real-time comment events for a specific post using Server-Sent Events.

**Endpoint:** `GET /api/v1/posts/:id/stream/comments`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Post ID (MongoDB ObjectID) |

**Response Headers:**

```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
X-Accel-Buffering: no
```

**HTTP Status:** 200 OK (streaming response)

**Example Request (curl):**

```bash
# Stream comment events in real-time (use -N flag to disable buffering)
curl -N "http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439013/stream/comments" \
  -H "Authorization: Bearer <access_token>"
```

**JavaScript/Browser Example:**

```javascript
const postId = '507f1f77bcf86cd799439013';
const token = localStorage.getItem('access_token');

const eventSource = new EventSource(
  `http://localhost:8080/api/v1/posts/${postId}/stream/comments`,
  {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  }
);

eventSource.addEventListener('new_comment', (event) => {
  const data = JSON.parse(event.data);
  console.log('New comment event:', data);
  
  if (data.type === 'new_comment') {
    console.log('Comment ID:', data.post_id);
    console.log('Comment data:', data.data);
    
    // Update UI with new comment
    const comment = data.data;
    const commentElement = document.createElement('div');
    commentElement.className = 'comment';
    commentElement.innerHTML = `
      <strong>${comment.user_name}</strong>
      <p>${comment.text}</p>
      <small>${new Date(comment.created_at).toLocaleString()}</small>
    `;
    document.getElementById('comments-container').appendChild(commentElement);
  }
});

eventSource.onerror = (error) => {
  console.error('SSE connection error:', error);
  eventSource.close();
};

// Clean up when done
window.addEventListener('beforeunload', () => {
  eventSource.close();
});
```

**Event Response Format:**

Events are streamed as JSON objects with the following structure:

```json
{
  "type": "new_comment",
  "post_id": "507f1f77bcf86cd799439020",
  "data": {
    "id": "507f1f77bcf86cd799439020",
    "post_id": "507f1f77bcf86cd799439013",
    "user_id": "507f1f77bcf86cd799439012",
    "user_name": "Jane Doe",
    "text": "Great post! I totally agree with this.",
    "created_at": "2024-07-15T14:30:00Z",
    "updated_at": "2024-07-15T14:30:00Z"
  }
}
```

**Event Types:**

| Type | Trigger | Description |
|------|---------|-------------|
| `new_comment` | Comment created | Emitted when a new comment is added to the post |

**Multi-Client Example:**

```javascript
// Server can handle multiple clients streaming the same post
// Each client independently receives events as they occur

// Client 1: Listening for comments
const client1 = new EventSource(url);
client1.addEventListener('new_comment', (e) => {
  console.log('Client 1 received:', e.data);
});

// Client 2: Also listening for the same post
const client2 = new EventSource(url);
client2.addEventListener('new_comment', (e) => {
  console.log('Client 2 received:', e.data);
});

// When someone creates a comment:
// curl -X POST ... creates comment
// Both Client 1 and Client 2 receive the event simultaneously
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid post id | Invalid MongoDB ObjectID format |
| 401 | unauthorized | Missing or invalid JWT token |
| 404 | post not found | Post does not exist or is soft-deleted |
| 500 | internal server error | Server error |

**Notes:**

- The connection remains open and streams events in real-time as comments are created on the post
- Connection may drop due to network issues or client timeout; implement reconnection logic on the client side
- Each client connection consumes one subscription; connections are cleaned up automatically when the client disconnects
- The endpoint respects JWT authentication
- Multiple clients can stream the same post simultaneously and all receive the same events
- Events are delivered in real-time with minimal latency (typically < 100ms)

---

### 4. Delete Comment

Deletes a comment. Only the comment author can delete their own comment.

**Endpoint:** `DELETE /api/v1/comments/:id`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Comment ID (MongoDB ObjectID) |

**Example Request:**

```bash
curl -X DELETE http://localhost:8080/api/v1/comments/507f1f77bcf86cd799439020 \
  -H "Authorization: Bearer <access_token>"
```

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

---

## Data Models

### Comment

```json
{
  "id": "string (MongoDB ObjectID)",
  "post_id": "string (MongoDB ObjectID of parent post)",
  "user_id": "string (MongoDB ObjectID of creator)",
  "user_name": "string (creator's display name)",
  "text": "string (1-1000 characters)",
  "created_at": "string (ISO 8601 timestamp)",
  "updated_at": "string (ISO 8601 timestamp)",
  "deleted_at": "string (ISO 8601 timestamp) or null"
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique comment identifier |
| post_id | string | Parent post ID |
| user_id | string | Creator's user ID |
| user_name | string | Creator's display name |
| text | string | Comment content |
| created_at | string | Creation timestamp (RFC 3339 format) |
| updated_at | string | Last update timestamp (RFC 3339 format) |
| deleted_at | string\|null | Soft delete timestamp (null if active) |

---

## Common Use Cases

### 1. Create a Simple Comment

```bash
curl -X POST http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439013/comments \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": "507f1f77bcf86cd799439013",
    "text": "Great post!"
  }'
```

### 2. Get All Comments for a Post

```bash
curl -X GET "http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439013/comments" \
  -H "Authorization: Bearer TOKEN"
```

### 3. Get Comments with Pagination (Oldest First)

```bash
curl -X GET "http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439013/comments?page=2&limit=10&sort=asc" \
  -H "Authorization: Bearer TOKEN"
```

### 4. Delete a Comment

```bash
curl -X DELETE http://localhost:8080/api/v1/comments/507f1f77bcf86cd799439020 \
  -H "Authorization: Bearer TOKEN"
```

### 5. Stream Comments in Real-Time

**Terminal 1: Start streaming comments**

```bash
curl -N "http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439013/stream/comments" \
  -H "Authorization: Bearer TOKEN"
```

**Terminal 2: Create a comment**

```bash
curl -X POST http://localhost:8080/api/v1/posts/507f1f77bcf86cd799439013/comments \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": "507f1f77bcf86cd799439013",
    "text": "This will appear in terminal 1!"
  }'
```

**Result in Terminal 1:** Event is received in real-time

```
data: {"type":"new_comment","post_id":"507f1f77bcf86cd799439020","data":{"id":"507f1f77bcf86cd799439020","post_id":"507f1f77bcf86cd799439013",...}}
```

---

## Validation Rules

| Field | Rule |
|-------|------|
| post_id | Required, valid MongoDB ObjectID |
| text (comment) | Required, 1-1000 characters |

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

---

## Rate Limiting

Currently no rate limiting is enforced on comment endpoints. This may be added in future versions.

---

## WebSocket vs SSE

This module uses **Server-Sent Events (SSE)** instead of WebSockets:

**Why SSE?**
- ✅ Simpler implementation and lower server resource usage
- ✅ Automatic reconnection handling
- ✅ Built into browser EventSource API
- ✅ Works over HTTP/1.1 and HTTP/2
- ✅ Unidirectional (server → client), which fits our use case

**Limitations of SSE:**
- ❌ One-way communication (we publish events, not receive commands)
- ❌ Browser same-origin policy applies

For comment events, SSE is ideal since we only need to push new comments to subscribers.

---

## Architecture Notes

### Real-Time Event System

The comment streaming feature uses a pub/sub event system:

1. **SSEManager**: Central event broker that manages subscriptions
2. **CommentHandler**: Publishes events when comments are created
3. **Subscribers**: Connected clients receive events in real-time

### Key Implementation Details

- Events are stored in buffered channels (10-item capacity) to prevent blocking
- Non-blocking send: if a subscriber's channel is full, the event is skipped
- Each post_id has its own subscriber map
- Subscriptions are cleaned up automatically when clients disconnect
- No database polling; events pushed immediately from service layer

### Extensibility

This architecture supports adding more event types:
- `comment_deleted` - when a comment is deleted
- `post_validated` - when a post receives validation
- `post_respected` - when a post receives respect
- `comment_edited` - when a comment is edited (if feature added)

Future implementations can add these event types without architectural changes.
