# Copilot Instructions for temp_backend

## Build, Test & Run Commands

### Prerequisites
- Go 1.25+
- Docker & Docker Compose (for MongoDB and MinIO)

### Starting Local Development Environment
```bash
docker-compose up -d
```

### Running the Application
```bash
go run cmd/api/main.go
```

The server starts on the port defined in environment variables (default: 8080). The `.env.example` file contains the required configuration. Copy it to `.env` and adjust values as needed.

### Building for Production
```bash
go build -o temp_backend cmd/api/main.go
```

### Testing
Currently no test framework is set up. When adding tests, follow Go conventions with `*_test.go` files colocated with the code being tested.

## Architecture Overview

This project follows **Strict Clean/Hexagonal Architecture**:

- **`cmd/api/main.go`** — Application entrypoint. Initializes config, database, S3/MinIO clients, wires dependencies, and starts the server with graceful shutdown.

- **`config/`** — Environment configuration loading. Config struct is populated from environment variables on startup.

- **`internal/domain/`** — Domain entities and sentinel errors.
  - `item.go` — `Item` struct with JSON/BSON tags for MongoDB serialization.
  - `errors.go` — Domain-level sentinel errors (`ErrNotFound`, `ErrInvalidInput`, `ErrUnauthorized`, etc.). These are intentionally broad for upper layers to map to HTTP status codes.

- **`internal/repository/`** — Data access layer (interfaces and implementations).
  - `item_repository.go` — Interface for item metadata persistence.
  - `mongo_item_repository.go` — MongoDB implementation.
  - `s3_repository.go` — S3/MinIO implementation for file storage.

- **`internal/service/`** — Business logic layer.
  - `item_service.go` — Orchestrates item metadata and file storage. Uses domain errors for error propagation.

- **`internal/handler/`** — HTTP request handlers (Fiber).
  - `item_handler.go` — Exposes CRUD endpoints and file download. Wraps service calls and propagates errors to middleware.

- **`internal/api/`** — Server setup and routing.
  - `server.go` — Fiber app initialization, middleware registration, graceful shutdown.
  - `routes.go` — Route definitions. API endpoints are versioned under `/api/v1`.

- **`internal/middleware/`** — HTTP middleware.
  - `error.go` — Global error handler middleware that converts domain errors to HTTP status codes.

- **`pkg/`** — Shared utilities and external integrations.
  - `mongodb/` — MongoDB client initialization.
  - `s3/` — S3/MinIO client initialization and utilities.

## Key Conventions

### Error Handling
- Domain layer defines **sentinel errors** in `internal/domain/errors.go` (e.g., `ErrNotFound`, `ErrInvalidInput`).
- Service layer wraps errors with context using `fmt.Errorf("context: %w", err)` to preserve the sentinel error type.
- Handler layer passes errors to the error middleware, which maps them to HTTP status codes based on the underlying error type.
- **Always check the error type** when handling errors that cross layers. Example:
  ```go
  if errors.Is(err, domain.ErrNotFound) {
      return c.Status(fiber.StatusNotFound).JSON(...)
  }
  ```

### Service Interfaces
- Services are defined as interfaces (e.g., `ItemService`) to enable dependency injection and testing.
- The concrete implementation is a private struct (e.g., `itemService`).
- Constructor functions return the interface type, not the concrete struct. Example:
  ```go
  func NewItemService(items repository.ItemRepository, storage repository.ObjectStorage) ItemService {
      return &itemService{...}
  }
  ```

### Repository Pattern
- Repositories provide interfaces for data access (e.g., `ItemRepository`, `ObjectStorage`).
- Implementations are concrete structs that handle MongoDB, S3, or other storage backends.
- All repository methods take `context.Context` as the first parameter for cancellation and timeouts.

### File Uploads & Storage
- Multipart file uploads are handled in handlers and passed to the service layer.
- The service layer manages file attachment logic, key generation, and storage.
- Object keys follow the pattern: `items/{itemID}/{randomUID}-{sanitizedFilename}{ext}`.
- Filenames are sanitized to allow only alphanumeric characters, hyphens, underscores, and dots.
- Content-Type is inferred from file extension if not provided in the upload request.

### Dependency Injection
- All dependencies are injected at initialization time (see `main.go`).
- The service layer receives repository interfaces, not concrete implementations.
- Handlers receive the service interface.
- No global state or singletons.

### Config Management
- Configuration is loaded from environment variables on startup using the `config` package.
- All required config is validated and available in the `Config` struct before the application runs.
- The `.env.example` file documents all required environment variables and their defaults.

### Logging
- Use `log/slog` for structured logging (initialized in `main.go` based on the `LOG_LEVEL` env var).
- Log level can be set to: `debug`, `info` (default), `warn`, or `error`.

### No Code Placeholders
- Never use `// TODO` comments or incomplete code.
- All implementations must be complete and functional.
- If a feature is incomplete, document it in a separate issue or plan file, not in the code.

## Technology Stack

- **Language**: Go 1.25+
- **Web Framework**: Fiber v2
- **Database**: MongoDB (with official driver)
- **Object Storage**: MinIO/S3 (with AWS SDK v2)
- **Logging**: log/slog (standard library)

## Development Workflow

1. **Local Setup**: Copy `.env.example` to `.env` and run `docker-compose up -d` to start MongoDB and MinIO.
2. **Building**: Use `go run cmd/api/main.go` during development.
3. **Code Organization**: Keep code in the appropriate layer (domain, repository, service, handler).
4. **Error Propagation**: Define new domain errors in `internal/domain/errors.go` for new error conditions.
5. **Testing**: When writing tests, use the standard `*_test.go` naming convention and colocate with the code under test.
