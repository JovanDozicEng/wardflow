# WardFlow — 3-Developer Implementation Plan

> **Purpose:** Splits the 10 MVP modules across 3 developers (Copilot agents) on independent feature branches with zero overlap. All module specs are defined in [`req-and-spec-pack.md`](./req-and-spec-pack.md).

---

## Current State

| Layer | Status |
|---|---|
| **Backend (Go)** | Auth handler, JWT middleware, basic router, DB connection — no domain modules yet |
| **Frontend (React/TS)** | Auth feature complete; `care-team` scaffolded (UI components exist, not wired to backend); shared UI primitives ready |
| **DB Schema** | Not yet defined for domain entities |

---

## Phase 0 — Shared Foundation (merge to `main` first)

**Branch:** `feature/foundation`

Everything that would block parallel work if missing:

- `encounters` table migration + Encounter model + basic CRUD handler
  - `GET /api/v1/encounters`
  - `POST /api/v1/encounters`
  - `GET /api/v1/encounters/{id}`
- `users` table migration (User model already exists in code; persist it)
- `audit_log` table migration + shared audit logging helper (`internal/audit/log.go`)
- OpenAPI skeleton (`openapi.yaml`) with shared components: error schema, pagination schema, auth scheme
- `docker-compose.yml` confirmed working with migrations running on startup

> ⚠️ This branch must be merged to `main` **before** the 3 feature branches begin.

---

## Developer A — `feature/clinical-core`

**Theme:** Core clinical-facing modules — defines what is happening with a patient right now.

### Modules
| # | Module | Spec section |
|---|---|---|
| 1 | Care Team Assignment | 8.1 |
| 2 | Patient Flow Tracking | 8.2 |
| 3 | Clinical Task Board | 8.3 |
| 4 | Daily Huddle Dashboard | 8.9 |

### Backend deliverables
- **DB migrations:** `care_team_assignments`, `handoff_notes`, `flow_state_transitions`, `tasks`, `task_assignment_events`
- **Packages:** `internal/careteam`, `internal/flow`, `internal/task`, `internal/dashboard`
- **API endpoints (~14):**
  ```
  GET  /api/v1/encounters/{id}/care-team/assignments
  POST /api/v1/encounters/{id}/care-team/assignments
  POST /api/v1/care-team/assignments/{assignmentId}/transfer
  GET  /api/v1/encounters/{id}/handoffs
  GET  /api/v1/encounters/{id}/flow
  POST /api/v1/encounters/{id}/flow/transitions
  POST /api/v1/encounters/{id}/flow/override        (privileged; requires reason)
  GET  /api/v1/tasks
  POST /api/v1/tasks
  PATCH /api/v1/tasks/{id}
  POST /api/v1/tasks/{id}/assign
  POST /api/v1/tasks/{id}/complete
  GET  /api/v1/tasks/{id}/history
  GET  /api/v1/dashboard/huddle
  ```

### Frontend deliverables
- Wire existing `features/care-team/` components to backend API
- `features/flow/` — flow timeline tab with state transition UI
- `features/tasks/` — task board with filters, SLA overdue indicators, assignment UI
- `pages/HuddleDashboard.tsx` — census + risk indicators + unit/department filters

---

## Developer B — `feature/operational-logistics`

**Theme:** Operational logistics — bed, transport, and discharge; tooling for operational staff.

### Modules
| # | Module | Spec section |
|---|---|---|
| 1 | Bed Management | 8.5 |
| 2 | Transport Requests | 8.6 |
| 3 | Discharge Planning Checklist | 8.7 |

