# WardFlow Frontend - Governance & Safety Features

This document describes the three governance and safety features implemented for Developer C.

## Features Implemented

### 1. Consults (`src/features/consults/`)

Complete consultation request management system.

**Key Components:**
- `ConsultInbox` - List view with status filters (all, pending, accepted, completed, declined, redirected)
- `ConsultCard` - Individual consult display with role-aware action buttons
- `ConsultForm` - Modal form for creating new consults
- `DeclineModal` - Modal for declining consults with required reason
- `RedirectModal` - Modal for redirecting consults to different services

**Features:**
- ✅ Urgency indicators (routine=gray, urgent=orange, emergent=red)
- ✅ Status badges with color coding
- ✅ Role-based actions (provider, consult, admin can manage)
- ✅ Accept, decline, redirect, and complete workflows
- ✅ Metadata tracking (created by, accepted by, timestamps)

**Pages:**
- `/consults` - Main consults page with inbox and filters

**API Integration:**
- `GET /api/v1/consults` - List consults
- `POST /api/v1/consults` - Create consult
- `POST /api/v1/consults/{id}/accept` - Accept
- `POST /api/v1/consults/{id}/decline` - Decline (with reason)
- `POST /api/v1/consults/{id}/redirect` - Redirect (with target + reason)
- `POST /api/v1/consults/{id}/complete` - Complete

---

### 2. Exceptions (`src/features/exceptions/`)

Exception event tracking with draft/finalize/correct workflow.

**Key Components:**
- `ExceptionList` - List with status and type filters
- `ExceptionForm` - Create/edit exception with JSON data
- `FinalizeModal` - Confirmation modal with immutability warning
- `CorrectionModal` - Create correction for finalized exceptions

**Features:**
- ✅ Draft exceptions can be edited
- ✅ Finalized exceptions are immutable
- ✅ Corrections create new exception events
- ✅ JSON data editor with validation
- ✅ Status badges (draft=yellow, finalized=green, corrected=blue)
- ✅ Role-based corrections (quality_safety, admin only)

**Pages:**
- `/exceptions` - Main exceptions page with list and filters

**API Integration:**
- `GET /api/v1/exceptions` - List exceptions
- `POST /api/v1/exceptions` - Create exception
- `PATCH /api/v1/exceptions/{id}` - Update draft exception
- `POST /api/v1/exceptions/{id}/finalize` - Finalize (make immutable)
- `POST /api/v1/exceptions/{id}/correct` - Create correction

---

### 3. Incidents (`src/features/incidents/`)

Safety incident reporting and review system.

**Key Components:**
- `IncidentForm` - Report new incidents (all users)
- `IncidentList` - List with status filters
- `IncidentReviewQueue` - Queue for quality_safety role
- `IncidentDetail` - Full incident view with status history timeline
- `StatusUpdateModal` - Update status (quality_safety, admin only)

**Features:**
- ✅ Anyone can report incidents
- ✅ Quality & safety team reviews incidents
- ✅ Status workflow: submitted → under_review → closed
- ✅ Severity levels (minor, moderate, severe, critical)
- ✅ Complete status history with timeline view
- ✅ Optional encounter association
- ✅ Harm indicators (JSON structure)

**Pages:**
- `/incidents/report` - Report new incident (all authenticated users)
- `/incidents/review` - Review queue (quality_safety, admin only)

**API Integration:**
- `GET /api/v1/incidents` - List incidents
- `POST /api/v1/incidents` - Create incident
- `GET /api/v1/incidents/{id}` - Get incident details
- `POST /api/v1/incidents/{id}/status` - Update status
- `GET /api/v1/incidents/{id}/status-history` - Get status history

---

## Architecture & Patterns

All three features follow the same established pattern:

```
features/{name}/
├── components/        # Feature-specific UI components
├── hooks/            # Custom hooks (useConsults, useConsultActions)
├── pages/            # Page-level components
├── services/         # API service layer
├── store/            # Zustand state management
├── types/            # TypeScript type definitions
└── index.ts          # Public exports
```

### State Management (Zustand)

Each feature has its own store:
- `useConsultStore` - Consults state
- `useExceptionStore` - Exceptions state
- `useIncidentStore` - Incidents state

### API Layer

Each feature has a service file with all API calls:
- `consultService.ts`
- `exceptionService.ts`
- `incidentService.ts`

Uses the shared Axios instance from `src/shared/utils/api.ts`.

### Hooks Pattern

**Data fetching hooks:**
- `useConsults(filters)` - Fetch and filter consults
- `useExceptions(filters)` - Fetch and filter exceptions
- `useIncidents(filters)` - Fetch and filter incidents

**Action hooks:**
- `useConsultActions()` - CRUD operations for consults
- `useExceptionActions()` - CRUD operations for exceptions
- `useIncidentActions()` - CRUD operations for incidents

### Role-Based Access Control

Uses `usePermissions()` hook from auth feature:
```typescript
const { hasAnyRole } = usePermissions();
const canManage = hasAnyRole(['provider', 'consult', 'admin']);
```

**Permissions:**
- **Consults**: provider, consult, admin can manage
- **Exceptions**: quality_safety, admin can correct finalized exceptions
- **Incidents**: All can report, quality_safety & admin can review

---

## Routes

Added to `src/shared/config/routes.ts`:

```typescript
CONSULT_LIST: '/consults'
EXCEPTION_LIST: '/exceptions'
INCIDENT_REPORT: '/incidents/report'
INCIDENT_REVIEW: '/incidents/review'
```

