# API Endpoints - Quick Reference Guide

## All Endpoints Summary

### Authentication
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/auth/register` | ❌ | Register new user |
| `POST` | `/auth/login` | ❌ | Login and get tokens |
| `POST` | `/auth/refresh` | ❌ | Get new access token |
| `POST` | `/auth/logout` | ✅ | Logout and invalidate tokens |

### Profile
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/profile` | ✅ | View current user profile |
| `PATCH` | `/profile` | ✅ | Update profile (name) |
| `POST` | `/profile/change-password` | ✅ | Change password |
| `DELETE` | `/profile` | ✅ | Delete account (soft delete) |

### Rooms
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/rooms` | ✅ | Create a new room |
| `GET` | `/rooms` | ✅ | List user's rooms |
| `GET` | `/rooms/:code` | ✅ | Get room by code |
| `GET` | `/rooms/by-id/:id` | ✅ | Get room by ID |
| `GET` | `/rooms/:code/members` | ✅ | Get room member IDs |
| `GET` | `/rooms/:code/users` | ✅ | Get room member details |
| `POST` | `/rooms/join` | ✅ | Join an existing room |
| `POST` | `/rooms/add-member-by-code` | ✅ | Add user to room by user code |
| `POST` | `/rooms/remove-member-by-code` | ✅ | Remove user from room by user code (owner only) |
| `DELETE` | `/rooms/:code` | ✅ | Delete or leave a room |

### Posts
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/posts` | ✅ | Create post (multipart with media) |
| `GET` | `/posts/:id` | ✅ | Get post by ID |
| `DELETE` | `/posts/:id` | ✅ | Delete post |
| `POST` | `/posts/:id/validate` | ✅ | Validate a post |
| `DELETE` | `/posts/:id/validate` | ✅ | Remove validation |
| `POST` | `/posts/:id/respect` | ✅ | Respect a post |
| `DELETE` | `/posts/:id/respect` | ✅ | Remove respect |
| `GET` | `/posts/:id/download` | ✅ | Download post media |
| `POST` | `/posts/:id/report` | ✅ | Report post for inappropriate content |
| `GET` | `/posts/stream/new` | ✅ | Stream new posts via SSE (?room_code=) |

### Items
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/items` | ✅ | List items |
| `POST` | `/items` | ✅ | Create item |
| `GET` | `/items/:id` | ✅ | Get item by ID |
| `PUT` | `/items/:id` | ✅ | Update item |
| `DELETE` | `/items/:id` | ✅ | Delete item |
| `GET` | `/items/:id/download` | ✅ | Download item file |

> ✅ = Requires JWT Bearer token in `Authorization` header
> ❌ = No authentication required

---

## Typical User Flows

### New User - Registration & First Post
```
1. POST /auth/register → Create account
2. POST /auth/login → Get access_token
3. POST /rooms → Create a room (get room_code)
4. POST /posts → Create post with room_code
5. POST /auth/logout → End session (optional)
```

### Joining Existing Room & Viewing Posts
```
1. POST /auth/login → Get access_token
2. POST /rooms/join → Join room using code
3. GET /rooms → List your rooms
4. GET /posts/:id → View posts in room
5. POST /posts → Create your own post in room
```

### Token Expiration
```
1. Access token expires (after 1 hour)
2. POST /auth/refresh with refresh_token
3. Get new access_token, continue using
```

### Password Reset
```
1. POST /profile/change-password (with old + new password)
2. All sessions logged out automatically
3. POST /auth/login again with new password
```

### Account Deletion
```
1. DELETE /profile (with access_token)
2. All sessions logged out
3. Account marked as deleted
```

---

## Key Features by Module

### Posts - Room-Scoped & Media Support

Posts are now **scoped to rooms**. Every post must belong to a specific room, and only room members can create or view posts in that room.

**Creating a Post:**
- User must be a member of the target room
- `room_code` is required in the request body
- Supports text + optional media: image, video, audio
- Request format: `multipart/form-data`

**Post Media Support:**
- **Images:** jpg, jpeg, png, gif, webp
- **Videos:** mp4, webm, mov, avi
- **Audio:** mp3, wav, m4a, aac, flac, ogg

**Example Create Post Request:**
```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer {access_token}" \
  -F "text=Check out this!" \
  -F "room_code=room-123" \
  -F "image=@photo.jpg"
