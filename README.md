# WardFlow

Inpatient/ED care-coordination system designed to reduce missed actions during handoffs, improve real-time operational visibility, and provide accountable tasking with compliance-ready audit trails.

## Architecture

- **Backend**: Go 1.25+ with PostgreSQL 16
- **Frontend**: React 18+ with TypeScript and Tailwind CSS
- **Database**: PostgreSQL with JSONB support
- **Deployment**: Podman/Docker containers

## Quick Start

### Backend (Go API)

```bash
cd backend
podman compose up -d
```

API available at `http://localhost:8080`

See [backend/AUTH_SETUP.md](backend/AUTH_SETUP.md) for authentication details.

### Frontend (React)

```bash
cd frontend
npm install
npm run dev
```

UI available at `http://localhost:5173`

See [frontend/SETUP.md](frontend/SETUP.md) for development guide.

## Features

### Implemented ✅

- **Authentication & Authorization**
  - JWT-based auth with 8 RBAC roles
  - User registration and login
  - Password management
  - Unit/Department visibility boundaries
  
- **Database Layer**
  - PostgreSQL connection with GORM
  - Connection pooling
  - Health checks
  - Audit-ready schema

- **Frontend Foundation**
  - React 18 with TypeScript
  - Tailwind CSS styling
  - Vite build system
  - Modern component architecture

### Planned 🚧

1. Care team assignment per encounter
2. Patient flow tracking
3. Clinical task board
4. Inter-department consult requests
5. Bed management
6. Transport requests
7. Discharge planning checklist
8. Exception workflows
9. Daily huddle dashboard
10. Quality/safety incident logging

## Project Structure

```
wardflow/
├── backend/          # Go API server
│   ├── cmd/api/      # Application entrypoint
│   ├── internal/     # Private packages
│   │   ├── config/   # Configuration
│   │   ├── handler/  # HTTP handlers
│   │   ├── middleware/ # Auth, RBAC, logging
│   │   ├── models/   # Domain models
│   │   └── router/   # Route setup
│   └── pkg/          # Public packages
│       ├── auth/     # JWT & auth service
│       ├── database/ # DB connection
│       └── logger/   # Logging utilities
├── frontend/         # React UI
│   ├── src/
│   │   ├── App.tsx   # Root component
│   │   ├── main.tsx  # Entry point
│   │   └── index.css # Tailwind styles
│   └── public/       # Static assets
└── docs/             # Requirements & specs
```

## Documentation

- [Backend Setup](backend/README.md)
- [Database Guide](backend/DATABASE.md)
- [Auth Implementation](backend/AUTH_SETUP.md)
- [Podman Commands](backend/PODMAN.md)
- [Frontend Setup](frontend/SETUP.md)
- [API Instructions](.github/copilot-instructions.md)
- [Requirements](docs/req-and-spec-pack.md)
- [Task Description](docs/task-description.md)

## Tech Stack

### Backend
- Go 1.25
- PostgreSQL 16
- GORM ORM
- JWT authentication
- bcrypt password hashing
- Podman containers

### Frontend
- React 18.3
- TypeScript 5.x
- Tailwind CSS 4.2
- Vite 8.x
- Modern hooks & functional components

## Development

### Backend Development

```bash
cd backend

# Local development
go run cmd/api/main.go

# With containers
podman compose up -d
podman logs -f wardflow-backend

# Run tests
go test ./...

# Build
go build -o bin/wardflow-api cmd/api/main.go
```

### Frontend Development

```bash
cd frontend

# Development server
npm run dev

# Production build
npm run build

# Preview production
npm run preview

# Lint
npm run lint
```

## API Endpoints

### Authentication
- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `POST /auth/logout` - User logout
- `GET /auth/me` - Get current user
- `POST /auth/change-password` - Change password

### Health
- `GET /health` - API health check

See [AUTH_SETUP.md](backend/AUTH_SETUP.md) for detailed API documentation.

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
JWT_SECRET=your-secret-key
```

### Frontend
```bash
VITE_API_URL=http://localhost:8080
```

## Contributing

1. Follow documented patterns in `.github/copilot-instructions.md`
2. Check `/docs` for architecture requirements
3. Use backend/frontend agent configurations
4. Maintain >80% test coverage for core logic
5. Follow Go best practices (backend)
6. Follow React best practices (frontend)

## License

Proprietary - WardFlow Care Coordination System

---

**Status**: Foundation complete - Backend API with auth + Frontend scaffold ready  
**Next**: Implement care team assignment and patient flow tracking
