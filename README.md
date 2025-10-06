# GPCI Compro API

Opinionated Go boilerplate for building HTTP + worker services with PostgreSQL, Redis, and Asynq. The project ships with a Fiber HTTP server, structured logging, environment-driven configuration, embedded Goose migrations, and a background queue processor.

## Features
- Fiber-based API with ready-to-wire handler layer and example routes.
- Structured logging via `slog` with environment-aware formatting.
- Centralized configuration loader (`internal/config`) covering app, database, Redis, and Asynq settings.
- PostgreSQL utilities (`internal/db`) with DSN helpers and connection pooling.
- Redis/Asynq queue toolkit with task, worker, and scheduler scaffolding (`internal/queue`).
- Docker image + compose stack for the API, worker, and one-shot migrator (bring your own Postgres/Redis).

## Prerequisites
- Go 1.25 or newer.
- Running PostgreSQL instance reachable with the credentials you set in `.env`.
- Running Redis instance (required for the worker / task queue).
- Optional: Docker & Docker Compose v2 if you prefer containerized workflows.

## Getting Started
1. Copy the sample environment file and adjust values to your setup:
   ```bash
   cp .env.example .env
   ```
2. Ensure your local Postgres and Redis services are running and match the connection info in the `.env`.
3. Download dependencies (first run only):
   ```bash
   go mod download
   ```

### Run the API locally
```bash
go run ./cmd/api
```
The API listens on `APP_PORT` (default `8080`) and exposes:
- `GET /` – empty 204 to indicate the service is up.
- `GET /api/ping` – returns `{ "pong": true }`.
- `GET /api/health` – basic health endpoint (registered via the router boilerplate).
- Example handlers in `internal/http/handlers` can be enabled and wired for custom logic.

### Run the background worker
```bash
go run ./cmd/worker
```
The worker connects to Redis and processes tasks registered in `internal/queue`. An example `notify:user` task logs its payload; extend `NotifyUserHandler` with your integration (email, push, etc.).

### Docker Compose workflow
The compose stack focuses on the application layers only:
```bash
docker compose up api worker
```
- `api` and `worker` services share the same build context.
- A `migrator` profile is available for one-off migrations:
  ```bash
  docker compose --profile migrate run --rm migrator
  ```
Bring your own Postgres/Redis services and expose them according to the `.env` connection strings.

## Database Migrations
- SQL files placed under `migrations/` are embedded and executed on API startup via Goose.
- To run migrations without starting the API, use the dedicated migrator container (see compose instructions) or add your own CLI flags/commands.

## Configuration Reference
Environment variables are parsed by `internal/config` and include sensible defaults:

| Section | Keys |
| --- | --- |
| App | `APP_NAME`, `APP_ENV`, `APP_PORT`, `APP_READ_TIMEOUT`, `APP_WRITE_TIMEOUT`, `APP_IDLE_TIMEOUT`, `SHUTDOWN_TIMEOUT` |
| Database | `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME` |
| Redis | `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB` |
| Asynq | `ASYNQ_CONCURRENCY`, `ASYNQ_QUEUE_DEFAULT`, `ASYNQ_QUEUE_CRITICAL` |

Adjust these values in `.env` for each environment (local, staging, production).

## Project Structure
```text
.
├── cmd/                # Entrypoints for the API and worker binaries
├── internal/config/    # Environment-first configuration loader
├── internal/db/        # Database connection helpers and transaction utilities
├── internal/http/      # Router + handler scaffolding
├── internal/queue/     # Asynq tasks, worker, and scheduler
├── internal/utils/     # Shared utilities (logger, etc.)
├── migrations/         # SQL migrations executed via Goose
├── Dockerfile          # Builder/runtime image for api + worker
└── docker-compose.yml  # API/worker/migrator services (without DB/Redis)
```

## Testing
```bash
go test ./...
```
Include integration tests as you flesh out repositories and services. For tests needing Postgres or Redis, point the environment variables at dedicated test instances.

## Next Steps
- Wire your domain modules under `internal/modules`.
- Flesh out handlers and repositories, enabling the commented examples.
- Add CI/CD scripts or workflows tailored to your deployment pipeline.
