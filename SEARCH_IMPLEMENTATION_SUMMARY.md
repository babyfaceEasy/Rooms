# Search Endpoints Implementation Summary

## Overview

Added comprehensive fuzzy search functionality to admin endpoints for users, admins, and audit logs. All search endpoints support pagination with total counts and are case-insensitive.

## Implementation Details

### Search Behavior
- **Type**: Fuzzy/Partial matching using MongoDB regex with `$regex` and case-insensitive flag (`$options: "i"`)
- **Case-Sensitive**: No (all searches are case-insensitive)
- **Pagination**: Yes (all results are paginated with limit/offset and total count)
- **Response Format**: Same as list endpoints - `PaginatedResponse[T]` with `data`, `total`, `limit`, `offset`

## New Endpoints

### 1. Search Admins
**Endpoint**: `GET /api/v1/admin/admins/search?search=query&limit=10&offset=0`

**Permission**: `manage_admins` (super_admin only)

**Searchable Fields**:
- `email` (case-insensitive)
- `name` (case-insensitive)

**Response Example**:
```json
{
  "data": [
    {
      "id": "507f1f77bcf86cd799439011",
      "email": "john.admin@example.com",
      "name": "John Admin",
      "role": "admin",
      "is_active": true
    }
  ],
  "total": 2,
  "limit": 10,
  "offset": 0
}
```

### 2. Search Customers/Users
**Endpoint**: `GET /api/v1/admin/customers/search?search=query&limit=10&offset=0`

**Searchable Fields**:
- `email` (case-insensitive)
- `name` (case-insensitive)

**Response Example**:
```json
{
  "data": [
    {
      "id": "507f191e810c19729de860ea",
      "email": "john@example.com",
      "name": "John Doe"
    }
  ],
  "total": 2,
  "limit": 10,
  "offset": 0
}
```

### 3. Search Logs
**Endpoint**: `GET /api/v1/admin/logs/search?search=query&limit=50&offset=0`

**Permission**: `view_logs`

**Searchable Fields**:
- `admin_email` (case-insensitive)
- `action` (case-insensitive)

**Response Example**:
```json
{
  "data": [
    {
      "id": "507f191e810c19729de860fa",
      "admin_id": "507f191e810c19729de860fb",
      "admin_email": "admin@example.com",
      "action": "USER_SUSPEND",
      "resource_type": "USER",
      "resource_id": "507f191e810c19729de860fc",
      "changes": {
        "reason": "Violated community guidelines"
      },
      "status": "SUCCESS",
      "created_at": "2024-10-31T10:30:00Z"
    }
  ],
  "total": 342,
  "limit": 50,
  "offset": 0
}
```

## Files Modified

### Repository Layer (3 files)

#### 1. `internal/repository/admin_repository.go`
- Added `SearchAdmins(ctx, query, limit, offset) ([]*domain.Admin, error)`
- Added `CountSearchAdmins(ctx, query) (int, error)`

#### 2. `internal/repository/mongo_admin_repository.go`
- Implemented `SearchAdmins()` with regex filtering
- Implemented `CountSearchAdmins()` for total count

#### 3. `internal/repository/user_repository.go`
- Added `SearchUsers(ctx, query, limit, offset) ([]*domain.User, error)`
- Added `CountSearchUsers(ctx, query) (int, error)`

#### 4. `internal/repository/mongo_user_repository.go`
- Implemented `SearchUsers()` with regex filtering
- Implemented `CountSearchUsers()` for total count

#### 5. `internal/repository/admin_repository.go` (AdminLogRepository)
- Added `SearchLogs(ctx, query, limit, offset) ([]*domain.AdminLog, error)`
- Added `CountSearchLogs(ctx, query) (int, error)`

#### 6. `internal/repository/mongo_admin_log_repository.go`
- Implemented `SearchLogs()` with regex filtering
- Implemented `CountSearchLogs()` for total count

### Service Layer (3 files)

#### 1. `internal/service/admin_user_service.go`
- Added `SearchAdmins()` interface method
- Implemented `SearchAdmins()` calling repo search + count

#### 2. `internal/service/admin_user_mgmt_service.go`
- Added `SearchUsers()` interface method
- Implemented `SearchUsers()` calling repo search + count

#### 3. `internal/service/admin_log_service.go`
- Added `SearchLogs()` interface method
- Implemented `SearchLogs()` calling repo search + count

### Handler Layer (3 files)

#### 1. `internal/handler/admin_user_handler.go`
- Added `SearchAdmins()` handler
- Validates `search` query parameter (required)
- Handles pagination parameters (limit, offset)
- Removes sensitive fields from response

#### 2. `internal/handler/admin_customer_handler.go`
- Added `SearchUsers()` handler
- Validates `search` query parameter (required)
- Handles pagination parameters (limit, offset)
- Removes sensitive fields from response

#### 3. `internal/handler/admin_logs_handler.go`
- Added `SearchLogs()` handler
- Validates `search` query parameter (required)
- Handles pagination parameters (limit, offset)

### API Layer (1 file)

#### `internal/api/routes.go`
- Added `GET /api/v1/admin/admins/search` route
- Added `GET /api/v1/admin/customers/search` route
- Added `GET /api/v1/admin/logs/search` route

### Documentation (1 file)

