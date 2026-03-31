# WardFlow

Inpatient/ED care-coordination system designed to reduce missed actions during handoffs, improve real-time operational visibility, and provide accountable tasking with compliance-ready audit trails.

## Architecture

- **Backend**: Go 1.25+ with PostgreSQL 16
- **Frontend**: React 18+ with TypeScript and Tailwind CSS
- **Database**: PostgreSQL with JSONB support
- **Deployment**: Podman/Docker containers

## Quick Start

### Full Stack with Podman (Recommended)

Run all services (PostgreSQL, backend API, frontend) with a single command:

```bash
cd backend
podman compose up -d --build
```

| Service  | URL                   |
|----------|-----------------------|
| Frontend | http://localhost:3000 |
| Backend  | http://localhost:8080 |
| Database | localhost:5432        |

> **Note:** The first run downloads images and compiles the Go binary — allow ~2 minutes for the backend to become healthy.

**View logs:**
```bash
podman compose logs -f              # all services
podman compose logs -f backend      # backend only
podman compose logs -f frontend     # nginx/frontend only
```

**Stop all services:**
```bash
podman compose down
```

**Rebuild after code changes:**
```bash
podman compose up -d --build
```

**Reset database (removes all data):**
```bash
podman compose down -v
podman compose up -d --build
```

---

### Alternative: Docker

The same commands work with `docker compose` instead of `podman compose`:

```bash
cd backend
docker compose up -d --build
```

---

### Local Development (without containers)

**Backend:**
```bash
cd backend
go run cmd/api/main.go
```

Requires a running PostgreSQL instance. Configure via environment variables or a `.env` file in `backend/`.

API available at `http://localhost:8080`

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

UI available at `http://localhost:5173`

> When running the frontend locally against a containerized backend, set `VITE_API_BASE_URL=http://localhost:8080/api/v1` in `frontend/.env`.

See [backend/AUTH_SETUP.md](backend/AUTH_SETUP.md) for authentication details.

---

## Default Admin Account

On first run, register a user via `POST /api/v1/auth/register` or the `/register` page, then update the role to `admin` directly in the database or via the Staff Management page once logged in.

---

## Features

### Implemented ✅

- **Authentication & Authorization**
  - JWT-based auth with 8 RBAC roles (nurse, provider, charge_nurse, operations, consult, transport, quality_safety, admin)
  - User registration and login
  - Unit/Department visibility boundaries enforced at API layer

- **Care Team Management**
  - Per-encounter care team assignment with role types
  - Structured handoff notes (patient status, pending tasks, priorities)
  - Handoff history and transfer tracking

- **Patient Flow Tracking**
  - Immutable flow state timeline per encounter
  - Audit trail with actor, timestamp, and reason

- **Clinical Task Board**
  - Tasks scoped to encounter, patient, or unit
  - Status lifecycle: open → in_progress → completed / escalated / cancelled

- **Consult Requests**
  - Accept / decline / redirect / complete workflows
  - Role-scoped consult inbox

- **Bed Management**
  - Status tracking: available, occupied, blocked, cleaning, maintenance
  - Capability tags and bed request matching
  - Concurrent assignment protection via DB-level row locking

- **Transport Requests**
  - Full lifecycle with immutable change events
  - Unit-scoped RBAC

- **Discharge Planning**
  - Structured checklist per discharge type
  - Required vs optional items with waiver support

- **Exception & Incident Logging**
  - Immutable event log (draft → finalized → corrected)
  - Quality/Safety incident review queue

- **Administration**
  - Department and Unit management
  - Staff management — assign roles, units, and departments per user
  - Inactive user toggle (blocks login without deleting records)

---

## Project Structure

