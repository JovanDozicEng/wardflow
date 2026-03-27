# WardFlow — Copilot Instructions

WardFlow is an inpatient/ED care-coordination system designed to reduce missed actions during handoffs, improve real-time operational visibility, and provide accountable tasking with compliance-ready audit trails.

## Architecture Overview

**Tech Stack:**
- **Backend:** Go 1.21+ (service-oriented, REST API)
- **Frontend:** React 18+ (SPA, role-aware UI)
- **Database:** PostgreSQL (relational + history)

**Key Architectural Principles:**
- **Audit-first design:** No destructive overwrites. Every state mutation must record `who`, `when`, `why`, and `source`.
- **Timeline as first-class citizen:** Entities like care team assignments, flow state transitions, task ownership, and bed status changes preserve full history as immutable events.
- **RBAC enforcement at the API layer** — never frontend-only. Users see only data within their authorized unit/department boundaries.
- **Idempotent API design** — critical for integration resilience; clients should supply idempotency keys.

**Scope (MVP):**
1. Care team assignment per encounter (with structured handoff notes)
2. Patient flow tracking (state timeline with audit)
3. Clinical task board (SLA tracking, escalation)
4. Inter-department consult requests (acceptance/decline workflows)
5. Bed management (status & capability tracking)
6. Transport requests + tracking
7. Discharge planning checklist (structured, auditable)
8. Exception workflows (draft → finalized → corrected)
9. Daily huddle dashboard (materialized, near-real-time)
10. Quality/safety incident logging (immutable until correction workflow)

**Out of Scope:** EHR order entry, billing/claims, full clinical documentation, medication administration.

## Core Data Model

### Entity Hierarchy
- **Encounter** — central anchor; links patient, unit, department, flow state, care team, tasks, consults, bed placement, transport, discharge
- **CareTeamAssignment** (historical) — encounters have many; each has userId, roleType, effective dates, handoff notes
- **FlowStateTransition** (immutable event) — encounters have many; each records state change with timestamp, actor, reason
- **Task** (scoped) — attached to encounter, patient, or unit; statuses: open, in_progress, completed, cancelled, escalated
- **ConsultRequest** (request/response pair) — from encounter to target service; statuses: pending, accepted, declined, completed, redirected, cancelled
- **Bed** + **BedStatusEvent** — unit beds with operational status (available, occupied, blocked, maintenance) and capability tags
- **TransportRequest** + **TransportChangeEvent** — encounter-scoped; audit all field changes
- **DischargeChecklist** + **DischargeChecklistItem** — structured checklist per discharge type; items have required/waived states
- **ExceptionEvent** + **Incident** — immutable until correction workflow; preserve historical corrections

### Core Fields to Always Include
- **Timestamps:** `createdAt`, `updatedAt`, sometimes explicit `startsAt`/`endsAt` for validity windows
- **Audit context:** `createdBy`, `updatedBy` (user IDs); audit log entries should store `reason` and `source` (e.g., "user_action" vs "system_event")
- **Soft delete readiness:** consider `deletedAt` nullable timestamp for logical deletion; do not physically remove audit events

## Build, Test & Lint Commands

> Project scaffolding in progress. Add these commands once backend/frontend are initialized.

**Expected patterns:**
- **Backend (Go):** `go build ./...`, `go test ./...`, `go test -v ./...` (single test), `golangci-lint run`, `go fmt ./...`
- **Frontend (React):** `npm run build`, `npm test`, `npm test -- --testNamePattern="<test_name>"`, `npm run lint`
- **Docker:** All services should run via `docker-compose up`

## Key Conventions

### Backend (Go)
- **Package structure:** Organize by domain (e.g., `internal/care`, `internal/task`, `internal/bed`); use `internal/` for private packages.
- **Error handling:** Wrap errors with context; never ignore errors. Use `fmt.Errorf("verb: %w", err)` pattern.
- **Request/response models:** Define in separate `dto` or `models` subpackage; use struct tags for JSON marshaling and validation.
- **Middleware:** Implement auth, RBAC, and logging as early middleware; attach user context to request.
- **Testing:** Prefer table-driven tests; mock database and external services; use `httptest` for handler tests.
- **Database:** Use GORM or sqlx; prepare parametrized queries to prevent injection; always scan error-safely.

