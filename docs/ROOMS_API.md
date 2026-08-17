# Rooms Module API Documentation

This document provides comprehensive API documentation for the Rooms module, including all endpoints, request/response formats, and examples.

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

## Endpoints

### 1. Create Room

Creates a new room with the authenticated user as the owner.

**Endpoint:** `POST /api/v1/rooms`

**Authentication:** Required

**Request Body:**

```json
{
  "name": "Conference Room A",
  "code": "CONF_A_001"
}
```

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| name | string | Yes | Room name (1-100 characters) |
| code | string | Yes | Unique room code (alphanumeric, hyphen, underscore; max 50 chars) |

**Success Response (201 Created):**

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "name": "Conference Room A",
    "code": "CONF_A_001",
    "created_by": "507f1f77bcf86cd799439012",
    "created_at": "2024-06-28T21:15:00Z",
    "updated_at": "2024-06-28T21:15:00Z"
  },
  "message": "room created successfully",
  "status": 201
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid input | Missing or invalid request body |
| 409 | room code already exists | The code is already taken |
| 500 | internal server error | Server error |

**Example Error Response (409):**

```json
{
  "error": "room code already exists",
  "status": 409
}
```

---

### 2. Add User to Room

Adds the authenticated user to a room by room code.

**Endpoint:** `POST /api/v1/rooms/join`

**Authentication:** Required

**Request Body:**

```json
{
  "code": "CONF_A_001"
}
```

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| code | string | Yes | Unique room code |

**Success Response (200 OK) - New Member Added:**

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "name": "Conference Room A",
    "code": "CONF_A_001",
    "created_by": "507f1f77bcf86cd799439012",
    "members": [
      "507f1f77bcf86cd799439013",
      "507f1f77bcf86cd799439014"
    ],
    "created_at": "2024-06-28T21:15:00Z",
    "updated_at": "2024-06-28T21:30:00Z"
  },
  "message": "user added to room successfully",
  "status": 200
}
```

**Success Response (200 OK) - Already a Member:**

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "name": "Conference Room A",
    "code": "CONF_A_001",
    "created_by": "507f1f77bcf86cd799439012",
    "members": [
      "507f1f77bcf86cd799439013",
      "507f1f77bcf86cd799439014"
    ],
    "created_at": "2024-06-28T21:15:00Z",
    "updated_at": "2024-06-28T21:30:00Z"
  },
  "message": "you are already a member of this room",
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid input | Missing code |
| 400 | cannot join own room | User is the room creator |
| 404 | room not found | Room with code doesn't exist |
| 500 | internal server error | Server error |

**Example Error Response (404):**

```json
{
  "error": "room not found",
  "status": 404
}
```

**Example Error Response (400 - Creator):**

```json
{
  "error": "cannot join own room",
  "message": "you are the creator of this room and are already a member",
  "status": 400
}
```

---

### 3. Add User to Room by User Code

Adds a specified user to a room using the user's unique customer code. This endpoint allows adding a user to a room without requiring the target user's MongoDB ObjectID.

**Endpoint:** `POST /api/v1/rooms/add-member-by-code`

**Authentication:** Required

**Request Body:**

```json
{
  "room_code": "CONF_A_001",
  "user_code": "12345678"
}
```

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| room_code | string | Yes | Unique room code |
| user_code | string | Yes | Target user's 8-digit customer code |

**Success Response (200 OK) - New Member Added:**

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "name": "Conference Room A",
    "code": "CONF_A_001",
    "created_by": "507f1f77bcf86cd799439012",
    "members": [
      "507f1f77bcf86cd799439013",
      "507f1f77bcf86cd799439014",
      "507f1f77bcf86cd799439015"
    ],
    "created_at": "2024-06-28T21:15:00Z",
    "updated_at": "2024-06-28T21:30:00Z"
  },
  "message": "user added to room successfully",
  "status": 200
}
```

