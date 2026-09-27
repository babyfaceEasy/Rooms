# Rooms Backend API

A production-ready Go backend API for a social media application with real-time room collaboration, user authentication, and content sharing capabilities.

## Overview

Rooms is a scalable backend service that provides:
- **User Authentication** — JWT-based authentication with secure session management
- **Room Management** — Create collaborative spaces, manage members, and control access
- **Content Sharing** — Post text, images, and videos with S3-compatible file storage
- **User Profiles** — Manage user information, passwords, and account settings

Built with **Go**, **Fiber**, **MongoDB**, and **AWS S3/MinIO**.

## Features

### 🔐 Authentication & Security
- **JWT Authentication** — Secure token-based authentication with 1-hour access tokens
- **Refresh Tokens** — Long-lived refresh tokens with 7-day TTL and multi-device support
- **Password Security** — bcrypt hashing with salt
- **Token Invalidation** — Logout invalidates all user sessions across devices
- **Protected Routes** — Middleware-enforced JWT validation on sensitive endpoints

### 👥 User Management
- **Registration** — New user signup with email verification requirements
- **Profile Management** — View and update user profile information
- **Password Management** — Change password with automatic token invalidation
- **Account Deletion** — Soft-delete accounts while preserving data for compliance

### 🏠 Room Management
- **Create Rooms** — Create collaborative spaces with unique codes
- **Member Management** — Add users to rooms by room code
- **Room Access Control** — Owner and member visibility controls
- **Leave Room** — Users can remove themselves from rooms
- **Room Deletion** — Owners can soft-delete entire rooms
- **List User Rooms** — View all rooms a user has created or joined

### 📝 Content & Posts
- **Create Posts** — Share text content with optional multimedia
- **Image Uploads** — Upload and store images with S3/MinIO
- **Video Uploads** — Upload and store videos with S3/MinIO
- **Post Retrieval** — Fetch post details with creator information
- **Post Deletion** — Soft-delete posts with ownership verification
- **File Validation** — Automatic file type checking before upload

## Architecture

### Clean Architecture Pattern

The project follows Clean Architecture principles with clear separation of concerns:

```
internal/
├── domain/              # Domain entities and business rules
│   ├── user.go
│   ├── refresh_token.go
│   ├── room.go
│   └── post.go
├── repository/          # Data persistence layer (MongoDB, S3)
│   ├── user_repository.go
│   ├── refresh_token_repository.go
│   ├── room_repository.go
│   ├── post_repository.go
│   └── s3_repository.go (ObjectStorage)
├── service/             # Business logic layer
│   ├── user_service.go
│   ├── auth_service.go
│   ├── room_service.go
│   └── post_service.go
├── handler/             # HTTP request handlers
│   ├── user_handler.go
│   ├── auth_handler.go
│   ├── room_handler.go
│   └── post_handler.go
├── middleware/          # HTTP middleware
│   ├── auth_middleware.go
│   └── error_handler.go
└── api/                 # API configuration
    ├── server.go
    └── routes.go

pkg/
├── mongodb/             # MongoDB client initialization
└── s3/                  # S3/MinIO client initialization
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` — Register new user
- `POST /api/v1/auth/login` — Login and receive tokens
- `POST /api/v1/auth/refresh` — Refresh access token
- `POST /api/v1/auth/logout` — Logout and invalidate tokens

### User Management
- `GET /api/v1/profile` — View authenticated user's profile
- `PATCH /api/v1/profile` — Update profile information
- `POST /api/v1/profile/change-password` — Change password
- `DELETE /api/v1/profile` — Delete account (soft delete)

### Rooms
- `POST /api/v1/rooms` — Create a new room
- `POST /api/v1/rooms/join` — Add user to room by code
- `GET /api/v1/rooms` — List all user's rooms
- `GET /api/v1/rooms/:code` — Get room details
- `GET /api/v1/rooms/:code/members` — List room members
- `DELETE /api/v1/rooms/:code` — Delete room or leave room

### Posts
- `POST /api/v1/posts` — Create post with optional image/video
- `GET /api/v1/posts/:id` — Retrieve post details
- `DELETE /api/v1/posts/:id` — Delete post (soft delete)

## Getting Started

### Prerequisites

- **Go** 1.21+
- **MongoDB** 4.4+
- **Docker** and **Docker Compose** (for local development)
- **Git**

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/babyfaceEasy/Rooms.git
   cd temp_backend
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

### Environment Variables

```bash
# App Configuration
APP_NAME=Rooms
APP_PORT=3000
LOG_LEVEL=info

# Database
MONGO_URI=mongodb://localhost:27017
MONGO_DB=rooms

# JWT
JWT_SECRET=your-super-secret-key-min-32-characters-long
ACCESS_TOKEN_TTL=1h
REFRESH_TOKEN_TTL=168h

# S3 / MinIO
S3_ENDPOINT_URL=http://localhost:9000      # MinIO for dev
S3_REGION=us-east-1
S3_BUCKET=rooms
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
```

