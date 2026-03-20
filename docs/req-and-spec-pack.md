# WardFlow — Technical Requirements, Specs, and Engineering Instructions
 
**Source:** Derived from Docmind document “Task Description” (`task-description`).
 
## 0) Goal / problem statement
WardFlow is an inpatient/ED care-coordination system aimed at reducing missed actions during handoffs, improving real-time operational visibility (patient flow + bed/transport/consult), and providing accountable tasking and compliance-ready audit trails.
 
## 1) Scope (MVP)
### In-scope modules (from the task description)
1. Care team assignment per encounter
2. Patient flow tracking (state timeline)
3. Clinical task board
4. Inter-department consult requests
5. Bed management
6. Transport requests
7. Discharge planning checklist
8. Exception workflows
9. Daily huddle dashboard
10. Quality/safety incident logging
 
### Out of scope (explicitly)
- EHR order entry / medication administration
- Billing/claims, coding, charge capture
- Full clinical documentation (notes) beyond structured handoff/exception fields
 
## 2) Personas & roles (RBAC)
Minimum roles (extendable):
- **Nurse** (unit-based)
- **Provider/Clinician**
- **Charge nurse / Unit coordinator**
- **Operations / Throughput**
- **Consult service (receiver)**
- **Transport staff / Dispatch**
- **Quality/Safety reviewer**
- **Admin**
 
RBAC must enforce:
- Unit/department visibility boundaries
- Privileged overrides (e.g., invalid flow transitions) only via explicit workflow + audit
 
## 3) Mandatory technical baseline (hard requirements)
- **Dockerisation:** application runs via containers.
- **OpenAPI specification:** API contracts explicitly defined.
- **Auth & Authz:** secured endpoints with proper access control.
- **Data storage:** persistent storage (DB choice open).
- **Error handling:** standardized error formats and HTTP codes.
- **Unit test coverage:** >80% coverage on *core logic*.
- **Frontend mandatory:** decoupled UI consumes the APIs.
- **REST API:** JSON input/output.
 
## 4) Non-functional requirements
### Security & privacy
- All endpoints require authentication except health/version endpoints.
- RBAC at API layer (never rely on frontend-only checks).
- Encrypt in transit (TLS) and at rest (DB volume encryption in production).
- Audit log for create/update/override actions on clinical/operational objects.
 
### Auditability & history (first-class)
- **No destructive overwrites** for timeline-like concepts: care team assignments, flow state transitions, task ownership changes, bed status changes, incident status changes.
- Every state mutation records: `who`, `when`, `why` (where applicable), and `source` (user action vs system).
 
### Reliability & performance (MVP targets)
- P95 API latency < 300ms for common reads (lists/detail) at moderate load.
- Support near-real-time updates via polling (MVP) and optional websocket/SSE later.
- Idempotent creates for integrations (client-supplied idempotency key).
 
### Compliance-ready behaviors
- Retention policy configurable (default: keep encounter lifecycle history). 
- Immutable “finalized” records for certain workflows (exceptions, incidents) with correction workflow.
 
## 5) System architecture (reference)
### Components
- **Frontend**: SPA (React/Vue/etc), role-aware navigation.
- **API service**: REST JSON + OpenAPI, responsible for authz, validation, business rules.
- **Database**: relational recommended (PostgreSQL) for history + audit queries.
- **Background worker** (optional MVP): SLA checks (overdue tasks), notifications, scheduled dashboard materialization.
 
### Integration points (design for, but optional MVP)
- AD/SSO (OIDC) for identity.
- HL7/FHIR feed for encounters/patients.
 
## 6) Data model (conceptual)
### Core entities
- `User(id, name, email, departments[], units[], roleBindings[])`
- `Encounter(id, patientId, unitId, departmentId, status, startedAt, endedAt?)`
- `CareTeamAssignment(id, encounterId, userId, roleType, startsAt, endsAt?, createdBy, createdAt, handoffNoteId?)`
- `HandoffNote(id, encounterId, fromUserId, toUserId, roleType, note, structuredFieldsJSON, createdAt)`
- `FlowStateTransition(id, encounterId, fromState?, toState, at, actorType, actorUserId?, reason?, sourceEventId?)`
- `Task(id, scopeType{encounter|patient|unit}, scopeId, title, details, status, priority, slaDueAt?, createdBy, createdAt, completedBy?, completedAt?)`
- `TaskAssignmentEvent(id, taskId, fromOwner?, toOwner, at, byUserId, reason?)`
- `ConsultRequest(id, encounterId, targetService, reason, urgency, status{pending|accepted|declined|completed|redirected|cancelled}, createdBy, createdAt, acceptedBy?, acceptedAt?, closedAt?, closeReason?)`
- `Bed(id, unitId, room, label, capabilitiesJSON, currentStatus, currentEncounterId?, updatedAt)`
- `BedStatusEvent(id, bedId, fromStatus?, toStatus, at, byUserId, reason?)`
- `BedRequest(id, encounterId, requiredCapabilitiesJSON, priority, status{pending|assigned|cancelled}, createdAt, createdBy)`
- `TransportRequest(id, encounterId, origin, destination, priority, status{pending|assigned|in_transit|completed|cancelled}, createdBy, createdAt, assignedTo?, assignedAt?, updatedAt)`
- `TransportChangeEvent(id, requestId, changedFieldsJSON, at, byUserId, reason)`
- `DischargeChecklist(id, encounterId, dischargeType, status, createdAt)`
- `DischargeChecklistItem(id, checklistId, code, label, required, status{open|done|waived}, completedBy?, completedAt?)`
- `ExceptionEvent(id, encounterId, type, status{draft|finalized|corrected}, requiredFieldsJSON, dataJSON, initiatedBy, initiatedAt, finalizedBy?, finalizedAt?)`
- `Incident(id, encounterId?, type, severity?, harmIndicatorsJSON?, eventTime, reportedBy, reportedAt, status{submitted|under_review|closed}, statusEvents[])`
 