#### `docs/ADMIN_API.md`
- Added comprehensive search endpoint documentation
- Included query parameters and response examples
- Added common search terms for logs

## Route Structure

```
Protected Routes (/api/v1/admin/)
├── Admin Management (/admins)
│   ├── GET /:id - Get admin by ID
│   ├── GET "" - List all admins (paginated)
│   ├── GET /search - Search admins (NEW)
│   ├── POST "" - Create admin
│   ├── PATCH /:id - Update admin
│   └── DELETE /:id - Delete admin
│
├── Customer Management (/customers)
│   ├── GET /:id - Get customer by ID
│   ├── GET "" - List all customers (paginated)
│   ├── GET /search - Search customers (NEW)
│   ├── POST /:id/suspend - Suspend customer
│   ├── DELETE /:id/suspend - Unsuspend customer
│   └── DELETE /:id - Delete customer
│
└── Logs (/logs)
    ├── GET "" - Get all logs (paginated)
    ├── GET /search - Search logs (NEW)
    └── GET /admin/:admin_id - Get logs for specific admin (paginated)
```

## Key Features

### 1. Fuzzy Search
- Uses MongoDB regex with case-insensitive matching
- Supports partial matches (e.g., searching "john" finds "John Doe", "johnny.admin@example.com")
- Example: `?search=joh` will match "John", "john@...", "Johnny", etc.

### 2. Pagination with Total Counts
- All search results include `total` field for complete pagination calculation
- Frontend can determine: `Math.ceil(total / limit)` = total pages
- Consistent response format with list endpoints

### 3. Permission-Based Access
- Admin search requires `manage_admins` permission (super_admin only)
- Customer search is accessible to any authenticated admin
- Log search requires `view_logs` permission

### 4. Database Performance
- Each search makes 2 queries: one for data (with skip/limit), one for total count
- Uses indexed fields (email, name) for optimal performance
- Soft delete filters respected (`deleted_at: nil`)

## Query Parameters

### Common to All Search Endpoints
- `search` (required): The search query term
  - Fuzzy/partial matching
  - Case-insensitive
  - Applied to email and name fields
  
- `limit` (optional, default: 10 for admins/customers, 50 for logs)
  - Maximum results per page
  - Must be > 0
  - Logs have max limit of 100
  
- `offset` (optional, default: 0)
  - Starting position in result set
  - Must be >= 0

## Usage Examples

### Search for Admin
```bash
curl -X GET "http://localhost:8080/api/v1/admin/admins/search?search=john&limit=10&offset=0" \
  -H "Authorization: Bearer <token>"
```

### Search for Customer
```bash
curl -X GET "http://localhost:8080/api/v1/admin/customers/search?search=doe&limit=20&offset=0" \
  -H "Authorization: Bearer <token>"
```

### Search Logs
```bash
curl -X GET "http://localhost:8080/api/v1/admin/logs/search?search=USER_SUSPEND&limit=50&offset=0" \
  -H "Authorization: Bearer <token>"
```

## Frontend Integration

### React/TypeScript Example
```typescript
async function searchAdmins(query: string, limit: number = 10, offset: number = 0) {
  const token = localStorage.getItem('adminAccessToken');
  
  const response = await fetch(
    `/api/v1/admin/admins/search?search=${encodeURIComponent(query)}&limit=${limit}&offset=${offset}`,
    {
      headers: { 'Authorization': `Bearer ${token}` }
    }
  );
  
  if (!response.ok) throw new Error('Search failed');
  
  const result = await response.json();
  
  return {
    admins: result.data,
    total: result.total,
    totalPages: Math.ceil(result.total / result.limit),
    currentPage: Math.floor(result.offset / result.limit) + 1
  };
}
```

## Soft Delete Handling

All search functions respect soft deletes:
- Filter: `deleted_at: null`
- Excludes deleted admins and users
- Logs are not soft-deleted (all logged actions are searchable)

## Error Responses

### Missing Search Parameter (400 Bad Request)
```json
{
  "error": "search query parameter is required"
}
```

### Internal Server Error (500)
```json
{
  "error": "failed to search users|admins|logs"
}
```

## Testing Checklist

- [ ] Search admins by email
- [ ] Search admins by name
- [ ] Search admins with partial match (e.g., "joh" finds "John")
- [ ] Search admins case-insensitive ("JOHN" finds "john")
- [ ] Verify pagination (limit, offset, total work correctly)
- [ ] Search customers by email
- [ ] Search customers by name
- [ ] Search logs by admin email
- [ ] Search logs by action
- [ ] Verify permission checks (manage_admins, view_logs)
- [ ] Verify sensitive fields are removed (passwords, tokens)
- [ ] Verify soft deleted items are excluded
- [ ] Test error cases (missing search param, invalid limit/offset)

## Performance Notes

- Search uses indexed fields (email, name) for O(log n) performance
- Two queries per request (data + count) - acceptable for current collection sizes
- If needed, can be optimized later with MongoDB aggregation `$facet` pipeline

## Breaking Changes

None. These are new endpoints - existing list endpoints remain unchanged.

## Deployment Notes

- No database migrations needed
- No index changes needed (email/name indexes already exist)
- Search functionality available immediately after deployment
- Frontend teams should update to use search endpoints for better UX
