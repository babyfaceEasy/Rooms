# Rooms Module - Quick Reference

## API Endpoints Summary

| Method | Endpoint | Description | Auth | Returns |
|--------|----------|-------------|------|---------|
| **POST** | `/api/v1/rooms` | Create room | ✅ | Room object |
| **POST** | `/api/v1/rooms/join` | Join room by code | ✅ | Room object |
| **GET** | `/api/v1/rooms` | List user's rooms | ✅ | Array of rooms |
| **GET** | `/api/v1/rooms/:code` | Get room details by code | ✅ | Room object |
| **GET** | `/api/v1/rooms/by-id/:id` | Get room details by ID | ✅ | Room object |
| **GET** | `/api/v1/rooms/:code/members` | List member IDs | ✅ | Member ID array |
| **GET** | `/api/v1/rooms/:code/users` | List member details | ✅ | User details array |
| **GET** | `/api/v1/rooms/:code/posts` | Get all posts in room | ✅ | Array of posts |
| **POST** | `/api/v1/rooms/:code/remove-member` | Remove member (owner only) | ✅ | Empty (null) |
| **POST** | `/api/v1/rooms/:code/leave` | Leave room (member only) | ✅ | Empty (null) |
| **DELETE** | `/api/v1/rooms/:code` | Delete/leave room | ✅ | Empty (null) |

---

## Quick Examples

### Create Room

```bash
curl -X POST http://localhost:3000/api/v1/rooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"My Room","code":"MY_ROOM_001"}'
```

### List My Rooms

```bash
curl http://localhost:3000/api/v1/rooms \
  -H "Authorization: Bearer $TOKEN"
```

### Join Room

```bash
curl -X POST http://localhost:3000/api/v1/rooms/join \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code":"MY_ROOM_001"}'
```

### View Room by Code

```bash
curl http://localhost:3000/api/v1/rooms/MY_ROOM_001 \
  -H "Authorization: Bearer $TOKEN"
```

### View Room by ID

```bash
curl http://localhost:3000/api/v1/rooms/by-id/507f1f77bcf86cd799439011 \
  -H "Authorization: Bearer $TOKEN"
```

### View Members (IDs only)

```bash
curl http://localhost:3000/api/v1/rooms/MY_ROOM_001/members \
  -H "Authorization: Bearer $TOKEN"
```

### View Users (Full Details)

```bash
curl http://localhost:3000/api/v1/rooms/MY_ROOM_001/users \
  -H "Authorization: Bearer $TOKEN"
```

### Get All Posts in Room

```bash
curl http://localhost:3000/api/v1/rooms/MY_ROOM_001/posts \
  -H "Authorization: Bearer $TOKEN"
```

### Remove Member from Room

```bash
curl -X POST http://localhost:3000/api/v1/rooms/MY_ROOM_001/remove-member \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"member_id":"507f1f77bcf86cd799439013"}'
```

### Leave Room

```bash
curl -X POST http://localhost:3000/api/v1/rooms/MY_ROOM_001/leave \
  -H "Authorization: Bearer $TOKEN"
```

### Leave/Delete Room

```bash
curl -X DELETE http://localhost:3000/api/v1/rooms/MY_ROOM_001 \
  -H "Authorization: Bearer $TOKEN"
```

---

## Response Schema

### Room Object

```javascript
{
  "id": "string (ObjectID hex)",
  "name": "string",
  "code": "string (unique)",
  "created_by": "string (user ID)",
  "members": ["string (user ID)", ...],
  "created_at": "string (ISO 8601)",
  "updated_at": "string (ISO 8601)"
}
```

### Members Response

```javascript
{
  "room_id": "string",
  "room_code": "string",
  "room_name": "string",
  "owner": "string (user ID)",
  "members": ["string (user ID)", ...],
  "count": number
}
```

### User Details Response

```javascript
{
  "id": "string (ObjectID hex)",
  "name": "string",
  "email": "string",
  "is_age_verified": boolean,
  "created_at": "string (ISO 8601)"
}
```

---

## Status Codes

- **200**: Success (GET, DELETE)
- **201**: Created (POST create/join)
- **400**: Bad request / validation error
- **403**: Forbidden (no access)
- **404**: Not found
- **409**: Conflict (duplicate code)
- **500**: Server error

---

## Validation Rules

### Room Name
- Required
- 1-100 characters
- Cannot be empty or whitespace

### Room Code
- Required
- Alphanumeric, hyphen (`-`), underscore (`_`) only
- Max 50 characters
- Must be unique
- No spaces or special characters allowed

---

## Key Features

✅ **Create Rooms** — Set name and unique code
✅ **Join Rooms** — Add yourself to existing room
✅ **List Rooms** — See all rooms you're in
✅ **View Details** — Get room info (owner/member only)
✅ **Member Management** — See who's in the room
✅ **Leave/Delete** — Exit or remove room
✅ **Soft Delete** — Data preserved, not actually removed
✅ **Authorization** — Owner/member access control

---

## Error Examples

### Invalid Code (Duplicate)

```json
{
  "error": "room code already exists",
  "status": 409
}
```

### Unauthorized Access

```json
{
  "error": "forbidden",
  "status": 403
}
```

### Not Found

```json
{
  "error": "room not found",
  "status": 404
}
```

### Invalid Input

```json
{
  "error": "invalid input",
  "status": 400
}
```

---

## Tips

1. **Store the token** after login for all authenticated requests
2. **Use room codes** as identifiers for users (more memorable than IDs)
3. **Check response status** to handle errors appropriately
4. **Members array** shows only direct members, not the owner
5. **Soft delete** means deleted rooms are gone from queries but data exists
