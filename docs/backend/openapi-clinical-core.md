# OpenAPI Documentation - Clinical Core Modules (Developer A)

## Overview

This document describes the OpenAPI/Swagger specification for the 4 clinical-core modules implemented by Developer A.

## Integration Instructions

The complete OpenAPI specification additions are in:
```
backend/openapi-clinical-core-addition.yaml
```

To integrate into the main `openapi.yaml` file:

1. **Copy Path Definitions**: Merge the paths section into the existing paths
2. **Copy Schema Definitions**: Add all schemas to `components.schemas`
3. **Verify References**: Ensure all `$ref` pointers resolve correctly

## API Endpoints Summary

### Module 1: Care Team Assignment (4 endpoints)

| Method | Path | Summary |
|--------|------|---------|
| GET | `/api/v1/encounters/{encounterId}/care-team` | Get current care team |
| POST | `/api/v1/encounters/{encounterId}/care-team/assignments` | Assign role to care team |
| POST | `/api/v1/care-team/assignments/{assignmentId}/transfer` | Transfer role with handoff |
| GET | `/api/v1/encounters/{encounterId}/handoffs` | List handoff notes |

**Key Features:**
- Active/historical assignment filtering
- Optional user details population
- Mandatory handoff notes for critical roles
- Pagination support

### Module 2: Flow Tracking (4 endpoints)

| Method | Path | Summary |
|--------|------|---------|
| GET | `/api/v1/encounters/{encounterId}/flow` | Get flow state timeline |
| GET | `/api/v1/encounters/{encounterId}/flow/current` | Get current state only |
| POST | `/api/v1/encounters/{encounterId}/flow/transitions` | Record state transition |
| POST | `/api/v1/encounters/{encounterId}/flow/override` | Override transition (privileged) |

**Key Features:**
- State machine validation
- Actor details population
- Override workflow with reason tracking
- RBAC enforcement (admin/operations only for override)

**State Machine:**
```
arrived → triage → provider_eval ⟶ diagnostics
                ↓              ↓
          discharge_ready ← admitted
                ↓
           discharged
```

### Module 3: Task Board (7 endpoints)

| Method | Path | Summary |
|--------|------|---------|
| GET | `/api/v1/tasks` | List tasks with filters |
| POST | `/api/v1/tasks` | Create task |
| GET | `/api/v1/tasks/{id}` | Get task details |
| PATCH | `/api/v1/tasks/{id}` | Update task |
| POST | `/api/v1/tasks/{id}/assign` | Assign/reassign task |
| POST | `/api/v1/tasks/{id}/complete` | Complete task |
| GET | `/api/v1/tasks/{id}/history` | Get assignment history |

**Key Features:**
- Multi-scope tasks (encounter/patient/unit)
- SLA tracking with overdue detection
- Complex filtering (status, priority, owner, overdue)
- RBAC-enforced assignment rules
- Owner details population option

**RBAC Rules:**
- **Assign**: Admin/charge nurse/operations can reassign any task; others can only assign unassigned tasks or take their own
- **Complete**: Admin/charge nurse can complete any; others only their assigned tasks

### Module 4: Dashboard (1 endpoint)

| Method | Path | Summary |
|--------|------|---------|
| GET | `/api/v1/dashboard/huddle` | Get aggregated huddle metrics |

**Key Features:**
- Real-time aggregated metrics
- Census and flow distribution
- Task metrics (overdue, high priority, unassigned)
- Risk indicators (triage >2hrs, missing care team, etc.)
- RBAC-filtered by unit/department

**Returned Metrics:**
- **Census**: Active encounters, expected discharges
- **Flow Distribution**: Count by state
- **Task Metrics**: Open, overdue, high priority, urgent, unassigned, completed today
- **Risk Indicators**: 5 key risk metrics
- **Drill-down Lists**: Overdue tasks, long stay patients, pending discharges

## Schema Definitions

### Care Team
- `RoleType` (enum): 8 clinical roles
- `CareTeamAssignment`: Historical assignment record
- `HandoffNote`: Structured handoff documentation
- Request/Response DTOs

### Flow Tracking
- `FlowState` (enum): 7 states
- `FlowStateTransition`: Immutable event record
- `ActorType` (enum): user/system
- Request/Response DTOs with override support

