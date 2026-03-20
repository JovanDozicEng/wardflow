Mandatory Technical Baseline:
 
Dockerisation: Application must run via containers.
OpenAPI Specification: API contracts must be explicitly defined.
Auth & Authz: Secured endpoints with proper access control.
Data Storage: Persistent storage implemented (choice of DB is open).
Error Handling: Standardized error formats and HTTP codes.
Unit Test Coverage: >80% coverage on core logic.
Frontend Mandatory: A decoupled UI must consume the APIs.
REST API: JSON as the standard input/output format.
# Module 3 — Care Coordination & Tasking (Inpatient/ED Workflow)
 
## 1. Care team assignment per encounter
 
### Business need
Care delivery breaks down when staff are unclear about who currently owns the patient or which responsibilities were transferred during handoff. This feature creates clear responsibility for each encounter and reduces missed actions during shift changes and departmental transfers.
 
### Business rules
- Every active encounter must have at least one assigned primary clinical owner.
- A role assignment must include the assigned user, role type, effective start time, and encounter reference.
- When a primary role is transferred, the previous assignment must be end-dated rather than overwritten.
- Handoffs for designated critical roles must require structured notes before completion.
- Assignment history must remain queryable for the full encounter lifecycle.
 
### Acceptance criteria
- Given an active encounter, when a user opens the care team view, then the current assigned staff by role are displayed.
- Given a shift change, when the outgoing nurse transfers ownership, then the system requires handoff details before the transfer is completed.
- Given a reassignment, when the new owner is saved, then the prior assignment remains visible in history with an end timestamp.
- Given a department transfer, when a new care team is assigned, then both old and new team assignments are preserved in audit history.
 
---
 
## 2. Patient flow tracking
 
### Business need
Operational teams need real-time visibility into where patients are in the care journey in order to manage throughput, identify bottlenecks, and reduce excessive wait times.
 
### Business rules
- Each encounter state change must be timestamped.
- State transitions must be linked to either a system event or a user action.
- Manual edits to flow states must capture the user and reason for correction.
- The system must preserve all prior state transitions rather than only the latest status.
- Invalid state transitions must be blocked unless performed by an authorized override workflow.
 
### Acceptance criteria
- Given a newly arrived patient, when intake marks arrival complete, then the patient flow timeline shows the arrival state with timestamp.
- Given a patient moves to provider evaluation, when that workflow step is completed, then the status updates and appears in the timeline.
- Given a staff member corrects an incorrect flow state, when the update is saved, then the system records who made the correction and why.
- Given a patient encounter, when an operations user reviews the timeline, then all completed state transitions are displayed in chronological order.
 
---
 
## 3. Clinical task board
 
### Business need
Care coordination work is often spread across verbal communication and fragmented notes, which increases the risk of missed actions. A task board creates a single accountable workspace for time-sensitive patient work.
 
### Business rules
- Every task must be associated with an encounter, patient, or operational unit.
- Tasks must support statuses such as open, in progress, completed, cancelled, or escalated.
- Tasks with defined SLA targets must be marked overdue when the due threshold passes without completion.
- Only authorized users can reassign or close tasks owned by another role.
- Completed tasks must retain completion metadata including user and timestamp.
 
### Acceptance criteria
- Given a new discharge planning action, when the task is created, then it appears in the relevant patient and unit task views.
- Given a task reaches its due time without completion, when the SLA threshold passes, then the task is marked overdue.
- Given a user with permission reassigns a task, when the update is saved, then the new owner is shown and assignment history is preserved.
- Given a completed task, when the encounter is reviewed later, then the system shows who completed it and when.
 
---
 
## 4. Inter-department consult requests
 
### Business need
Consult requests are frequently handled through paging, calls, or informal messages, which makes them hard to track. This feature formalizes specialist coordination and reduces delays caused by ambiguous requests or missing ownership.
 
### Business rules
- A consult request must include requesting encounter, target department or service, reason, and urgency.
- A consult cannot be completed unless it has first been accepted or otherwise acknowledged by the receiving team.
- Declined or redirected consults must capture a reason.
- Consult status changes must be timestamped and auditable.
- Unaccepted consults past a configured threshold must be flagged for escalation.
 
### Acceptance criteria
- Given a clinician submits a consult, when all required fields are completed, then the receiving department sees it in their pending consult queue.
- Given a receiving team declines a consult, when the action is submitted, then a decline reason is required and visible to the requester.
- Given a consult is not acknowledged within the configured time, when the threshold is reached, then the system flags it as overdue.
- Given a requester views a consult, when they open the consult details, then they can see the full status history and current owner.
 
---
 
## 5. Bed management
 
### Business need
Poor bed coordination causes admission delays, hallway boarding, and wasted capacity. This feature improves visibility into real bed availability and supports safe placement decisions.
 
### Business rules
- Each bed must have one active operational status at a time.
- Beds marked occupied, blocked, or under maintenance cannot be assigned to another patient.
- Placement rules must evaluate configured compatibility constraints before assignment.
- A patient waiting for placement must have a visible pending bed request status.
- Bed status changes must record who made the change and when.
 