**Success Response (200 OK) - Already a Member:**

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "name": "Conference Room A",
    "code": "CONF_A_001",
    "created_by": "507f1f77bcf86cd799439012",
    "members": [
      "507f1f77bcf86cd799439013",
      "507f1f77bcf86cd799439014"
    ],
    "created_at": "2024-06-28T21:15:00Z",
    "updated_at": "2024-06-28T21:30:00Z"
  },
  "message": "user is already a member of this room",
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid input | Missing room_code or user_code |
| 404 | user not found | User with the specified user_code doesn't exist |
| 404 | room not found | Room with the specified room_code doesn't exist |
| 500 | internal server error | Server error |

**Example Error Response (404 - User Not Found):**

```json
{
  "error": "user not found",
  "status": 404
}
```

**Example Error Response (404 - Room Not Found):**

```json
{
  "error": "room not found",
  "status": 404
}
```

---

### 4. List User's Rooms

Retrieves all rooms the authenticated user is part of (as owner or member).

**Endpoint:** `GET /api/v1/rooms`

**Authentication:** Required

**Request Parameters:** None

**Success Response (200 OK):**

```json
{
  "data": [
    {
      "id": "507f1f77bcf86cd799439011",
      "name": "Conference Room A",
      "code": "CONF_A_001",
      "created_by": "507f1f77bcf86cd799439012",
      "members": [
        "507f1f77bcf86cd799439013",
        "507f1f77bcf86cd799439014"
      ],
      "created_at": "2024-06-28T21:15:00Z",
      "updated_at": "2024-06-28T21:30:00Z"
    },
    {
      "id": "507f1f77bcf86cd799439015",
      "name": "Meeting Room B",
      "code": "MEET_B_002",
      "created_by": "507f1f77bcf86cd799439016",
      "members": [
        "507f1f77bcf86cd799439012"
      ],
      "created_at": "2024-06-27T10:00:00Z",
      "updated_at": "2024-06-28T15:45:00Z"
    }
  ],
  "count": 2,
  "status": 200
}
```

**Empty Response (200 OK):**

