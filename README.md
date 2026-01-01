# Project Manager API

Go REST API for managing projects and project-scoped tasks, built with reliability/operability basics (request IDs, logging, panic recovery, graceful shutdown, health endpoints). Supports MemoryStore (default) and Postgres (via `DATABASE_URL`).

## Quickstart (Docker Compose)

Starts Postgres → runs migrations (one-shot) → starts API:

```bash
docker compose up --build
```

### Endpoints:

GET http://localhost:4000/livez

GET http://localhost:4000/readyz

### Stop:

```bash
docker compose down
```

### Reset DB (delete volume):

docker compose down -v

### Example Requests

Create a project:

curl -i -X POST http://localhost:4000/v1/projects \
 -H 'Content-Type: application/json' \
 -d '{"name":"Alpha"}'

List projects:

curl -i "http://localhost:4000/v1/projects?page=1&page_size=20"

Create a task:

curl -i -X POST http://localhost:4000/v1/projects/<projectId>/tasks \
 -H 'Content-Type: application/json' \
 -d '{"title":"First task","description":"Ship it"}'

Update a task (PATCH):

curl -i -X PATCH http://localhost:4000/v1/projects/<projectId>/tasks/<taskId> \
 -H 'Content-Type: application/json' \
 -d '{"status":"doing"}'

Local Run (no Docker)

MemoryStore (default):

go run ./cmd/api

Postgres:

export DATABASE_URL="postgres://pm:pm@localhost:5432/pm?sslmode=disable"
go run ./cmd/api

Config

ADDR (default :4000)

DATABASE_URL (set => Postgres, empty => MemoryStore)

SHUTDOWN_TIMEOUT (default 10s)

Tests
go test ./... -count=1

Notes

Migrations are embedded and run by /app/migrate (compose migrate service).

Image runs as non-root (least privilege).

sqlc generated code is committed; regenerate with sqlc generate.