### Tasks
- `TaskStatus` (enum): 5 statuses
- `TaskPriority` (enum): 4 levels
- `ScopeType` (enum): encounter/patient/unit
- `Task`: Main task entity with SLA
- `TaskAssignmentEvent`: Immutable assignment history
- Request/Response DTOs

### Dashboard
- `HuddleMetrics`: Top-level aggregation
- `FlowDistribution`: State counts
- `RiskIndicators`: 5 risk metrics
- `TaskMetrics`: Task board overview
- `TaskSummary`, `EncounterSummary`: Drill-down items

## Authentication & Authorization

**All endpoints require authentication** via Bearer JWT token except system health endpoints.

**RBAC Enforcement:**
- Care team transfers: Any authenticated user (TODO: add role restrictions)
- Flow overrides: Admin and Operations roles only
- Task assignment: Admin/charge nurse/operations for any task; others limited
- Task completion: Admin/charge nurse for any; others only assigned tasks
- Dashboard: Filtered by user's authorized units/departments

## Error Responses

All endpoints return standardized error format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable summary",
    "details": [{"field": "...", "issue": "..."}],
    "correlationId": "..."
  }
}
```

**Common Error Codes:**
- `400`: `VALIDATION_ERROR`, `INVALID_REQUEST`, `TRANSITION_FAILED`, `ASSIGNMENT_FAILED`
- `401`: `UNAUTHORIZED`
- `403`: `FORBIDDEN`
- `404`: `NOT_FOUND`
- `500`: `INTERNAL_ERROR`

## Pagination

Endpoints that support pagination use:
- `limit` (query param): Max items (1-100, default 30)
- `offset` (query param): Skip items (default 0)

Response includes:
```json
{
  "data": [...],
  "total": 123,
  "limit": 30,
  "offset": 0
}
```

## Audit Trail

All state-changing operations are logged to `audit_log` table with:
- `who`: User ID
- `when`: Timestamp (UTC)
- `what`: Action (CREATE/UPDATE/DELETE/OVERRIDE)
- `why`: Reason (if provided)
- `before`/`after`: State snapshots

## Testing the API

### Using curl:

```bash
# Get JWT token
TOKEN=$(curl -s http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  | jq -r '.token')

# Get care team
curl http://localhost:8080/api/v1/encounters/{id}/care-team \
  -H "Authorization: Bearer $TOKEN"

# Get flow timeline
curl http://localhost:8080/api/v1/encounters/{id}/flow \
  -H "Authorization: Bearer $TOKEN"

# List tasks
curl "http://localhost:8080/api/v1/tasks?status=open&overdue=true" \
  -H "Authorization: Bearer $TOKEN"

# Get dashboard
curl "http://localhost:8080/api/v1/dashboard/huddle?unitId=ICU-1" \
  -H "Authorization: Bearer $TOKEN"
```

### Using Swagger UI:

1. Start backend: `go run ./cmd/api/main.go`
2. Open: http://localhost:8080/swagger (if Swagger UI is configured)
3. Authorize with JWT token
4. Test endpoints interactively

## Validation Rules

### Care Team
- `userId` and `roleType` are required for assignment
- Critical roles (primary_nurse, attending_provider) require handoff note on transfer

### Flow
- Transitions must follow state machine (unless override)
- Override requires `reason` field
- Only admin/operations can override

### Tasks
- `scopeType`, `scopeId`, and `title` are required
- `priority` defaults to medium if not specified
- `toOwnerId` can be null to unassign

### Dashboard
- Filters to user's authorized units/departments automatically
- Non-admin users must have unit/department restrictions configured

## Future Enhancements

- [ ] Add Swagger UI middleware
- [ ] Add request/response examples to schemas
- [ ] Add webhook specifications for real-time updates
- [ ] Add batch operation endpoints
- [ ] Add GraphQL alternative specification
- [ ] Add rate limiting documentation
- [ ] Add API versioning strategy

## References

- OpenAPI 3.1 Specification: https://spec.openapis.org/oas/v3.1.0
- Swagger Editor: https://editor.swagger.io/
- Main OpenAPI file: `backend/openapi.yaml`
- Additions file: `backend/openapi-clinical-core-addition.yaml`