```

**Typical Flow:**
1. Join a room: `POST /rooms/join` with room code
2. Create post in that room: `POST /posts` with `room_code`
3. View post: `GET /posts/:id` (only if you're room member)

### Rooms

Users can create and join rooms to collaborate. Posts are scoped to rooms.

**Typical Flow:**
1. Create room: `POST /rooms` (you become owner)
2. Get room code: Shown in create response
3. Share code with others
4. Others join: `POST /rooms/join` with room code
5. Create posts together

---

## Status Codes

| Code | Meaning |
|------|---------|
| `200` | ✅ Success |
| `201` | ✅ Created |
| `204` | ✅ No Content |
| `400` | ❌ Bad Request (validation error) |
| `401` | ❌ Unauthorized (missing/invalid token) |
| `403` | ❌ Forbidden (not room member, insufficient permissions) |
| `404` | ❌ Not Found |
| `409` | ❌ Conflict (email/room code already exists) |
| `500` | ❌ Internal Server Error |

---

## Common Error Messages

| Error | Code | Endpoint(s) |
|-------|------|------------|
| `invalid input` | 400 | Most endpoints |
| `invalid email` | 400 | Register |
| `invalid password` | 400 | Register, Change Password |
| `age verification required` | 400 | Register |
| `email already exists` | 409 | Register |
| `invalid credentials` | 401 | Login |
| `invalid or expired refresh token` | 401 | Refresh |
| `unauthorized` | 401 | Protected endpoints |
| `not a member of this room` | 403 | Posts (create/view) |
| `room not found` | 404 | Rooms |
| `cannot join own room` | 400 | Rooms (join) |
| `post not found` | 404 | Posts |
| `user not found` | 404 | Profile, Update, Change Password |
| `room code already exists` | 409 | Create room |
| `post text is required` | 400 | Create post |
| `text exceeds maximum length` | 400 | Create post |
| `invalid image type` | 400 | Create post with image |
| `invalid video type` | 400 | Create post with video |
| `invalid audio type` | 400 | Create post with audio |
| `failed to upload image` | 500 | Create post |
| `failed to upload video` | 500 | Create post |
| `failed to upload audio` | 500 | Create post |
| `ALREADY_VALIDATED` | 409 | Validate post (already validated) |
| `NOT_VALIDATED` | 400 | Remove validation (not validated) |
| `ALREADY_RESPECTED` | 409 | Respect post (already respected) |
| `NOT_RESPECTED` | 400 | Remove respect (not respected) |

---

## Token Info

### Access Token
- **TTL:** 1 hour
- **Used For:** Authorization header in protected requests
- **Format:** `Authorization: Bearer {access_token}`

### Refresh Token
- **TTL:** 7 days
- **Used For:** Getting new access tokens
- **Storage:** Hashed in database, never exposed in responses after login

---

## Password Requirements

✅ Valid: `SecurePass123!`
- At least 8 characters
- 1+ uppercase letter (A-Z)
- 1+ lowercase letter (a-z)
- 1+ digit (0-9)
- 1+ special character (!@#$%^&*)

❌ Invalid Examples:
- `pass123` (no uppercase, no special char)
- `Pass123` (no special char)
- `Pass!` (too short, no digit)
- `PASSWORD123!` (no lowercase)

---

## Quick cURL Examples

### Authentication

#### Register
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "SecurePass123!",
    "age_verified": true
  }'
```

#### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

### Rooms

#### Create Room
```bash
curl -X POST http://localhost:8080/api/v1/rooms \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{"name": "My Awesome Room"}'
```

#### Join Room
```bash
curl -X POST http://localhost:8080/api/v1/rooms/join \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{"code": "room-code-123"}'
```