### Cross-cutting
- `AuditLog(id, entityType, entityId, action, at, byUserId, ip?, userAgent?, beforeJSON?, afterJSON?)`
 
## 7) API standards (apply to all endpoints)
- Base path: `/api/v1`.
- Auth: OAuth2/OIDC JWT bearer (or equivalent) with scopes/roles.
- Pagination: `limit` + `cursor` (preferred) or `limit` + `offset` (MVP acceptable).
- Filtering: by `unitId`, `departmentId`, `encounterId`, `status`, date ranges.
- Concurrency: optimistic locking using `ETag`/`If-Match` or `version` fields for mutable resources.
 
### Standard error format
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
 
## 8) Module specs
 
### 8.1 Care team assignment per encounter
**Business rules (source):**
- Every active encounter has ≥1 primary clinical owner.
- Assignment includes user, role type, effective start, encounter ref.
- Transfers end-date prior assignment (no overwrite).
- Critical-role handoffs require structured notes before completion.
- Assignment history queryable for full encounter lifecycle.
 
**Key workflows**
- Assign role → creates `CareTeamAssignment(startsAt=now)`.
- Transfer primary role → requires `HandoffNote`, end-dates prior assignment.
- View care team → returns current assignments + history.
 
**API (minimum)**
- `GET /encounters/{id}/care-team`
- `POST /encounters/{id}/care-team/assignments`
- `POST /care-team/assignments/{assignmentId}/transfer`
- `GET /encounters/{id}/handoffs`
 
**Acceptance criteria (source)**
- Care team view shows current assigned staff by role.
- Ownership transfer requires handoff details.
- Reassignment preserves prior assignment with end timestamp.
- Department transfer preserves old/new team assignments in audit history.
 
### 8.2 Patient flow tracking
**Business rules (source)**
- Each state change timestamped.
- Transitions linked to system event or user action.
- Manual edits capture user + reason.
- Preserve all prior transitions.
- Block invalid transitions unless authorized override workflow.
 
**Flow model**
- Define canonical states per facility (config-driven), e.g. `arrived`, `triage`, `provider_eval`, `diagnostics`, `admitted`, `discharge_ready`, `discharged`.
 
**API (minimum)**
- `GET /encounters/{id}/flow`
- `POST /encounters/{id}/flow/transitions`
- `POST /encounters/{id}/flow/override` (privileged; requires reason)
 
**Acceptance criteria (source)**
- Arrival completion appears in timeline.
- Provider evaluation updates status + timeline.
- Correction records who/why.
- Ops review shows all transitions chronologically.
 
### 8.3 Clinical task board
**Business rules (source)**
- Tasks associated with encounter/patient/unit.
- Statuses: `open`, `in_progress`, `completed`, `cancelled`, `escalated`.
- SLA tasks marked overdue when due passes.
- Only authorized users can reassign/close tasks owned by another role.
- Completed tasks retain completion metadata.
 
**API (minimum)**
- `GET /tasks?scopeType=&scopeId=&status=&overdue=`
- `POST /tasks`
- `PATCH /tasks/{id}`
- `POST /tasks/{id}/assign`
- `POST /tasks/{id}/complete`
- `GET /tasks/{id}/history`
 
**Acceptance criteria (source)**
- New action appears in patient + unit views.
- SLA threshold marks overdue.
- Reassignment shows new owner + preserves assignment history.
- Later review shows who completed + when.
 
### 8.4 Inter-department consult requests
**Business rules (source)**
- Consult includes encounter, target service, reason, urgency.
- Cannot complete unless accepted/acknowledged.
- Declined/redirected require reason.
 
**API (minimum)**
- `GET /consults?unitId=&status=`
- `POST /consults`
- `POST /consults/{id}/accept`
- `POST /consults/{id}/decline`
- `POST /consults/{id}/redirect`
- `POST /consults/{id}/complete`
 
**Acceptance criteria (source)**
- Submission shows in receiving department pending queue.
- Decline requires reason.
 
### 8.5 Bed management
**Business rules (source)**
- Each bed has one active operational status.
- Occupied/blocked/maintenance beds cannot be assigned.
- Placement rules evaluate compatibility constraints.
- Patient waiting for placement has visible pending bed request status.
- Bed status changes record who made change.
 
