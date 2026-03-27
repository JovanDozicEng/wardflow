# Phase 0 (Foundation) Implementation Summary

## Completed ✅

All Phase 0 requirements have been successfully implemented and the code compiles without errors.

### 1. Updated `internal/models/response.go`
- ✅ Replaced flat ErrorResponse with spec-compliant error format
- ✅ Added `APIError` struct with Code, Message, Details, CorrelationID
- ✅ Added `FieldError` struct for field-level validation errors
- ✅ Added `ErrorEnvelope` wrapper struct
- ✅ Updated `PaginatedResponse` with optional Cursor field for cursor-based pagination
- ✅ Kept `HealthResponse` unchanged

### 2. Created `internal/httputil/response.go`
- ✅ Centralized response helpers for all handlers and middleware
- ✅ `RespondJSON()` - writes JSON response with proper Content-Type header
- ✅ `RespondError()` - writes spec-compliant error with correlation ID from context
- ✅ `RespondValidationError()` - writes validation errors with field details
- ✅ `CorrelationIDFromContext()` - extracts correlation ID using typed context key

### 3. Created `internal/middleware/correlation.go`
- ✅ Generates UUID v4 correlation ID per request
- ✅ Stores correlation ID in context using typed key (`httputil.CorrelationIDKey`)
- ✅ Sets `X-Correlation-ID` response header
- ✅ Uses `crypto/rand` for secure UUID generation

### 4. Updated `internal/handler/auth.go`
- ✅ Replaced local `respondJSON`/`respondError` with `httputil` functions
- ✅ Updated all error responses to use proper error codes:
  - `METHOD_NOT_ALLOWED` for 405 responses
  - `VALIDATION_ERROR` for 400 validation errors
  - `CONFLICT` for 409 email exists
  - `UNAUTHORIZED` for 401 invalid credentials
  - `FORBIDDEN` for 403 inactive users
  - `NOT_FOUND` for 404 user not found
  - `INTERNAL_ERROR` for 500 server errors

### 5. Updated `internal/middleware/auth.go`
- ✅ Replaced local `respondError` with `httputil.RespondError`
- ✅ Updated all middleware functions to use spec-compliant error responses
- ✅ Removed old local helper function

### 6. Created `internal/models/audit_log.go`
- ✅ GORM model with UUID primary key (PostgreSQL `gen_random_uuid()`)
- ✅ All required fields: EntityType, EntityID, Action, At, ByUserID
- ✅ Optional fields: IP, UserAgent, Reason, CorrelationID
- ✅ Source field with default 'user_action'
- ✅ JSONB fields for BeforeJSON and AfterJSON
- ✅ Proper indexes on key fields
- ✅ Custom table name: "audit_log"

### 7. Created `internal/audit/log.go`
- ✅ `Entry` struct for audit log data
- ✅ `Log()` function that writes audit entries non-fatally
- ✅ Extracts IP from `r.RemoteAddr` (strips port)
- ✅ Extracts UserAgent from request header
- ✅ Extracts CorrelationID from context
- ✅ Marshals Before/After to JSON strings
- ✅ Defaults Source to "user_action" if not specified
- ✅ Logs warnings on failures instead of failing requests

### 8. Created `internal/encounter/` package

#### `models.go`
- ✅ `EncounterStatus` type with constants: active, discharged, cancelled
- ✅ `Encounter` GORM model with all required fields
- ✅ UUID primary key with PostgreSQL default
- ✅ Proper indexes on PatientID, UnitID, DepartmentID, Status
- ✅ Audit fields: CreatedBy, UpdatedBy, CreatedAt, UpdatedAt
- ✅ `CreateEncounterRequest` DTO with optional StartedAt
- ✅ `UpdateEncounterRequest` DTO with optional fields
- ✅ `ListEncountersFilter` for query parameters

#### `repository.go`
- ✅ `Repository` struct wrapping database
- ✅ `Create()` - creates new encounter
- ✅ `GetByID()` - retrieves by ID, returns `ErrNotFound` if not found
- ✅ `List()` - filters by UnitID, DepartmentID, Status with pagination
- ✅ `Update()` - updates encounter, returns `ErrNotFound` if not found
- ✅ Returns count for pagination

#### `service.go`
- ✅ `Service` struct with business logic
- ✅ `Create()` - validates required fields, defaults StartedAt to now
- ✅ `GetByID()` - delegates to repository
- ✅ `List()` - delegates to repository
- ✅ `Update()` - validates status transitions (cannot re-activate discharged/cancelled)
- ✅ Updates UpdatedBy field

#### `handler.go`
- ✅ `List()` - GET handler with query param parsing
- ✅ `Create()` - POST handler with unit access check
- ✅ `GetByID()` - GET handler using Go 1.22+ `PathValue("id")`
- ✅ `Update()` - PATCH handler with unit access check
- ✅ All handlers enforce unit access (unless admin)
- ✅ Writes audit logs on Create and Update
- ✅ Uses `httputil` for all responses

#### `routes.go`
- ✅ `RegisterRoutes()` function
- ✅ All routes under `/api/v1` prefix
- ✅ Uses Go 1.22+ method+path patterns:
  - `GET /api/v1/encounters`
  - `POST /api/v1/encounters`
  - `GET /api/v1/encounters/{id}`
  - `PATCH /api/v1/encounters/{id}`
- ✅ All routes wrapped with `AuthMiddleware`

### 9. Updated `internal/router/router.go`
- ✅ Added `encounter.RegisterRoutes()` call
- ✅ Added `readyz` endpoint: `GET /readyz` returns `{"status":"ready"}`
- ✅ Wrapped mux with `CorrelationID` middleware (inside CORS)
- ✅ Middleware order: CORS → CorrelationID → Routes → Auth (per route)

### 10. Updated `cmd/api/main.go`
- ✅ Imported `internal/encounter` package
- ✅ Updated `runMigrations()` to include:
  - `&models.User{}`
  - `&models.AuditLog{}`
  - `&encounter.Encounter{}`

## Build Status ✅
```bash
$ go build ./...
# Success - no errors

$ go build -v ./cmd/api
# Success - binary created
```

## API Endpoints

### Auth (existing)
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/logout` (protected)
- `GET /auth/me` (protected)
- `POST /auth/change-password` (protected)

### Encounters (new)
- `GET /api/v1/encounters` (protected)
  - Query params: `unitId`, `departmentId`, `status`, `limit`, `offset`
- `POST /api/v1/encounters` (protected)
- `GET /api/v1/encounters/{id}` (protected)
- `PATCH /api/v1/encounters/{id}` (protected)

### Health/Readiness
- `GET /health` - database health check
- `GET /readyz` - readiness probe

## Error Response Format

All errors now follow the spec:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      {
        "field": "email",
        "issue": "email is required"
      }
    ],
    "correlationId": "a1b2c3d4-e5f6-4789-a0b1-c2d3e4f5g6h7"
  }
}
```

## Audit Logging

All Create and Update operations on encounters are automatically logged to the `audit_log` table with:
- Before/After state (JSON)
- User ID, IP, User-Agent
- Correlation ID for request tracing
- Timestamp and action type

## Tech Stack Confirmation
- ✅ Go 1.25
- ✅ GORM for ORM
- ✅ PostgreSQL database
- ✅ stdlib `net/http` (no external router)
- ✅ Module: `github.com/wardflow/backend`

## Next Steps

Phase 0 foundation is complete. Ready for:
- Phase 1: Patient management
- Phase 2: Task system
- Phase 3: Real-time features
- Phase 4: Analytics