#### List Your Rooms
```bash
curl -X GET http://localhost:8080/api/v1/rooms \
  -H "Authorization: Bearer {access_token}"
```

### Posts (Room-Scoped)

#### Create Post (Text Only)
```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer {access_token}" \
  -F "text=This is my first post!" \
  -F "room_code=room-123"
```

#### Create Post with Image
```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer {access_token}" \
  -F "text=Check out this photo!" \
  -F "room_code=room-123" \
  -F "image=@/path/to/image.jpg"
```

#### Create Post with Video
```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer {access_token}" \
  -F "text=Watch this!" \
  -F "room_code=room-123" \
  -F "video=@/path/to/video.mp4"
```

#### Create Post with Audio
```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer {access_token}" \
  -F "text=Listen to this!" \
  -F "room_code=room-123" \
  -F "audio=@/path/to/podcast.mp3"
```

#### Get Post
```bash
curl -X GET http://localhost:8080/api/v1/posts/{post_id} \
  -H "Authorization: Bearer {access_token}"
```

#### Delete Post
```bash
curl -X DELETE http://localhost:8080/api/v1/posts/{post_id} \
  -H "Authorization: Bearer {access_token}"
```

### Profile

#### View Profile
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer {access_token}"
```

#### Update Profile
```bash
curl -X PATCH http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{"name": "Jane Doe"}'
```

#### Change Password
```bash
curl -X POST http://localhost:8080/api/v1/profile/change-password \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "OldPass123!",
    "new_password": "NewPass456!"
  }'
```

#### Delete Account
```bash
curl -X DELETE http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer {access_token}"
```

---

## Response Examples

### Successful Login Response
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNTA3ZjFmNzdiY2Y4NmNkNzk5NDM5MDExIiwiZW1haWwiOiJqb2huQGV4YW1wbGUuY29tIiwiZXhwIjoxNzE5NjY0NzIwLCJpYXQiOjE3MTk2NjExMjB9.abc123...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNTA3ZjFmNzdiY2Y4NmNkNzk5NDM5MDExIiwiZXhwIjoxNzI3NDQwNTIwLCJpYXQiOjE3MTk2NjExMjB9.xyz789...",
    "expires_in": 3600,
    "token_type": "Bearer"
  },
  "message": "login successful",
  "status": 200
}
```

### Create Room Response (201)
```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "name": "My Awesome Room",
    "code": "room-abc123",
    "owner_id": "507f1f77bcf86cd799439012",
    "members": ["507f1f77bcf86cd799439012"],
    "created_at": "2024-06-28T21:30:00Z",
    "updated_at": "2024-06-28T21:30:00Z"
  },
  "message": "room created successfully",
  "status": 201
}
```

### Create Post Response (201) - With Room Context
```json
{
  "data": {
    "id": "507f1f77bcf86cd799439014",
    "room_id": "507f1f77bcf86cd799439011",
    "room_code": "room-abc123",
    "room_name": "My Awesome Room",
    "user_id": "507f1f77bcf86cd799439012",
    "text": "Check out this photo!",
    "image": "posts/image/507f1f77bcf86cd799439011/abc123-sunset.jpg",
    "video": null,
    "audio": null,
    "created_at": "2024-06-28T21:32:00Z",
    "updated_at": "2024-06-28T21:32:00Z"
  },
  "message": "post created successfully",
  "status": 201
}
```

### Create Post Response (201) - With Audio
```json
{
  "data": {
    "id": "507f1f77bcf86cd799439015",
    "room_id": "507f1f77bcf86cd799439011",
    "room_code": "room-abc123",
    "room_name": "My Awesome Room",
    "user_id": "507f1f77bcf86cd799439012",
    "text": "Listen to this podcast!",
    "image": null,
    "video": null,
    "audio": "posts/audio/507f1f77bcf86cd799439011/xyz789-episode.mp3",
    "created_at": "2024-06-28T21:35:00Z",
    "updated_at": "2024-06-28T21:35:00Z"
  },
  "message": "post created successfully",
  "status": 201
}
```

