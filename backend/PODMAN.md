# WardFlow Backend - Podman Environment

## Quick Start

```bash
# Start all services (PostgreSQL + Backend)
podman compose up -d

# Stop all services
podman compose down

# View logs
podman logs wardflow-backend
podman logs wardflow-postgres

# Follow logs in real-time
podman logs -f wardflow-backend

# Rebuild after code changes
podman compose build backend
podman compose up -d --force-recreate backend

# Complete rebuild (fresh start)
podman compose down -v  # Warning: removes database data
podman compose build --no-cache
podman compose up -d
```

## Container Status

```bash
# List running containers
podman compose ps

# Check container health
podman healthcheck run wardflow-backend
podman healthcheck run wardflow-postgres

# Inspect container details
podman inspect wardflow-backend
```

## Database Management

```bash
# Connect to PostgreSQL
podman exec -it wardflow-postgres psql -U wardflow -d wardflow

# Run SQL file
podman exec -i wardflow-postgres psql -U wardflow -d wardflow < migrations/001_init.sql

# Backup database
podman exec wardflow-postgres pg_dump -U wardflow wardflow > backup.sql

# Restore database
podman exec -i wardflow-postgres psql -U wardflow -d wardflow < backup.sql

# Check database size
podman exec wardflow-postgres psql -U wardflow -d wardflow -c "\l+"
```

## Testing API

```bash
# Health check
curl http://localhost:8080/health

# With formatting
curl -s http://localhost:8080/health | jq .
```

## Troubleshooting

```bash
# View recent logs
podman logs --tail 50 wardflow-backend
podman logs --tail 50 wardflow-postgres

# Check container resource usage
podman stats wardflow-backend wardflow-postgres

# Restart specific service
podman compose restart backend
podman compose restart postgres

# Shell into backend container (if needed for debugging)
podman exec -it wardflow-backend /bin/sh

# Shell into PostgreSQL container
podman exec -it wardflow-postgres /bin/sh
```

## Network

The services communicate via the `wardflow-network` bridge network:
- Backend → PostgreSQL: `postgres:5432`
- Host → Backend: `localhost:8080`
- Host → PostgreSQL: `localhost:5432`

## Volumes

- `backend_postgres_data`: Persistent PostgreSQL data storage

To reset database completely:
```bash
podman compose down -v  # Warning: deletes all data!
podman compose up -d
```

## Environment Variables

Environment variables are configured in `docker-compose.yml`. To override:

1. **Create .env file** (for local development):
```bash
cp .env.example .env
# Edit .env with your values
```

2. **Or modify docker-compose.yml** directly for the environment you need.

## Production Considerations

For production deployment:
1. Set `ENV=production` in docker-compose.yml
2. Use strong `DB_PASSWORD` and `JWT_SECRET`
3. Enable SSL: `DB_SSLMODE=require`
4. Set up proper volume backups
5. Configure resource limits in docker-compose.yml
6. Use reverse proxy (nginx/traefik) for HTTPS termination