### Frontend (React)
- **Component structure:** Functional components + hooks; co-locate logic with UI.
- **State management:** Use Context API for auth/user/permissions; consider Zustand or Redux Toolkit for complex cross-component state.
- **API integration:** Create service layer (e.g., `services/api.ts`) to centralize calls; include error handling and retry logic.
- **Role-aware rendering:** Access user context to conditionally render features; never trust frontend-only access control.
- **Form handling:** Use controlled components; validate on change and submit; show clear error messages tied to field-level errors.
- **Testing:** Jest + React Testing Library; test user interactions, not implementation details.

### Database & Schema
- **Migrations:** Version all schema changes; include both `up` and `down` scripts; consider SQL-based or ORM-based approach (e.g., GORM AutoMigrations or Flyway).
- **Audit logging:** Create `audit_log` table with `(id, table_name, record_id, action, user_id, reason, old_values, new_values, created_at)`.
- **Immutable events:** Design tables like `flow_state_transition`, `task_assignment_event`, `bed_status_event` with no update — only insert.
- **Constraints:** Enforce business rules at schema level (e.g., FK for user IDs, unique constraints for active assignments per encounter).

### API Design
- **Endpoint patterns:**
  - `POST /encounters/{encounterId}/care-team` — assign a role to the encounter
  - `GET /encounters/{encounterId}/care-team?activeOnly=true` — list current care team
  - `GET /encounters/{encounterId}/flow-timeline` — list flow state transitions in order
  - `POST /tasks` — create task (scoped to encounter, patient, or unit)
  - `PATCH /tasks/{taskId}` — update status, owner, or close with metadata
  - `POST /consults` — create consult request
  - `PATCH /consults/{consultId}/accept|decline` — formal workflow actions
- **Error responses:** Use standard format: `{ "error": "code", "message": "details", "fieldErrors": {...} }`
- **Pagination:** Support `limit` and `offset` query params; return `{ "data": [...], "total": N, "limit": L, "offset": O }`
- **OpenAPI:** All endpoints must be documented in OpenAPI spec (Swagger); regenerate after code changes.

### RBAC & Permissions
- **Roles:** Nurse, Provider, Charge Nurse, Operations, Consult Service, Transport, Quality/Safety, Admin
- **Scope:** Users typically see data within assigned unit/department; privileged actions (e.g., admin override, reassign critical roles) require explicit audit trail
- **Enforcement points:**
  - Extract user + roles from token (JWT or session); attach to request context
  - Check permissions before data access and state mutations
  - Log privileged actions with reason

### Naming Conventions
- **Table names:** snake_case, plural or singular (choose one, be consistent); e.g., `care_team_assignment`, `flow_state_transition`
- **Column names:** snake_case; foreign keys as `{entity_id}`; boolean flags as `is_*` (e.g., `is_active`)
- **API fields:** camelCase in JSON (use Go struct tags to map)
- **Go functions:** camelCase; receiver methods prefixed with receiver type abbreviated (e.g., `(e *Encounter) IsActive()`)

## Documentation References

- **Full task description & business rules:** `docs/task-description.md`
- **Technical requirements & spec pack:** `docs/req-and-spec-pack.md`
- **Backend agent guidance:** `.copilot/.github/agents/backend-agent.md`
- **Frontend agent guidance:** `.copilot/.github/agents/frontend-agent.md`

## When to Reach Out

- Architectural decisions affecting multiple modules
- Schema changes impacting audit/history design
- RBAC policy changes or new roles
- Cross-service integration patterns
- Performance/scaling concerns

Always check `/docs` when making recommendations about architecture or conventions.