### Quick Start with Docker Compose

```bash
# Start all services (API, MongoDB, MinIO)
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop services
docker-compose down
```

### Manual Local Development

1. **Start MongoDB**
   ```bash
   # Using Docker
   docker run -d -p 27017:27017 --name mongodb mongo:latest

   # Or use your local MongoDB installation
   mongod
   ```

2. **Start MinIO** (for S3 testing)
   ```bash
   docker run -d -p 9000:9000 -p 9001:9001 \
     -e MINIO_ROOT_USER=minioadmin \
     -e MINIO_ROOT_PASSWORD=minioadmin \
     minio/minio server /data --console-address ":9001"
   ```

3. **Build and run the backend**
   ```bash
   go build -o temp_backend cmd/api/main.go
   ./temp_backend
   ```

4. **Access the API**
   ```bash
   curl http://localhost:3000/health
   ```

## Testing

### Run All Tests
```bash
go test ./...
```

### Run Specific Module Tests
```bash
# Authentication tests
go test ./internal/service -v -k Auth

# Room tests
go test ./internal/service -v -k Room

# Post tests
go test ./internal/service -v -k Post

# Handler tests
go test ./internal/handler -v
```

### Test Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Current Test Status
- **Unit Tests**: 150+ passing
- **Domain Layer**: 30+ tests (validation, entities)
- **Service Layer**: 60+ tests (business logic, authorization)
- **Handler Layer**: 50+ tests (HTTP endpoints, error handling)
- **Integration**: Multi-layer coverage

## API Documentation

### Main Documentation
- **Full API Reference** — See `docs/api/` for comprehensive endpoint documentation
- **Authentication Guide** — `docs/api/01-authentication.md`

### Module-Specific Docs
- **Rooms API** — `docs/ROOMS_API.md` and `docs/ROOMS_QUICK_REFERENCE.md`
- **Posts API** — `docs/POSTS_API.md` and `docs/POSTS_QUICK_REFERENCE.md`

### Quick Testing
For quick testing endpoints, see individual module documentation with curl examples:

```bash
# Create a user
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "SecurePass123!",
    "is_over_13": true
  }'

# Login
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'

# Create a room (use access token from login)
curl -X POST http://localhost:3000/api/v1/rooms \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Study Group",
    "code": "STUDY_001"
  }'

# Create a post with image
curl -X POST http://localhost:3000/api/v1/posts \
  -H "Authorization: Bearer <access_token>" \
  -F "text=My first post!" \
  -F "image=@/path/to/image.jpg"
```

## Database Schema

### Collections

#### users
```json
{
  "_id": ObjectId,
  "name": "string",
  "email": "string (unique)",
  "password": "string (hashed with bcrypt)",
  "is_over_13": boolean,
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "deleted_at": "timestamp or null"
}
```

#### refresh_tokens
```json
{
  "_id": ObjectId,
  "user_id": ObjectId,
  "token_hash": "string (SHA256)",
  "device_id": "string (optional)",
  "expires_at": "timestamp",
  "created_at": "timestamp"
}
```

#### rooms
```json
{
  "_id": ObjectId,
  "name": "string",
  "code": "string (unique)",
  "created_by": ObjectId,
  "members": [ObjectId],
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "deleted_at": "timestamp or null"
}
```

#### posts
```json
{
  "_id": ObjectId,
  "user_id": ObjectId,
  "text": "string (1-5000 chars)",
  "image": "string (S3 path) or null",
  "video": "string (S3 path) or null",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "deleted_at": "timestamp or null"
}
```

## File Storage

### S3/MinIO Storage Structure

Files are organized by type and user:

```
posts/
├── images/
│   └── {user_id}/
│       ├── sunset.jpg
│       ├── vacation.png
│       └── ...
└── videos/
    └── {user_id}/
        ├── tutorial.mp4
        ├── vlog.webm
        └── ...
```

### Supported File Types

| Type | Extensions |
|------|-----------|
| Images | jpg, jpeg, png, gif, webp |
| Videos | mp4, webm, mov, avi |

### Local Development (MinIO)

MinIO console available at `http://localhost:9001`

**Credentials (from docker-compose):**
- Username: `minioadmin`
- Password: `minioadmin`

## Project Structure

