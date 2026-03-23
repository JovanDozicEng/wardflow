# WardFlow Backend

Go backend service for the WardFlow inpatient/ED care-coordination system.

## Prerequisites

- Go 1.22 or later
- PostgreSQL 14+ (or use Podman/Docker for containerized setup)
- Podman & Podman Compose (recommended) or Docker & Docker Compose

## Getting Started

### Containerized Development (Recommended)

**Using Podman:**
```bash
# Start all services (backend + PostgreSQL)
podman compose up -d

# View logs
podman logs -f wardflow-backend

# Stop all services
podman compose down
```

See [PODMAN.md](./PODMAN.md) for complete Podman command reference.

### Local Development (Without Containers)

1. **Set up environment variables:**
```bash
cp .env.example .env
# Edit .env with your configuration
```

2. **Run the application:**
```bash
go run cmd/api/main.go
```

3. **Build the application:**
```bash
go build -o bin/wardflow-api cmd/api/main.go
```

### Project Structure

```
backend/
├── cmd/
│   └── api/              # Application entrypoint
├── internal/             # Private application code
│   ├── config/           # Configuration management
│   ├── handler/          # HTTP handlers (controllers)
│   ├── middleware/       # HTTP middleware (auth, logging, etc.)
│   ├── models/           # Domain models and DTOs
│   ├── repository/       # Data access layer
│   └── service/          # Business logic layer
└── pkg/                  # Public packages
    └── logger/           # Logging utilities
```

## Development Commands

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Format code
go fmt ./...

# Run linter (requires golangci-lint)
golangci-lint run

# Build for production
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/wardflow-api cmd/api/main.go
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ENV` | Environment (development/production) | `development` |
| `PORT` | HTTP server port | `8080` |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `wardflow` |
| `DB_PASSWORD` | PostgreSQL password | `` |
| `DB_NAME` | PostgreSQL database name | `wardflow` |
| `DB_SSLMODE` | PostgreSQL SSL mode | `disable` |
| `JWT_SECRET` | JWT signing secret | `` |
| `JWT_EXPIRATION_HOURS` | JWT token expiration | `24` |

## Testing

The project follows table-driven testing patterns:

```go
func TestExample(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        // test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Architecture

- **Clean Architecture**: Separation of concerns with clear boundaries between layers
- **Repository Pattern**: Data access abstraction
- **Dependency Injection**: Interface-based design for testability
- **Context-Aware**: All operations support context cancellation
- **Audit-First**: Immutable event logging for compliance

See `/docs` for detailed architecture documentation.