All routes are protected and require authentication.

---

## UI Components Used

All features use the shared UI component library:

- `Button` - Action buttons with variants (primary, secondary, danger)
- `Card` - Container component for content
- `Modal` - Dialog overlays
- `Badge` - Status indicators
- `Input` - Text inputs
- `Select` - Dropdown selects
- `Spinner` - Loading indicators

### Styling

- **Tailwind CSS** for utility-first styling
- **Color Conventions:**
  - Blue: Primary actions, accepted, under_review
  - Green: Completed, finalized, success
  - Yellow: Pending, draft, warnings
  - Red: Declined, errors, critical severity
  - Orange: Urgent severity
  - Purple: Redirected
  - Gray: Routine, closed, cancelled

---

## Error Handling

All features implement error handling:
1. Try-catch in action hooks
2. Error state in stores
3. Error displays in UI components
4. Console logging for debugging

**Note:** Toast notifications are TODO (placeholders in place).

---

## Form Validation

- Required fields marked with `*`
- Client-side validation before API calls
- JSON validation for structured data fields
- Error messages displayed inline

---

## Testing Checklist

### Consults
- [ ] Create new consult with all urgency levels
- [ ] Accept pending consult
- [ ] Decline consult with reason
- [ ] Redirect consult to another service
- [ ] Complete accepted consult
- [ ] Filter by status
- [ ] Role-based button visibility

### Exceptions
- [ ] Create draft exception with JSON data
- [ ] Edit draft exception
- [ ] Finalize draft exception
- [ ] Attempt to edit finalized (should be disabled)
- [ ] Create correction for finalized exception
- [ ] Filter by status and type
- [ ] Role-based correction access

### Incidents
- [ ] Report incident with all fields
- [ ] Report incident without optional fields
- [ ] View incident list with filters
- [ ] Access review queue (quality_safety role)
- [ ] Update incident status
- [ ] View status history timeline
- [ ] Role-based access to review page

---

## Next Steps / TODs

1. **Toast Notifications**: Integrate toast notifications for success/error messages
2. **Form Libraries**: Consider using react-hook-form + zod for more robust validation
3. **Loading States**: Add skeleton loaders for better UX
4. **Pagination**: Add pagination for large lists
5. **Search**: Add search functionality to lists
6. **Export**: Add export functionality for reports
7. **Attachments**: Add file upload support for incidents
8. **Real-time Updates**: Consider WebSocket for live updates
9. **Analytics**: Add dashboards for governance metrics
10. **Audit Trail**: Enhance audit logging for all actions

---

## File Structure Summary

```
src/features/
├── consults/
│   ├── components/
│   │   ├── ConsultCard.tsx
│   │   ├── ConsultForm.tsx
│   │   ├── ConsultInbox.tsx
│   │   ├── DeclineModal.tsx
│   │   └── RedirectModal.tsx
│   ├── hooks/
│   │   ├── useConsultActions.ts
│   │   └── useConsults.ts
│   ├── pages/
│   │   └── ConsultsPage.tsx
│   ├── services/
│   │   └── consultService.ts
│   ├── store/
│   │   └── consultStore.ts
│   ├── types/
│   │   └── consult.types.ts
│   └── index.ts
├── exceptions/
│   ├── components/
│   │   ├── CorrectionModal.tsx
│   │   ├── ExceptionForm.tsx
│   │   ├── ExceptionList.tsx
│   │   └── FinalizeModal.tsx
│   ├── hooks/
│   │   ├── useExceptionActions.ts
│   │   └── useExceptions.ts
│   ├── pages/
│   │   └── ExceptionsPage.tsx
│   ├── services/
│   │   └── exceptionService.ts
│   ├── store/
│   │   └── exceptionStore.ts
│   ├── types/
│   │   └── exception.types.ts
│   └── index.ts
└── incidents/
    ├── components/
    │   ├── IncidentDetail.tsx
    │   ├── IncidentForm.tsx
    │   ├── IncidentList.tsx
    │   ├── IncidentReviewQueue.tsx
    │   └── StatusUpdateModal.tsx
    ├── hooks/
    │   ├── useIncidentActions.ts
    │   └── useIncidents.ts
    ├── pages/
    │   ├── IncidentReportPage.tsx
    │   └── IncidentReviewPage.tsx
    ├── services/
    │   └── incidentService.ts
    ├── store/
    │   └── incidentStore.ts
    ├── types/
    │   └── incident.types.ts
    └── index.ts
```

---

## Integration with Backend

All features are ready to integrate with the Go backend API:
- Base URL: `http://localhost:8080/api/v1`
- Auth: Bearer token in Authorization header (handled by Axios interceptor)
- Content-Type: application/json

Backend endpoints should match the service layer calls.

---

## Developer Notes

**Best Practices Followed:**
- ✅ Functional components with hooks
- ✅ TypeScript for type safety
- ✅ Separation of concerns (components, hooks, services, stores)
- ✅ Reusable UI components
- ✅ Consistent error handling
- ✅ Role-based access control
- ✅ Responsive design with Tailwind
- ✅ Consistent naming conventions
- ✅ Clear comments and documentation

**Performance Considerations:**
- Zustand for lightweight state management
- Filtered API calls to reduce data transfer
- Optimistic UI updates where appropriate
- Lazy loading for modals

---

For questions or issues, refer to the main project documentation or contact the development team.