```
temp_backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── server.go            # Fiber server setup
│   │   └── routes.go            # Route registration
│   ├── domain/                  # Domain entities
│   │   ├── user.go
│   │   ├── refresh_token.go
│   │   ├── room.go
│   │   └── post.go
│   ├── repository/              # Data access layer
│   │   ├── user_repository.go
│   │   ├── refresh_token_repository.go
│   │   ├── room_repository.go
│   │   ├── post_repository.go
│   │   └── s3_repository.go
│   ├── service/                 # Business logic
│   │   ├── user_service.go
│   │   ├── auth_service.go
│   │   ├── room_service.go
│   │   └── post_service.go
│   ├── handler/                 # HTTP handlers
│   │   ├── user_handler.go
│   │   ├── auth_handler.go
│   │   ├── room_handler.go
│   │   └── post_handler.go
│   ├── middleware/              # HTTP middleware
│   │   ├── auth_middleware.go
│   │   └── error_handler.go
│   └── config/                  # Configuration
│       └── config.go
├── pkg/
│   ├── mongodb/                 # MongoDB initialization
│   │   └── mongodb.go
│   └── s3/                      # S3/MinIO initialization
│       └── s3.go
├── config/
│   └── config.go                # Config loading
├── docs/                        # API documentation
│   ├── POSTS_API.md
│   ├── POSTS_QUICK_REFERENCE.md
│   ├── ROOMS_API.md
│   ├── ROOMS_QUICK_REFERENCE.md
│   └── api/
├── tests/                       # Integration tests
│   └── e2e/
├── docker-compose.yml           # Docker Compose setup
├── Dockerfile                   # Docker image
├── go.mod                       # Go module dependencies
└── README.md                    # This file
```

## Dependencies

### Core
- **Fiber** — High-performance Go web framework
- **MongoDB Driver** — MongoDB Go client
- **AWS SDK v2** — S3 client for file storage

### Security & Auth
- **JWT** — Token-based authentication
- **bcrypt** — Password hashing

### Utilities
- **Go standard library** — context, time, fmt, log

See `go.mod` for complete dependency list.

## Build & Deployment

### Build for Local Development
```bash
go build -o temp_backend cmd/api/main.go
```

### Build for Production
```bash
go build -ldflags="-s -w" -o temp_backend cmd/api/main.go
```

### Docker Build
```bash
docker build -t rooms-backend:latest .
docker run -p 3000:3000 --env-file .env rooms-backend:latest
```

## Security Considerations

### Authentication
- ✅ JWT tokens with HS256 signing
- ✅ Short-lived access tokens (1 hour)
- ✅ Long-lived refresh tokens with rotation
- ✅ Bearer token extraction from Authorization header
- ✅ Constant-time password comparison

### Data Protection
- ✅ bcrypt password hashing with salt
- ✅ Soft delete for GDPR compliance
- ✅ User-namespaced file storage in S3
- ✅ Refresh tokens stored as SHA256 hashes

### Authorization
- ✅ Middleware-enforced JWT validation
- ✅ Owner-only operations (delete room, delete post)
- ✅ Member access control for rooms
- ✅ Multi-layer authorization checks

### File Upload
- ✅ File type validation before upload
- ✅ Extension-based validation
- ✅ S3 bucket isolation
- ✅ User-namespaced paths prevent overwrites

## Performance

### Optimizations
- **Connection Pooling** — MongoDB connection pool management
- **Indexing** — Unique indexes on email and room codes
- **Soft Delete Filtering** — Indexed queries exclude deleted records
- **Token Caching** — JWT validation optimized with proper TTLs
- **Multipart Streaming** — Large file uploads handled efficiently

### Scaling Considerations
- Stateless service design for horizontal scaling
- MongoDB replica sets for high availability
- S3 for distributed file storage
- Redis could be added for caching and sessions

## Roadmap

### Phase 1: Core Features ✅
- [x] User registration and authentication
- [x] JWT tokens with refresh capability
- [x] Room creation and member management
- [x] Post creation with image/video support
- [x] Comprehensive API documentation
- [x] Unit and integration tests

### Phase 2: Enhancements (Planned)
- [ ] Post feed with pagination
- [ ] Like/unlike posts
- [ ] Comments on posts
- [ ] Direct messaging
- [ ] Notifications
- [ ] User search and discovery

### Phase 3: Advanced Features (Planned)
- [ ] Email verification
- [ ] Password reset
- [ ] Multi-factor authentication
- [ ] OAuth2 integration (Google, GitHub)
- [ ] Real-time WebSocket updates
- [ ] Analytics and metrics

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Create a feature branch (`git checkout -b feat/your-feature`)
2. Commit your changes with clear messages (`git commit -m "feat: describe feature"`)
3. Write tests for new features
4. Ensure all tests pass (`go test ./...`)
5. Create a pull request with a detailed description

## License

This project is licensed under the MIT License - see LICENSE file for details.

## Support

For issues, questions, or suggestions:
- Open an issue on GitHub
- Check existing documentation in `docs/`
- Review API reference for endpoint details

## Authors

**Team:** babyfaceEasy/Rooms  
**Backend Lead:** @babyfaceEasy  
**Technologies:** Go, Fiber, MongoDB, S3/MinIO

## Changelog

### Version 1.0.0 (Current)
- ✅ Complete authentication module with JWT
- ✅ User management (registration, profile, password, deletion)
- ✅ Room management with member controls
- ✅ Posts module with S3 image/video support
- ✅ 150+ unit tests passing
- ✅ Comprehensive API documentation
- ✅ Clean Architecture implementation
- ✅ Production-ready code

---

**Last Updated:** June 29, 2024  
**Status:** ✅ Production Ready