### Error Response - Not Room Member (403)
```json
{
  "error": "not a member of this room",
  "status": 403
}
```

### Error Response - Missing Field (400)
```json
{
  "error": "invalid input",
  "status": 400
}
```

---

## Implementation Notes

### Understanding Room-Scoped Posts

**Key Principle:** Posts exist within rooms. Users must be room members to create or view posts.

1. **Create Room:** User becomes owner automatically
2. **Share Room Code:** Owner/members share code with others
3. **Join Room:** Others use room code to join
4. **Post in Room:** Only members can create posts (use room_code)
5. **View Posts:** Only members can see posts in the room

### Post Media Support

Posts support up to 3 media types simultaneously:
- **Image:** One per post (jpg, jpeg, png, gif, webp)
- **Video:** One per post (mp4, webm, mov, avi)
- **Audio:** One per post (mp3, wav, m4a, aac, flac, ogg)

All media is uploaded to S3 (MinIO in dev, AWS S3 in production).

### For Frontend Developers
1. Store `access_token` in memory or session storage
2. Store `refresh_token` in secure HTTP-only cookie
3. When creating posts, include `room_code` in form data
4. Display room context with post responses (room_name, room_code)
5. Handle 403 errors when user isn't room member
6. Implement room joining before post creation
7. When access token expires (401), refresh automatically
8. Clear all tokens on logout
9. Redirect to login when 401 received

### For Mobile Developers
1. Store tokens securely (Keychain on iOS, Keystore on Android)
2. Include room_code when creating posts
3. Display room information with posts
4. Check 403 responses - may need to join room first
5. Implement token refresh before expiration
6. Handle 401 responses gracefully
7. Clear tokens on logout

### For API Testing
1. Use Postman/Insomnia collections
2. Save tokens in environment variables after login
3. Create/join a room first, get the room code
4. Include room_code in post creation requests
5. Use `{{access_token}}` in protected request headers
6. Test room membership validation (403 errors)
7. Test all media types (image, video, audio)
8. Test error cases thoroughly

---

## Common Issues & Solutions

### Authentication Issues

**Issue:** `invalid credentials` on login
- **Solution:** Check email and password are correct

**Issue:** `unauthorized` on protected endpoint
- **Solution:** Include `Authorization: Bearer {token}` header

**Issue:** Token expired (401)
- **Solution:** Use refresh token to get new access token via `POST /auth/refresh`

**Issue:** `invalid or expired refresh token`
- **Solution:** Login again to get new tokens

### Room & Post Issues

**Issue:** `not a member of this room` (403) when creating post
- **Solution:** Join the room first with `POST /rooms/join` using the room code

**Issue:** `room not found` (404) when creating post
- **Solution:** Verify the room_code is correct and the room hasn't been deleted

**Issue:** `post text is required` (400)
- **Solution:** Include the `text` field in your request (1-5000 characters)

**Issue:** `text exceeds maximum length` (400)
- **Solution:** Post text must be 5000 characters or less

**Issue:** `invalid image type` (400)
- **Solution:** Image must be: jpg, jpeg, png, gif, or webp

**Issue:** `invalid video type` (400)
- **Solution:** Video must be: mp4, webm, mov, or avi

**Issue:** `invalid audio type` (400)
- **Solution:** Audio must be: mp3, wav, m4a, aac, flac, or ogg

**Issue:** Media file didn't upload but post was created
- **Solution:** Check the media file format and size; some endpoints may fail silently

### Profile Issues

**Issue:** Password change failed but showing success
- **Solution:** Verify old password is correct

**Issue:** Can't login after password change
- **Solution:** Normal behavior - all tokens invalidated, must login with new password

**Issue:** Room code already exists
- **Solution:** Generated room codes can occasionally collide; try creating room again

---

## Next Steps

For more detailed information, see:
- **Authentication:** `docs/api/01-authentication.md`
- **Posts Module:** `docs/POSTS_API.md` (includes audio support & media handling)
- **Rooms Module:** `docs/ROOMS_API.md`
- **Testing:** `docs/api/TESTING.md`