```
wardflow/
├── backend/                  # Go API server
│   ├── cmd/api/              # Application entrypoint
│   ├── internal/             # Domain packages
│   │   ├── bed/              # Bed management
│   │   ├── careteam/         # Care team & handoffs
│   │   ├── consult/          # Consult requests
│   │   ├── discharge/        # Discharge checklists
│   │   ├── encounter/        # Encounter management
│   │   ├── exception/        # Exception events
│   │   ├── flow/             # Flow state transitions
│   │   ├── handler/          # Shared HTTP handlers
│   │   ├── incident/         # Safety incidents
│   │   ├── middleware/        # Auth, RBAC, logging
│   │   ├── models/           # Shared domain models
│   │   ├── router/           # Route registration
│   │   ├── task/             # Clinical tasks
│   │   └── transport/        # Transport requests
│   ├── pkg/                  # Public packages
│   │   ├── auth/             # JWT & auth service
│   │   ├── database/         # DB connection & health
│   │   └── logger/           # Structured logging
│   ├── docker-compose.yml    # Full stack compose file
│   └── Dockerfile            # Backend container
├── frontend/                 # React SPA
│   ├── src/
│   │   ├── features/         # Domain feature modules
│   │   ├── pages/            # Route-level pages
│   │   └── shared/           # Shared UI & utilities
│   ├── Dockerfile            # Nginx container (multi-stage)
│   └── nginx.conf            # SPA fallback + API proxy
└── docs/                     # Requirements & specs
```

---

## Development

### Backend

```bash
cd backend

# Local run
go run cmd/api/main.go

# Tests
go test ./...

# Build binary
go build -o bin/wardflow-api cmd/api/main.go

# Container (with live logs)
podman compose up -d --build
podman compose logs -f backend
```

### Frontend

```bash
cd frontend

# Development server (with HMR)
npm run dev

# Type check + production build
npm run build

# Lint
npm run lint
```

---

## API Endpoints

### Authentication (`/api/v1/auth/`)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login, returns JWT |
| POST | `/api/v1/auth/logout` | Invalidate session |
| GET  | `/api/v1/auth/me` | Get current user |
| POST | `/api/v1/auth/change-password` | Change password |

### Core Resources (`/api/v1/`)
| Resource | Endpoints |
|----------|-----------|
| Encounters | `GET/POST /encounters`, `GET/PATCH /encounters/:id` |
| Care Team | `GET/POST /encounters/:id/care-team`, `PATCH /care-team/assignments/:id/transfer` |
| Tasks | `GET/POST /tasks`, `PATCH /tasks/:id` |
| Consults | `GET/POST /consults`, `PATCH /consults/:id/accept\|decline\|complete` |
| Beds | `GET/POST /beds`, `PATCH /beds/:id/status`, `POST /bed-requests` |
| Transport | `GET/POST /transport`, `PATCH /transport/:id` |
| Discharge | `POST /discharge/init`, `GET /discharge/:encounterId`, `PATCH /discharge/items/:id` |
| Incidents | `GET/POST /incidents`, `PATCH /incidents/:id/status` |
| Exceptions | `GET/POST /exceptions`, `PATCH /exceptions/:id/finalize\|correct` |

### Health
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | API + database health check |

---

## Environment Variables

### Backend
```bash
ENV=development
PORT=8080
DB_HOST=postgres
DB_PORT=5432
DB_USER=wardflow
DB_PASSWORD=wardflow_dev_password
DB_NAME=wardflow
DB_SSLMODE=disable
JWT_SECRET=your-secret-change-in-production
JWT_EXPIRATION_HOURS=24
```

### Frontend
```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

> In Docker/Podman, the frontend container proxies `/api/` and `/auth/` to the backend via nginx — `VITE_API_BASE_URL` is not needed.

---

## Documentation

- [Backend Auth Setup](backend/AUTH_SETUP.md)
- [Database Guide](backend/DATABASE.md)
- [Podman Commands](backend/PODMAN.md)
- [Frontend Setup](frontend/SETUP.md)
- [Requirements](docs/req-and-spec-pack.md)
- [Task Description](docs/task-description.md)

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.25, net/http (stdlib router), GORM |
| Database | PostgreSQL 16, JSONB, uuid-ossp |
| Auth | JWT (HS256), bcrypt |
| Frontend | React 18, TypeScript 5, Vite 8 |
| Styling | Tailwind CSS 4 |
| State | Zustand, React Query patterns |
| Containers | Podman / Docker, nginx |

---

## License

Proprietary — WardFlow Care Coordination System

---

**Status**: MVP feature-complete · All 10 modules implemented · Full-stack containerized