```json
{
  "data": [],
  "count": 0,
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 500 | internal server error | Server error |

---

### 5. Get Room Details

Retrieves details of a specific room. Only the room owner or members can access.

**Endpoint:** `GET /api/v1/rooms/:code`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| code | string | Yes | Unique room code |

**Success Response (200 OK):**

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "name": "Conference Room A",
    "code": "CONF_A_001",
    "created_by": "507f1f77bcf86cd799439012",
    "members": [
      "507f1f77bcf86cd799439013",
      "507f1f77bcf86cd799439014"
    ],
    "created_at": "2024-06-28T21:15:00Z",
    "updated_at": "2024-06-28T21:30:00Z"
  },
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 403 | forbidden | User is not owner or member |
| 404 | room not found | Room doesn't exist |
| 500 | internal server error | Server error |

**Example Error Response (403):**

```json
{
  "error": "forbidden",
  "status": 403
}
```

---

### 6. Get Room Details by ID

Retrieves details of a specific room by its ID. Only the room owner or members can access.

**Endpoint:** `GET /api/v1/rooms/by-id/:id`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Room ID (MongoDB ObjectID hex) |

**Success Response (200 OK):**

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "name": "Conference Room A",
    "code": "CONF_A_001",
    "created_by": "507f1f77bcf86cd799439012",
    "members": [
      "507f1f77bcf86cd799439013",
      "507f1f77bcf86cd799439014"
    ],
    "created_at": "2024-06-28T21:15:00Z",
    "updated_at": "2024-06-28T21:30:00Z"
  },
  "message": "room details retrieved successfully",
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid room id | Invalid ObjectID format |
| 403 | forbidden | User is not owner or member |
| 404 | room not found | Room doesn't exist |
| 500 | internal server error | Server error |

**Example Error Response (400):**

```json
{
  "error": "invalid room id",
  "status": 400
}
```

**Example Error Response (403):**

```json
{
  "error": "you do not have permission to access this room",
  "status": 403
}
```

---

### 7. Get Room Members

Lists all members of a room. Only the room owner or members can access.

**Endpoint:** `GET /api/v1/rooms/:code/members`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| code | string | Yes | Unique room code |

**Success Response (200 OK):**

```json
{
  "data": {
    "room_id": "507f1f77bcf86cd799439011",
    "room_code": "CONF_A_001",
    "room_name": "Conference Room A",
    "owner": "507f1f77bcf86cd799439012",
    "members": [
      "507f1f77bcf86cd799439013",
      "507f1f77bcf86cd799439014",
      "507f1f77bcf86cd799439015"
    ],
    "count": 3
  },
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 403 | forbidden | User is not owner or member |
| 404 | room not found | Room doesn't exist |
| 500 | internal server error | Server error |

---

### 8. Get Room Users

Retrieves full user details for all members of a room. Only the room owner or members can access.

**Endpoint:** `GET /api/v1/rooms/:code/users`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| code | string | Yes | Unique room code |

**Success Response (200 OK):**

```json
{
  "data": [
    {
      "id": "507f1f77bcf86cd799439012",
      "code": "12345678",
      "name": "John Doe",
      "email": "john@example.com",
      "is_age_verified": true,
      "creator": true,
      "created_at": "2024-06-20T10:30:00Z"
    },
    {
      "id": "507f1f77bcf86cd799439013",
      "code": "87654321",
      "name": "Jane Smith",
      "email": "jane@example.com",
      "is_age_verified": true,
      "creator": false,
      "created_at": "2024-06-22T14:15:00Z"
    }
  ],
  "count": 2,
  "message": "room users retrieved successfully",
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 403 | forbidden | User is not owner or member |
| 404 | room not found | Room doesn't exist |
| 500 | internal server error | Server error |

**Example Error Response (403):**

```json
{
  "error": "you do not have permission to access this room",
  "status": 403
}
```

**Note:** This endpoint returns the owner and all members with their full user details (name, email, etc.), unlike `/rooms/:code/members` which only returns user IDs. Each user object includes a `creator` boolean field — `true` for the room creator, `false` for all other members.

---

### 9. Remove Member from Room

Removes a member from a room. Only the room owner can perform this action.

**Endpoint:** `POST /api/v1/rooms/:code/remove-member`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| code | string | Yes | Unique room code |

**Request Body:**

```json
{
  "member_id": "507f1f77bcf86cd799439013"
}
```

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| member_id | string | Yes | MongoDB ObjectID hex of the member to remove |

**Success Response (200 OK):**

```json
{
  "data": null,
  "message": "member removed from room successfully",
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid input | Member not found in room members list |
| 400 | invalid member id | Invalid ObjectID format |
| 403 | forbidden | User is not the room owner |
| 404 | room not found | Room doesn't exist |
| 500 | internal server error | Server error |

**Example Error Response (403):**

```json
{
  "error": "you do not have permission to perform this action",
  "status": 403
}
```

**Example Error Response (400 - Member not found):**

```json
{
  "error": "invalid input",
  "status": 400
}
```

---

### 10. Leave Room

Removes the authenticated user from a room. Only members can leave a room; room owners/creators cannot use this endpoint.

**Endpoint:** `POST /api/v1/rooms/:code/leave`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| code | string | Yes | Unique room code |

**Request Body:** None

**Success Response (200 OK):**

```json
{
  "data": null,
  "message": "you have left the room successfully",
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid input | User is not a member of the room |
| 403 | forbidden | User is the room owner/creator (cannot leave their own room) |
| 404 | room not found | Room doesn't exist |
| 500 | internal server error | Server error |

**Example Error Response (403 - Owner):**

```json
{
  "error": "room owners cannot leave their own room",
  "message": "delete the room instead if you want to remove it",
  "status": 403
}
```

**Example Error Response (400 - Not a Member):**

```json
{
  "error": "invalid input",
  "status": 400
}
```

---

### 11. Delete or Leave Room

**Deprecated**: Use `POST /api/v1/rooms/:code/leave` for members to leave a room (see Endpoint #10).

This endpoint maintains backward compatibility and handles both delete (owner) and leave (member) operations.

**Endpoint:** `DELETE /api/v1/rooms/:code`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| code | string | Yes | Unique room code |

**Success Response (200 OK) - Owner Deletes:**

```json
{
  "data": null,
  "message": "room deleted successfully",
  "status": 200
}
```

**Success Response (200 OK) - Member Leaves:**

```json
{
  "data": null,
  "message": "you have left the room successfully",
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid input | User is not a member (can't leave) |
| 404 | room not found | Room doesn't exist |
| 500 | internal server error | Server error |

**Example Error Response (400):**

```json
{
  "error": "invalid input",
  "status": 400
}
```

---

### 12. Get Posts by Room Code

Retrieves all posts for a specific room. Only room members can retrieve posts from a room.

**Endpoint:** `GET /api/v1/rooms/:code/posts`

**Authentication:** Required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| code | string | Yes | Unique room code |

**Success Response (200 OK):**

```json
{
  "data": [
    {
      "id": "507f1f77bcf86cd799439020",
      "room_id": "507f1f77bcf86cd799439011",
      "room_code": "TEAM_001",
      "room_name": "Team Meeting Room",
      "user_id": "507f1f77bcf86cd799439013",
      "text": "Hello team! Let's discuss the Q3 roadmap.",
      "image": null,
      "video": null,
      "audio": null,
      "created_at": "2024-06-28T21:15:00Z",
      "updated_at": "2024-06-28T21:15:00Z"
    },
    {
      "id": "507f1f77bcf86cd799439021",
      "room_id": "507f1f77bcf86cd799439011",
      "room_code": "TEAM_001",
      "room_name": "Team Meeting Room",
      "user_id": "507f1f77bcf86cd799439014",
      "text": "Great idea! I've already started the Q3 planning document.",
      "image": "https://s3.amazonaws.com/temp-bucket/items/...",
      "video": null,
      "audio": null,
      "created_at": "2024-06-28T21:16:30Z",
      "updated_at": "2024-06-28T21:16:30Z"
    }
  ],
  "count": 2,
  "message": "posts retrieved successfully",
  "status": 200
}
```

**Empty Response (200 OK):**

```json
{
  "data": [],
  "count": 0,
  "message": "posts retrieved successfully",
  "status": 200
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 403 | forbidden | User is not a member of the room |
| 404 | room not found | Room doesn't exist |
| 500 | internal server error | Server error |

**Example Error Response (403 - Not a Member):**

```json
{
  "error": "not a member of this room",
  "status": 403
}
```

**Example Error Response (404):**

```json
{
  "error": "room not found",
  "status": 404
}
```

---

## HTTP Status Codes

| Code | Meaning | Usage |
|------|---------|-------|
| 200 | OK | Successful GET, DELETE |
| 201 | Created | Successful POST (create/join) |
| 400 | Bad Request | Invalid input or validation error |
| 403 | Forbidden | User lacks permission to access resource |
| 404 | Not Found | Resource (room, user) doesn't exist |
| 409 | Conflict | Resource already exists (duplicate code) |
| 500 | Internal Server Error | Unexpected server error |

---

## Common Response Format

All responses follow a consistent JSON structure:

**Success (with data):**

```json
{
  "data": { /* actual response data */ },
  "message": "optional success message",
  "status": 200
}
```

**Success (without data):**

```json
{
  "data": null,
  "message": "operation successful",
  "status": 200
}
```

**Error:**

```json
{
  "error": "error message",
  "status": 400
}
```

---

## Usage Examples

### Example 1: Complete Room Workflow

#### Step 1: Create a Room

```bash
curl -X POST http://localhost:3000/api/v1/rooms \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Team Meeting Room",
    "code": "TEAM_001"
  }'
```

Response:

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "name": "Team Meeting Room",
    "code": "TEAM_001",
    "created_by": "user_123",
    "created_at": "2024-06-28T21:15:00Z",
    "updated_at": "2024-06-28T21:15:00Z"
  },
  "message": "room created successfully",
  "status": 201
}
```

#### Step 2: Another User Joins the Room

```bash
curl -X POST http://localhost:3000/api/v1/rooms/join \
  -H "Authorization: Bearer <another_user_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "TEAM_001"
  }'
```

Response:

```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "name": "Team Meeting Room",
    "code": "TEAM_001",
    "created_by": "user_123",
    "members": ["user_456"],
    "created_at": "2024-06-28T21:15:00Z",
    "updated_at": "2024-06-28T21:20:00Z"
  },
  "message": "user added to room successfully",
  "status": 201
}
```

#### Step 3: Owner Views Room Members

```bash
curl -X GET http://localhost:3000/api/v1/rooms/TEAM_001/members \
  -H "Authorization: Bearer <access_token>"
```

Response:

```json
{
  "data": {
    "room_id": "507f1f77bcf86cd799439011",
    "room_code": "TEAM_001",
    "room_name": "Team Meeting Room",
    "owner": "user_123",
    "members": ["user_456"],
    "count": 1
  },
  "status": 200
}
```

#### Step 4: List User's Rooms

```bash
curl -X GET http://localhost:3000/api/v1/rooms \
  -H "Authorization: Bearer <access_token>"
```

Response:

```json
{
  "data": [
    {
      "id": "507f1f77bcf86cd799439011",
      "name": "Team Meeting Room",
      "code": "TEAM_001",
      "created_by": "user_123",
      "members": ["user_456"],
      "created_at": "2024-06-28T21:15:00Z",
      "updated_at": "2024-06-28T21:20:00Z"
    }
  ],
  "count": 1,
  "status": 200
}
```

#### Step 5: Member Leaves Room

```bash
curl -X DELETE http://localhost:3000/api/v1/rooms/TEAM_001 \
  -H "Authorization: Bearer <another_user_token>"
```

Response:

```json
{
  "data": null,
  "message": "you have left the room successfully",
  "status": 200
}
```

#### Step 6: Owner Deletes Room

```bash
curl -X DELETE http://localhost:3000/api/v1/rooms/TEAM_001 \
  -H "Authorization: Bearer <access_token>"
```

Response:

```json
{
  "data": null,
  "message": "room deleted successfully",
  "status": 200
}
```

---

## Notes for Frontend Engineers

### Authentication

- All endpoints require a valid JWT access token in the `Authorization` header
- Tokens are obtained via `POST /api/v1/auth/login`
- Tokens expire after 1 hour; use refresh token to get a new one
- If you receive a 401 Unauthorized, the token is invalid or expired

### Room Codes

- Must be unique across all rooms
- Alphanumeric characters, hyphens (`-`), and underscores (`_`) are allowed
- Spaces and special characters are not allowed
- Maximum 50 characters
- Case-sensitive (e.g., `CONF_A_001` and `conf_a_001` are different)

### Authorization

- **Create Room**: User must be authenticated
- **Join Room**: User must be authenticated
- **View Room/Members**: User must be room owner OR member
- **Delete/Leave Room**: 
  - If owner: soft-deletes the room (visible to no one)
  - If member: removes user from room members

### Soft Delete

- When a room is deleted by the owner, it's marked as deleted (soft-deleted)
- Soft-deleted rooms don't appear in any queries
- Data is preserved in the database for audit/compliance purposes

### Response Formats

- All timestamps are in UTC and ISO 8601 format
- All IDs are MongoDB ObjectID hex strings (24 characters)
- Member arrays contain user IDs; use these to identify users in your application

---

## Troubleshooting

### 403 Forbidden on GET /rooms/:code or GET /rooms/:code/members

**Problem**: User is trying to access a room they're not part of.

**Solution**: Only room owners and members can view room details and member lists. Ensure the authenticated user is either:
- The room creator (`created_by` field)
- In the room's members list

### 409 Conflict on POST /rooms

**Problem**: Room code already exists.

**Solution**: Choose a different, unique room code.

### 404 Not Found

**Problem**: Room with the specified code doesn't exist.

**Solution**: Verify the room code is correct and the room hasn't been deleted.

### 400 Bad Request on DELETE /rooms/:code (for member)

**Problem**: User tried to leave a room they're not a member of.

**Solution**: Only members can leave. If the user created the room, they must delete it instead.

---

## Future Enhancements

- Room descriptions and topics
- Room metadata (capacity, location, etc.)
- Room search and filtering
- Room activity logs
- Real-time member notifications
- Voice/video integration