**API (minimum)**
- `GET /beds?unitId=&status=`
- `POST /beds/{id}/status` (creates `BedStatusEvent`)
- `POST /encounters/{id}/bed-requests`
- `POST /bed-requests/{id}/assign` (assign bed)
 
**Acceptance criteria (source)**
- Discharge → bed cleaning required → no longer assignable.
- Isolation requirement excludes non-compatible beds (or blocks assignment).
 
### 8.6 Transport requests
**Business rules (source)**
- Request includes encounter context, origin, destination, priority.
- Cannot complete unless assigned/accepted by transport.
- Post-assignment changes (pickup/destination/priority) must be tracked.
 
**API (minimum)**
- `GET /transport/requests?status=&unitId=`
- `POST /transport/requests`
- `POST /transport/requests/{id}/accept`
- `PATCH /transport/requests/{id}` (tracked changes)
- `POST /transport/requests/{id}/complete`
 
**Acceptance criteria (source)**
- Submitting shows in dispatch queue.
- Acceptance sets status to assigned.
 
### 8.7 Discharge planning checklist
**Business rules (source)**
- Checklist associated with encounter.
- Items configurable required/optional by discharge type or unit.
- Required items must be done before discharge-complete, unless override.
 
**API (minimum)**
- `POST /encounters/{id}/discharge-checklist/init`
- `GET /encounters/{id}/discharge-checklist`
- `POST /discharge-checklist/items/{itemId}/complete`
- `POST /encounters/{id}/discharge/complete` (validates checklist; supports override)
 
**Acceptance criteria (source)**
- Workflow start generates configured items.
- Finalize discharge blocked (or override required) if required items incomplete.
- Sign-off stores user + timestamp.
- View shows completed + outstanding.
 
### 8.8 Exception workflows
**Business rules (source)**
- Exception tied to encounter + type.
- Mandatory documentation fields required before completion.
- Record who initiated/finalized.
- Certain types trigger downstream notifications/tasks.
- Finalized records immutable except via correction workflow.
 
**API (minimum)**
- `GET /exceptions?encounterId=&type=&status=`
- `POST /exceptions`
- `POST /exceptions/{id}/finalize`
- `POST /exceptions/{id}/correct`
 
**Acceptance criteria (source)**
- LWBS requires configured documentation fields.
- AMA discharge finalization captures responsible user + timestamps.
- Follow-up triggers related task/notification.
- Direct overwrite blocked unless correction workflow used.
 
### 8.9 Daily huddle dashboard
**Business rules (source)**
- Reflect current encounter/workflow states.
- High-risk/delayed visually distinguishable.
- Filter by unit/department.
- Authorization limits visible scope.
- Drill-down respects same access controls.
 
**API (minimum)**
- `GET /dashboard/huddle?unitId=&departmentId=`
 
**Acceptance criteria (source)**
- Unit manager sees census + risk indicators.
- Expected discharges listed.
- Delayed consult/overdue task visibly flagged.
- Filter updates metrics + lists.
 
### 8.10 Quality/safety incident logging
**Business rules (source)**
- Incident includes type, event time, reporting user.
- Severity/harm indicators required by type.
- Review status changes tracked submission→closure.
- Only designated roles finalize reviews.
- Incident records auditable.
 
**API (minimum)**
- `POST /incidents`
- `GET /incidents?unitId=&status=&type=`
- `POST /incidents/{id}/status` (creates status event)
 
**Acceptance criteria (source)**
- Logging fall stores type, time, encounter ref.
- Severity required where applicable.
- Reviewer updates tracked.
- Unauthorized close blocked.
 
## 9) Frontend requirements
Minimum screens:
- Login / session expired
- Unit patient list (filters by unit, state, risks)
- Encounter detail with tabs: Care team, Flow timeline, Tasks, Consults, Bed/Placement, Transport, Discharge checklist, Exceptions, Incidents
- Consult inbox (by service)
- Transport dispatch queue
- Daily huddle dashboard
- Incident review queue (quality/safety)
 
UI behavior:
- Role-based actions (buttons hidden + server-enforced).
- Inline audit/history views for assignments/transitions.
 
## 10) Engineering instructions (implementation constraints)
### API implementation
- Generate/maintain `openapi.yaml` (source-controlled). 
- Use request validation (schema) at boundary.
- Enforce RBAC via middleware + per-entity scope checks.
- Use UTC timestamps everywhere; store `createdAt`, `updatedAt`, and event times separately.
 
### Testing
- Unit tests for business rules and transitions (target >80% core logic).
- Integration tests for RBAC + invalid transition blocking.
- Contract tests to ensure OpenAPI stays accurate.
 
### Docker & runtime
- Provide `docker-compose.yml` for local dev: api + db + frontend.
- Provide one-command startup: `docker compose up`.
- Provide migration command/run on startup.
 
### Observability
- Structured JSON logs with `correlationId` per request.
- Health endpoints: `GET /healthz` (liveness), `GET /readyz` (readiness).
 
---
 
## Appendix: Source citations
- Docmind: “Task Description” (`task-description`) — module rules and acceptance criteria.