### Acceptance criteria
- Given a discharged patient leaves a room, when the bed status changes to cleaning required, then it is no longer available for assignment.
- Given a patient requires isolation, when staff attempt bed assignment, then non-compatible beds are excluded or blocked.
- Given a bed is marked available, when an eligible patient is assigned, then the bed status changes to reserved or occupied according to workflow.
- Given an operations user reviews bed turnover, when they inspect status history, then they can see all recent bed lifecycle changes with timestamps.
 
---
 
## 6. Transport requests
 
### Business need
Patient movement delays slow down diagnostics, discharges, and bed turnover. A structured transport workflow improves coordination and reduces uncertainty between departments.
 
### Business rules
- Every transport request must include patient or encounter context, origin, destination, and priority.
- A transport request cannot be completed unless it has first been assigned or accepted by transport staff.
- Changes to pickup, destination, or priority after assignment must be logged.
- Cancelled requests must capture a cancellation reason.
- Active transport requests must be visible in dispatcher and department-facing queues.
 
### Acceptance criteria
- Given a nurse submits a transport request, when all required details are entered, then the request appears in the transport dispatch queue.
- Given transport staff accept a request, when the assignment is saved, then the status changes to assigned.
- Given a department checks transport progress, when they open the request, then they can see the current status and timestamps.
- Given a request is cancelled, when the cancellation is submitted, then the system requires and stores a reason.
 
---
 
## 7. Discharge planning checklist
 
### Business need
Discharges are often delayed by missing paperwork, incomplete patient education, or coordination gaps. A discharge checklist reduces last-minute confusion and helps ensure patients leave with the right next steps.
 
### Business rules
- A discharge checklist must be associated with the encounter.
- Checklist items can be configured as required or optional by discharge type or unit.
- Required checklist items must be completed before the encounter can be marked discharge-complete, unless an override is used.
- Sign-off items must capture the approving user and time.
- Removed or skipped required items must capture a reason.
 
### Acceptance criteria
- Given a patient is entering discharge planning, when the workflow starts, then the system generates the configured checklist items for that encounter.
- Given a required item remains incomplete, when staff attempt to finalize discharge, then the system blocks completion or requires an override.
- Given a clinician signs off on discharge instructions, when the item is completed, then the sign-off user and timestamp are stored.
- Given a care coordinator reviews discharge readiness, when they open the checklist, then they can see all completed and outstanding items.
 
---
 
## 8. Exception workflows
 
### Business need
Exceptions in patient flow create legal, compliance, operational, and billing impacts. Standardized workflows reduce risk caused by incomplete or inconsistent documentation.
 
### Business rules
- Each exception event must be tied to an encounter and an exception type.
- Defined exception types must require mandatory documentation fields before completion.
- Exception workflows must record who initiated and finalized the exception.
- Certain exception types must trigger downstream notifications or review tasks.
- Exception records must remain immutable after finalization except through a tracked correction workflow.
 
### Acceptance criteria
- Given a patient leaves without being seen, when staff select that exception type, then the system requires the configured documentation fields before completion.
- Given an AMA discharge is finalized, when the record is saved, then the system captures the responsible user and timestamps.
- Given an exception type requires follow-up, when the workflow completes, then the related task or notification is triggered.
- Given a finalized exception record, when a user attempts to directly overwrite it, then the system blocks the change unless a correction workflow is used.
 
---
 
## 9. Daily huddle dashboard
 
### Business need
Shift-start huddles and operational reviews require a fast, shared picture of patient flow risks and discharge opportunities. This feature helps leaders prioritize action and align teams around today’s operational plan.
 
### Business rules
- Dashboard data must reflect current encounter and workflow states from connected module processes.
- High-risk or delayed items must be visually distinguishable from routine items.
- Metrics shown in the dashboard must be filterable by operational scope such as unit or department.
- Users may only see units or departments they are authorized to access.
- Patient-level drill-down must respect the same access controls as source workflows.
 
### Acceptance criteria
- Given a unit manager opens the dashboard, when data loads, then they see current census and operational risk indicators for their unit.
- Given expected discharges exist for the day, when the dashboard is viewed, then those patients are listed in the discharge-focused summary.
- Given a delayed consult or overdue task exists, when the dashboard is refreshed, then the item is visibly flagged.
- Given a user filters by unit, when the filter is applied, then all visible metrics and patient lists update accordingly.
 
---
 
## 10. Quality/safety incident logging
 
### Business need
Hospitals need a reliable way to capture safety events for compliance, learning, and process improvement. A dedicated incident workflow supports accountability and helps identify recurring operational risks.
 
### Business rules
- Every incident must include a type, event time, and reporting user.
- Severity and harm indicators must be recorded where required by incident type.
- Incident review status changes must be tracked from submission through closure.
- Only designated quality/safety roles can finalize incident reviews.
- Incident records must remain auditable, including any follow-up actions or review outcomes.
 
### Acceptance criteria
- Given a staff member logs a patient fall, when the form is submitted, then the incident is stored with type, time, and encounter reference.
- Given an incident requires severity classification, when the report is completed, then the system requires that field before submission.
- Given a reviewer updates the incident workflow, when they move it to under review or closed, then the status history is recorded.
- Given a non-authorized user attempts to close an incident, when they submit the action, then the system blocks it.
 
Delivery date: 30.03.2026