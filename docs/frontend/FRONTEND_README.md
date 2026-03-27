# WardFlow Frontend

React application for WardFlow inpatient/ED care coordination system.

## Architecture

### Project Structure

```
src/
├── features/              # Feature modules (domain-driven)
│   ├── auth/             # Authentication module
│   │   ├── components/   # Auth-specific components
│   │   ├── hooks/        # useAuth, usePermissions
│   │   ├── pages/        # Login, Register pages
│   │   ├── services/     # Auth API calls
│   │   ├── store/        # Zustand auth store
│   │   └── types/        # Auth TypeScript types
│   │
│   └── care-team/        # Care Team Assignment module
│       ├── components/   # Care team components
│       ├── hooks/        # useCareTeam
│       ├── pages/        # Care team pages
│       ├── services/     # Care team API
│       ├── store/        # Zustand store
│       └── types/        # Care team types
│
├── shared/               # Shared across features
│   ├── components/       # Reusable UI components
│   │   ├── ui/          # Button, Input, Card, Modal, etc.
│   │   ├── layout/      # Header, Sidebar, Layout
│   │   └── feedback/    # Loading, Error, Empty states
│   ├── hooks/           # useApi, useDebounce, etc.
│   ├── utils/           # API client, formatters, validators
│   ├── types/           # Common TypeScript types
│   └── config/          # Route definitions
│
├── lib/                 # Third-party configurations
│   └── router.tsx       # React Router setup
│
├── pages/               # Top-level route pages
│   ├── DashboardPage.tsx
│   ├── NotFoundPage.tsx
│   └── UnauthorizedPage.tsx
│
├── App.tsx              # Root component
├── main.tsx             # Entry point
└── index.css            # Tailwind imports
```

### Tech Stack

- **React 18.3** - Modern hooks, concurrent features
- **TypeScript** - Type safety
- **Vite 8** - Build tool and dev server
- **Tailwind CSS 4.2** - Utility-first styling
- **React Router 6** - Client-side routing
- **Zustand 4** - Lightweight state management
- **Axios 1** - HTTP client with interceptors
- **React Hook Form 7** - Form handling
- **Zod 3** - Schema validation
- **date-fns 3** - Date/time utilities
- **Lucide React** - Icon library
- **React Hot Toast** - Notifications

## Key Patterns

### 1. Feature-Based Organization
Each feature module contains all related code (components, hooks, services, types).

### 2. Zustand Stores
```typescript
// features/auth/store/authStore.ts
export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  isAuthenticated: false,
  login: async (credentials) => { /* ... */ },
  logout: () => { /* ... */ },
}));
```

### 3. API Service Layer
```typescript
// features/care-team/services/careTeamService.ts
export const careTeamService = {
  getByEncounter: (id) => api.get(`/encounters/${id}/care-team`),
  assign: (data) => api.post('/care-team/assignments', data),
};
```

### 4. Custom Hooks
```typescript
// features/auth/hooks/useAuth.ts
export const useAuth = () => {
  const { user, login, logout } = useAuthStore();
  return { user, login, logout };
};
```

### 5. RBAC Permission Checks
```typescript
const { hasRole, hasPermission } = usePermissions();

{hasRole('charge_nurse') && (
  <Button onClick={onAssign}>Assign Role</Button>
)}
```

## Development

### Setup
```bash
npm install
```

### Dev Server
```bash
npm run dev
# Opens on http://localhost:5173
```

### Build
```bash
npm run build
# Outputs to dist/
```

### Lint
```bash
npm run lint
```

## Environment Variables

Create `.env` file:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## Backend Integration

API client automatically adds auth token to requests:

```typescript
// shared/utils/api.ts intercepts all requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
```

## Module Status

### ✅ Implemented (Structure Only - TODOs in place)

**Phase 1 - Foundation:**
- ✅ Shared UI components (Button, Input, Select, Card, Modal, Badge, Spinner)
- ✅ Layout components (Header, Sidebar, Layout, PageHeader)
- ✅ Feedback components (ErrorBoundary, LoadingState, EmptyState)
- ✅ Shared hooks (useApi, useDebounce, useLocalStorage, useMediaQuery)
- ✅ Shared utilities (API client, formatters, validators, constants)
- ✅ TypeScript types (User, Encounter, Role, RBAC, API responses)
- ✅ Router setup with route constants

**Phase 2 - Authentication:**
- ✅ Auth store (Zustand with persistence)
- ✅ Auth service (login, register, logout, me)
- ✅ Auth hooks (useAuth, usePermissions)
- ✅ Login/Register pages (placeholder)
- ✅ Protected route component

**Phase 3 - Care Team:**
- ✅ Care team types
- ✅ Care team service (API calls)
- ✅ Care team store
- ✅ Care team hooks
- ✅ Component files created (empty)

### ⏳ Next Steps (Implementation Needed)

1. Implement UI component logic (Button variants, Input validation, etc.)
2. Build Login/Register forms with React Hook Form + Zod
3. Implement Care Team components (List, Member card, Forms, History)
4. Add other modules (Tasks, Consults, Beds, Transport, etc.)
5. Connect pages to Layout wrapper
6. Add navigation to Sidebar
7. Implement Dashboard page
8. Add tests (React Testing Library)

## Notes

- **RBAC:** All permission checks are client-side hints. Backend enforces permissions.
- **Audit Trail:** Components should display full history (created/updated by/at).
- **Type Safety:** All API responses typed to match backend DTOs.
- **Error Handling:** Global error interceptor in axios, toast notifications on errors.
- **State:** Zustand for feature state, React Hook Form for form state.

## Contributing

Follow established patterns:
- Feature modules are self-contained
- Shared code goes in shared/
- Use TypeScript strict mode
- Follow naming conventions (PascalCase components, camelCase utils)
- Add TODO comments for implementation details
