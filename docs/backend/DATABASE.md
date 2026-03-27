# Database Connection

## Overview

WardFlow uses PostgreSQL 16+ with GORM as the ORM. The database layer is designed with:

- **Connection pooling** for optimal performance
- **Context-aware operations** for timeout control
- **Transaction support** with automatic commit/rollback
- **Health checks** for monitoring
- **Audit-first design** following the documented architecture

## Connection Configuration

### Environment Variables

```bash
# Database Host & Credentials
DB_HOST=postgres                    # Database host
DB_PORT=5432                        # Database port
DB_USER=wardflow                    # Database user
DB_PASSWORD=wardflow_dev_password   # Database password
DB_NAME=wardflow                    # Database name
DB_SSLMODE=disable                  # SSL mode (require in production)

# Connection Pool Settings
DB_MAX_OPEN_CONNS=25               # Maximum open connections
DB_MAX_IDLE_CONNS=10               # Maximum idle connections
DB_CONN_MAX_LIFETIME_MINUTES=5     # Connection max lifetime
```

### Production Recommendations

- Set `DB_SSLMODE=require` for encrypted connections
- Adjust `DB_MAX_OPEN_CONNS` based on workload (typically 25-100)
- Keep `DB_MAX_IDLE_CONNS` at ~40% of max open connections
- Set `DB_CONN_MAX_LIFETIME_MINUTES` to 5-15 minutes

## Usage Examples

### Basic Connection

```go
import "github.com/wardflow/backend/pkg/database"

// Create database config from app config
dbCfg := &database.Config{
    Host:            cfg.DBHost,
    Port:            cfg.DBPort,
    User:            cfg.DBUser,
    Password:        cfg.DBPassword,
    DBName:          cfg.DBName,
    SSLMode:         cfg.DBSSLMode,
    MaxOpenConns:    cfg.DBMaxOpenConns,
    MaxIdleConns:    cfg.DBMaxIdleConns,
    ConnMaxLifetime: time.Duration(cfg.DBConnMaxLifetime) * time.Minute,
    LogLevel:        cfg.LogLevel,
}

// Connect
db, err := database.Connect(dbCfg)
if err != nil {
    return fmt.Errorf("failed to connect: %w", err)
}
defer db.Close()
```

### Using Context

```go
ctx := context.Background()

// Ping with timeout
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

if err := db.Ping(ctx); err != nil {
    return fmt.Errorf("ping failed: %w", err)
}

// Query with context
var count int64
err := db.WithContext(ctx).Model(&User{}).Count(&count).Error
```

### Transactions

```go
ctx := context.Background()

err := db.Transaction(ctx, func(tx *database.DB) error {
    // Create user
    user := &User{Name: "John Doe"}
    if err := tx.Create(user).Error; err != nil {
        return err // Automatic rollback
    }
    
    // Create related record
    profile := &Profile{UserID: user.ID}
    if err := tx.Create(profile).Error; err != nil {
        return err // Automatic rollback
    }
    
    return nil // Automatic commit
})
```

### Health Checks

```go
ctx := context.Background()

health, err := db.HealthCheck(ctx)
if err != nil {
    logger.Error("database unhealthy: %v", err)
}

// health contains:
// - status: "healthy" or "unhealthy"
// - open_connections: current open connections
// - in_use: connections in use
// - idle: idle connections
// - max_open_conns: configured maximum
```

## GORM Features

### Prepared Statements

Enabled by default for better performance:
```go
db.DB.PrepareStmt = true
```

### Skip Default Transactions

Read operations don't use transactions by default:
```go
db.DB.SkipDefaultTransaction = true
```

### UTC Timestamps

All timestamps use UTC:
```go
db.DB.NowFunc = func() time.Time {
    return time.Now().UTC()
}
```

## Common Patterns

### Repository Pattern

```go
type UserRepository struct {
    db *database.DB
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
    if err != nil {
        return nil, fmt.Errorf("find user: %w", err)
    }
    return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    return r.db.WithContext(ctx).Create(user).Error
}
```

### Pagination

```go
func (r *Repository) List(ctx context.Context, limit, offset int) ([]Model, int64, error) {
    var items []Model
    var total int64
    
    db := r.db.WithContext(ctx).Model(&Model{})
    
    // Get total count
    if err := db.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // Get paginated results
    if err := db.Limit(limit).Offset(offset).Find(&items).Error; err != nil {
        return nil, 0, err
    }
    
    return items, total, nil
}
```

### Error Handling

```go
import "errors"
import "gorm.io/gorm"

result := db.First(&user, id)
if result.Error != nil {
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        return nil, ErrNotFound
    }
    return nil, fmt.Errorf("query failed: %w", result.Error)
}
```

## Testing

Database tests are marked with `t.Skip()` by default. To run integration tests:

1. Start PostgreSQL test instance
2. Remove `t.Skip()` from test functions
3. Run tests: `go test ./pkg/database/...`

## Troubleshooting

### Connection Issues

```bash
# Test connection from host
podman exec wardflow-postgres psql -U wardflow -d wardflow -c "SELECT 1;"

# Check logs
podman logs wardflow-backend | grep database
podman logs wardflow-postgres | tail -20
```

### Pool Exhaustion

If you see "too many connections" errors:
- Increase `DB_MAX_OPEN_CONNS`
- Check for connection leaks (always use context, close cursors)
- Monitor connection pool with health endpoint

### Slow Queries

Enable query logging in development:
```bash
LOG_LEVEL=debug
```

## Next Steps

1. **Migrations**: Set up schema migration system
2. **Models**: Define domain models with GORM tags
3. **Repositories**: Implement data access layer
4. **Audit Tables**: Create audit log tables per docs
5. **Indexes**: Add appropriate indexes for performance

See `/docs/req-and-spec-pack.md` for data model specifications.
