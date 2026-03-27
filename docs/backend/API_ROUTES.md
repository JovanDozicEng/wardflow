# WardFlow Backend API Routes

## Base URL
- Development: `http://localhost:8080`
- Production: TBD

## Authentication

All protected endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

## Public Endpoints

### Health & Readiness
```http
GET /health
```
Returns database health status.

**Response** (200 OK):
```json
{
  "status": "ok",
  "database": "connected"
}
```

```http
GET /readyz
```
Returns readiness status for Kubernetes probes.

**Response** (200 OK):
```json
{
  "status": "ready"
}
```

### Authentication

```http
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe",
  "role": "nurse"
}
```

**Response** (201 Created):
```json
{
  "message": "registration successful",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "nurse"
  }
}
```

```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response** (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "nurse"
  }
}
```

## Protected Endpoints

### Authentication (Protected)

```http
POST /auth/logout
Authorization: Bearer <token>
```

**Response** (200 OK):
```json
{
  "message": "logout successful"
}
```

```http
GET /auth/me
Authorization: Bearer <token>
```

**Response** (200 OK):
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "nurse",
  "unitIds": ["unit-1", "unit-2"],
  "isActive": true
}
```

```http
POST /auth/change-password
Authorization: Bearer <token>
Content-Type: application/json

{
  "oldPassword": "password123",
  "newPassword": "newpassword456"
}
```

**Response** (200 OK):
```json
{
  "message": "password changed successfully"
}
```

### Encounters

```http
GET /api/v1/encounters?unitId=unit-1&status=active&limit=20&offset=0
Authorization: Bearer <token>
```

**Query Parameters:**
- `unitId` (optional) - Filter by unit ID
- `departmentId` (optional) - Filter by department ID
- `status` (optional) - Filter by status (active, discharged, cancelled)
- `limit` (optional, default: 20) - Number of results per page
- `offset` (optional, default: 0) - Pagination offset

**Response** (200 OK):
```json
{
  "data": [
    {
      "id": "encounter-uuid",
      "patientId": "patient-uuid",
      "unitId": "unit-1",
      "departmentId": "dept-1",
      "status": "active",
      "startedAt": "2024-03-27T10:00:00Z",
      "endedAt": null,
      "createdBy": "user-uuid",
      "updatedBy": "user-uuid",
      "createdAt": "2024-03-27T10:00:00Z",
      "updatedAt": "2024-03-27T10:00:00Z"
    }
  ],
  "total": 42,
  "limit": 20,
  "offset": 0
}
```

```http
POST /api/v1/encounters
Authorization: Bearer <token>
Content-Type: application/json

{
  "patientId": "patient-uuid",
  "unitId": "unit-1",
  "departmentId": "dept-1",
  "startedAt": "2024-03-27T10:00:00Z"  // optional, defaults to now
}
```

**Response** (201 Created):
```json
{
  "id": "encounter-uuid",
  "patientId": "patient-uuid",
  "unitId": "unit-1",
  "departmentId": "dept-1",
  "status": "active",
  "startedAt": "2024-03-27T10:00:00Z",
  "endedAt": null,
  "createdBy": "user-uuid",
  "updatedBy": "user-uuid",
  "createdAt": "2024-03-27T10:00:00Z",
  "updatedAt": "2024-03-27T10:00:00Z"
}
```

```http
GET /api/v1/encounters/{id}
Authorization: Bearer <token>
```

**Response** (200 OK):
```json
{
  "id": "encounter-uuid",
  "patientId": "patient-uuid",
  "unitId": "unit-1",
  "departmentId": "dept-1",
  "status": "active",
  "startedAt": "2024-03-27T10:00:00Z",
  "endedAt": null,
  "createdBy": "user-uuid",
  "updatedBy": "user-uuid",
  "createdAt": "2024-03-27T10:00:00Z",
  "updatedAt": "2024-03-27T10:00:00Z"
}
```

```http
PATCH /api/v1/encounters/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "status": "discharged",
  "endedAt": "2024-03-27T18:00:00Z",
  "unitId": "unit-2"  // optional, for transfers
}
```

**Response** (200 OK):
```json
{
  "id": "encounter-uuid",
  "patientId": "patient-uuid",
  "unitId": "unit-2",
  "departmentId": "dept-1",
  "status": "discharged",
  "startedAt": "2024-03-27T10:00:00Z",
  "endedAt": "2024-03-27T18:00:00Z",
  "createdBy": "user-uuid",
  "updatedBy": "user-uuid",
  "createdAt": "2024-03-27T10:00:00Z",
  "updatedAt": "2024-03-27T18:05:00Z"
}
```

## Error Responses

All errors follow this format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": [
      {
        "field": "fieldName",
        "issue": "specific issue"
      }
    ],
    "correlationId": "a1b2c3d4-e5f6-4789-a0b1-c2d3e4f5g6h7"
  }
}
```

### Error Codes

- `METHOD_NOT_ALLOWED` - HTTP method not allowed (405)
- `VALIDATION_ERROR` - Request validation failed (400)
- `CONFLICT` - Resource conflict (409)
- `UNAUTHORIZED` - Authentication required or invalid (401)
- `FORBIDDEN` - Insufficient permissions (403)
- `NOT_FOUND` - Resource not found (404)
- `INTERNAL_ERROR` - Server error (500)

## Response Headers

All responses include:
- `Content-Type: application/json`
- `X-Correlation-ID: <uuid>` - For request tracing

## CORS

Allowed origins (development):
- http://localhost:5173
- http://localhost:5174
- http://localhost:5175
- http://localhost:5176
- http://localhost:3000

Allowed methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
Allowed headers: Authorization, Content-Type
Exposed headers: X-Correlation-ID

## Rate Limiting

Not yet implemented (planned for Phase 3).

## Versioning

API version is included in the path for new endpoints:
- `/api/v1/...` - Current version
- Legacy endpoints (auth) remain at root level for backward compatibility

## Pagination

List endpoints support offset-based pagination:
- `limit` - Number of results (default: 20)
- `offset` - Skip N results (default: 0)
- Response includes `total` count for calculating pages

Cursor-based pagination support is planned for Phase 3.