### Backend deliverables
- **DB migrations:** `beds`, `bed_status_events`, `bed_requests`, `transport_requests`, `transport_change_events`, `discharge_checklists`, `discharge_checklist_items`
- **Packages:** `internal/bed`, `internal/transport`, `internal/discharge`
- **API endpoints (~12):**
  ```
  GET  /api/v1/beds
  POST /api/v1/beds/{id}/status                           (creates BedStatusEvent)
  POST /api/v1/encounters/{id}/bed-requests
  POST /api/v1/bed-requests/{id}/assign
  GET  /api/v1/transport/requests
  POST /api/v1/transport/requests
  POST /api/v1/transport/requests/{id}/accept
  PATCH /api/v1/transport/requests/{id}                  (tracked changes)
  POST /api/v1/transport/requests/{id}/complete
  POST /api/v1/encounters/{id}/discharge-checklist/init
  GET  /api/v1/encounters/{id}/discharge-checklist
  POST /api/v1/discharge-checklist/items/{itemId}/complete
  POST /api/v1/encounters/{id}/discharge/complete         (validates checklist; supports override)
  ```

### Frontend deliverables
- `features/beds/` — bed board per unit with operational status indicators
- `features/transport/` — transport dispatch queue page
- `features/discharge/` — discharge checklist tab in encounter detail view

---

## Developer C — `feature/governance-safety`

**Theme:** Governance, safety, and inter-department coordination.

### Modules
| # | Module | Spec section |
|---|---|---|
| 1 | Inter-department Consult Requests | 8.4 |
| 2 | Exception Workflows | 8.8 |
| 3 | Quality/Safety Incident Logging | 8.10 |

### Backend deliverables
- **DB migrations:** `consult_requests`, `exception_events`, `incidents`, `incident_status_events`
- **Packages:** `internal/consult`, `internal/exception`, `internal/incident`
- **API endpoints (~10):**
  ```
  GET  /api/v1/consults
  POST /api/v1/consults
  POST /api/v1/consults/{id}/accept
  POST /api/v1/consults/{id}/decline
  POST /api/v1/consults/{id}/redirect
  POST /api/v1/consults/{id}/complete
  GET  /api/v1/exceptions
  POST /api/v1/exceptions
  POST /api/v1/exceptions/{id}/finalize
  POST /api/v1/exceptions/{id}/correct
  GET  /api/v1/incidents
  POST /api/v1/incidents
  POST /api/v1/incidents/{id}/status                     (creates status event)
  ```

### Frontend deliverables
- `features/consults/` — consult inbox (by service) + consult submission form
- `features/exceptions/` — exception workflow UI with structured required fields + correction flow
- `features/incidents/` — incident logging form + incident review queue page

---

## Shared Conventions (all developers must follow)

### Backend (Go)
- One package per domain under `internal/{domain}/` with: `handler.go`, `service.go`, `repository.go`, `models.go`
- Register routes in a domain-specific `routes.go`, wired into the main router
- Every state-changing handler writes to `audit_log` via shared `internal/audit` package
- Event tables (`*_events`, `*_transitions`) are **insert-only** — no updates, no deletes
- Standard error format for all responses:
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Human-readable summary",
      "details": [{ "field": "...", "issue": "..." }],
      "correlationId": "..."
    }
  }
  ```
- UTC timestamps everywhere; `createdAt` / `updatedAt` on all entities
- Unit test coverage >80% on core business logic; table-driven tests preferred

### Frontend (React/TS)
- Feature folder structure: `features/{name}/components/`, `hooks/`, `pages/`, `services/`, `types/`
- All API calls go through `shared/utils/api.ts` (centralized fetch instance with auth headers)
- Role-aware rendering via `usePermissions` hook (already exists in `features/auth/hooks/`)
- Register new routes in `shared/config/routes.ts`

### OpenAPI
- Each developer appends their endpoint specs to `openapi.yaml` in their feature branch
- Paths are non-overlapping by design — merge conflicts are not expected

---

## Dependency Map

```
Phase 0: feature/foundation  ──► merge to main
                                      │
              ┌───────────────────────┼───────────────────────┐
              ▼                       ▼                       ▼
   feature/clinical-core   feature/operational-    feature/governance-
        (Dev A)              logistics (Dev B)        safety (Dev C)
```

All three feature branches are **fully independent** after Phase 0 merges.

## Merge Order
1. `feature/foundation` → `main`
2. `feature/clinical-core`, `feature/operational-logistics`, `feature/governance-safety` → `main` (any order; paths are non-conflicting)
