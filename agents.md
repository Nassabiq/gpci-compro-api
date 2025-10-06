# agents.md — Codex untuk Aplikasi Golang + Fiber

Dokumen ini adalah “single source of truth” untuk standar, konvensi, dan praktek terbaik pembangunan layanan berbasis **Go + Fiber v2** dengan Postgres (tanpa ORM), SQLC, Goose, Redis/RabbitMQ, Docker, dan CI. Tujuan: kode konsisten, mudah di-debug, dan siap produksi.

---

## 1) Prinsip Desain

- **Sederhana dulu**: minim dependensi, tanpa ORM; gunakan `database/sql` + **sqlc**.
- **Explicit over magic**: dependency injection manual, error bubble-up dengan konteks.
- **12-factor**: konfigurasi dari environment; stateless; logs ke stdout.
- **Secure by default**: sane defaults untuk CORS, rate limit, TLS termination, input validation.
- **Observability first**: struktur log, trace id korelatif, metrics siap scrape.
- **Testable**: business logic bebas dari Fiber context; mockable interface.

---


---

## 2) Struktur Proyek (Revisi: Modular di `internal/modules`)

> Kita pakai **modular monolith**: setiap fitur hidup di `internal/modules/<nama-module>` dan sudah mengemas **repository**, **service/usecase**, serta **kontrak data** (**payload/response/dto**) di dalamnya. Handler HTTP berada di `internal/http/handler/<module>` agar routing tetap centralized. Dengan pola ini, modul bisa diekstrak menjadi microservice jika kelak diperlukan.

```
.
├─ cmd/
│  └─ api/
│     └─ main.go
├─ internal/
│  ├─ app/                    # wiring root: fiber app, routes, middlewares, DI container
│  │  ├─ router.go            # registrasi route per module
│  │  └─ container.go         # inisialisasi shared deps (db, redis, mq, logger)
│  ├─ config/                 # load env, defaults, validation
│  ├─ http/
│  │  ├─ middleware/          # auth, rbac, cors, rate-limit, recover, request-id
│  │  ├─ handler/
│  │  │  ├─ auth/             # handler auth (contoh shared)
│  │  │  └─ products/         # handler untuk module products (tipis: delegasi ke service)
│  │  └─ response/            # uniform response & error encoder
│  ├─ modules/
│  │  ├─ products/
│  │  │  ├─ domain/           # entity (core business types)
│  │  │  │  ├─ dto.go         # DTO internal antar-layer (opsional)
│  │  │  │  ├─ payload.go     # request models (validasi) — dipakai handler->service
│  │  │  │  └─ response.go    # response models (API-safe)
│  │  │  ├─ repo/             # akses data (sqlc/redis/dll)
│  │  │  │  ├─ postgres/      # sqlc generated + adapter
│  │  │  │  └─ cache/         # redis cache spesifik module
│  │  │  ├─ service/          # usecases (business rules), interface + impl
│  │  │  ├─ wiring.go         # provider untuk DI (construct repo+service)
│  │  │  └─ README.md         # kontrak module singkat
│  │  └─ (module-lain)...
│  ├─ mq/                     # rabbitmq producer/consumer umum (atau per module jika spesifik)
│  ├─ jobs/                   # background workers, schedulers
│  ├─ telemetry/              # logger, tracing, metrics
│  ├─ task/                   # CLI tasks: migrate, seed, repair, backfill
│  └─ util/                   # helper generic (uuid, time, hash, crypto)
├─ db/
│  ├─ schema/                 # goose migrations (up/down)
│  └─ queries/                # sqlc queries dipecah per module
│     ├─ products/            # queries khusus module products
│     └─ (module-lain)...
├─ pkg/                       # lib reusable (opsional)
├─ sqlc.yaml
├─ docker/
│  ├─ Dockerfile
│  └─ docker-compose.yml
├─ Makefile
└─ .github/workflows/ci.yml
```

### Catatan Implementasi
- **sqlc output per module**: hasil generate diarahkan ke `internal/modules/<module>/repo/postgres` agar coupling tetap rendah.
- **Contract data**: gunakan `domain/payload.go` untuk **request** (payload), `domain/response.go` untuk **response** (hanya field yang boleh keluar), dan `domain/dto.go` untuk struktur internal antar-layer.
- **Handler tipis**: parsing + validasi → panggil `service` → encode response via `internal/http/response`.
- **Wiring modul**: `internal/modules/<module>/wiring.go` expose constructor `Provide(container *app.Container) *ModuleDeps` yang mengembalikan service yang siap dipakai handler.
- **Pengujian**: test service tanpa Fiber; mock interface repo di dalam module.

---

### Contoh Mini: Module `products`

**1) Domain (payload/response/dto)**
```go
// internal/modules/products/domain/payload.go
package domain

type CreateProductPayload struct {
  Name        string  `json:"name" validate:"required,min=2"`
  Description *string `json:"description,omitempty"`
  Price       int64   `json:"price" validate:"gte=0"`
}

// internal/modules/products/domain/response.go
package domain

type ProductResponse struct {
  ID          string  `json:"id"`
  Name        string  `json:"name"`
  Description *string `json:"description,omitempty"`
  Price       int64   `json:"price"`
  CreatedAt   string  `json:"created_at"`
}

// internal/modules/products/domain/dto.go
package domain

type Product struct { // entity (internal)
  ID          string
  Name        string
  Description *string
  Price       int64
  CreatedAt   int64 // unix or time.Time sesuai preferensi
}
```

**2) Repository (sqlc + adapter)**
```
db/queries/products/
  products.sql     # GetByID, List, Insert, Update, Delete
```
```sql
-- name: CreateProduct :one
INSERT INTO products (id, name, description, price)
VALUES ($1, $2, $3, $4)
RETURNING id, name, description, price, created_at;
```
```go
// internal/modules/products/repo/postgres/repo.go
package postgres

type Repo interface {
  Create(ctx context.Context, p domain.Product) (domain.Product, error)
  Get(ctx context.Context, id string) (domain.Product, error)
  List(ctx context.Context, page, size int, sort string) ([]domain.Product, int64, error)
}
```

**3) Service (usecase)**  
`service/service.go` berisi interface `Service` + implementasi yang konsumsi `Repo`:
```go
type Service interface {
  Create(ctx context.Context, in domain.CreateProductPayload) (domain.ProductResponse, error)
  Find(ctx context.Context, id string) (domain.ProductResponse, error)
  List(ctx context.Context, q ListQuery) (Paged[domain.ProductResponse], error)
}
```

**4) Handler**  
`internal/http/handler/products/handler.go`:
```go
func (h *Handler) Create(c *fiber.Ctx) error {
  var in domain.CreateProductPayload
  if err := c.BodyParser(&in); err != nil { return response.BadRequest(c, err) }
  if err := h.Validate.Struct(in); err != nil { return response.Validation(c, err) }
  out, err := h.Svc.Create(c.Context(), in)
  if err != nil { return response.MapError(c, err) }
  return response.Created(c, out)
}
```

**5) Router**  
`internal/app/router.go`:
```go
func RegisterProductRoutes(r fiber.Router, h *products.Handler) {
  g := r.Group("/products")
  g.Post("/", h.Create)      // POST /api/v1/products
  g.Get("/", h.List)         // GET  /api/v1/products
  g.Get("/:id", h.Detail)    // GET  /api/v1/products/:id
  g.Put("/:id", h.Update)    // PUT  /api/v1/products/:id
  g.Delete("/:id", h.Delete) // DEL  /api/v1/products/:id
}
```

**6) sqlc.yaml (per-module output)**
```yaml
version: "2"
sql:
  - engine: "postgresql"
    schema: "db/schema"
    queries: "db/queries/products"
    gen:
      go:
        package: "postgres"                   # package name untuk module repo
        out: "internal/modules/products/repo/postgres"
        emit_json_tags: true
        overrides:
          - db_type: "citext"
            go_type: "string"
```

Dengan setup ini, modul **mandiri**: ketika menambah module baru cukup meniru struktur `products` dan registrasi route di `router.go`.


## 3) Konfigurasi & Environment

**Variabel penting (contoh `.env`):**
```
APP_NAME=gpci-api
APP_ENV=production
APP_ADDR=:8080

DB_DSN=postgres://user:pass@postgres:5432/app?sslmode=disable
DB_MAX_OPEN=30
DB_MAX_IDLE=10

REDIS_ADDR=redis:6379
REDIS_DB=0

RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/

JWT_SIGNING_KEY=change-me
JWT_TTL=15m
JWT_REFRESH_TTL=720h

CORS_ALLOW_ORIGINS=https://example.com
RATE_LIMIT_RPS=10
```

**Loader:**
- Gunakan `github.com/caarlos0/env/v11` atau manual `os.LookupEnv`.
- Validasi boundary (misal TTL minimal 1m, RPS 1–1000). Fail fast bila invalid.

---

## 4) Logging, Tracing, Metrics

- **Logger**: Zerolog/Zap, JSON line; field wajib: `ts`, `level`, `msg`, `req_id`, `span_id`, `trace_id`, `remote_ip`, `path`, `latency_ms`.
- **Trace**: OpenTelemetry SDK; propagasi `traceparent` dari header.
- **Metrics**: Prometheus (`/metrics`). Gunakan counter untuk 2xx/4xx/5xx, histogram durasi, gauge pool DB.

Middleware Fiber:
- `RequestID` → simpan di context.
- `Recover` → log panic + 500 JSON aman.
- `Logger` (kustom) → uniform fields.
- `RateLimiter` → leaky bucket berbasis memory/redis.
- `CORS` → whitelist dari env.

---

## 5) Kontrak HTTP (API Guidelines)

**Standar Path & Versi**
- Prefix: `/api/v1`.
- Koleksi: `/api/v1/users`, item: `/api/v1/users/:id`.

**Query & Pagination**
- `page`, `page_size` (default 1 & 20, maks 100).
- Sorting: `sort=created_at:desc,name:asc`.
- Filtering: `filter[field]=value` (multi ok).

**Request/Response**
- `Content-Type: application/json; charset=utf-8`.
- Response sukses:
```json
{
  "data": {...},
  "meta": {"page":1,"page_size":20,"total":132,"sort":["created_at:desc"]},
  "trace_id": "d3b07..."
}
```

**Error model (uniform)**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "invalid input",
    "details": [{"field":"email","reason":"must be email"}]
  },
  "trace_id": "d3b07..."
}
```
Kode `error.code` (stabil): `BAD_REQUEST`, `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `RATE_LIMITED`, `INTERNAL`, `VALIDATION_ERROR`.

---

## 6) Validasi & Sanitasi

- Gunakan `go-playground/validator/v10` untuk DTO.
- Pisahkan **DTO** (request), **Entity** (domain), **Model** (db/sqlc).
- Normalisasi input (trim, lowercasing sesuai field).
- Batasi payload size (mis. 1MB) via `app.Server().MaxRequestBodySize`.

---

## 7) Domain: Auth & RBAC

**Fitur minimal:**
- Login email+password → **access token (JWT)** + **refresh token**.
- Refresh endpoint → rotate refresh (token jangka panjang disimpan/blacklist).
- Logout → revoke refresh token.
- RBAC: roles, permissions; cek via middleware `Require(“perm:users.read”)`.

**Skema ringkas (Postgres):**
```sql
-- db/schema/0001_init.up.sql
CREATE TABLE users (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email         citext UNIQUE NOT NULL,
  password_hash text   NOT NULL,
  name          text   NOT NULL,
  is_active     bool   NOT NULL DEFAULT true,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE roles (
  id serial PRIMARY KEY,
  code text UNIQUE NOT NULL,
  name text NOT NULL
);

CREATE TABLE permissions (
  id serial PRIMARY KEY,
  code text UNIQUE NOT NULL,
  name text NOT NULL
);

CREATE TABLE user_roles (
  user_id uuid REFERENCES users(id) ON DELETE CASCADE,
  role_id int  REFERENCES roles(id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, role_id)
);

CREATE TABLE role_permissions (
  role_id int REFERENCES roles(id) ON DELETE CASCADE,
  perm_id int REFERENCES permissions(id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, perm_id)
);

-- index tambahan dan trigger updated_at disarankan
```

**JWT**
- Alg: HS256, claim minimal: `sub`, `exp`, `iat`, `roles`, `perms`.
- TTL access (mis. 15m), refresh (mis. 30–90 hari).
- Simpan refresh token (hash) di DB/Redis (revocation & rotation).

**Middleware contoh:**
- `AuthJWT`: verifikasi, inject subject/roles ke context.
- `RBAC`: cek `perms` atau union dari roles→permissions.

---

## 8) Data Access: sqlc + Goose

**Goose**
- Folder: `db/schema`.
- Konvensi file: `0001_init.up.sql`, `0001_init.down.sql`.
- Perintah:
  - `goose -dir db/schema postgres "$DB_DSN" up`
  - `make migrate-up`, `make migrate-down`.

**SQLC**
- Folder queries: `db/queries/*.sql`.
- `sqlc.yaml` (ringkas):
```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "db/queries"
    schema: "db/schema"
    gen:
      go:
        package: "postgres"
        out: "internal/repo/postgres"
        emit_json_tags: true
        overrides:
          - db_type: "citext"
            go_type: "string"
```

**Contoh query (named):**
```sql
-- db/queries/users.sql
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, name)
VALUES ($1, $2, $3)
RETURNING *;
```

**Transaksi**
- Bungkus use-case yang butuh konsistensi (multi-repo) dengan helper:
```go
func WithTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
  tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
  if err != nil { return err }
  if err := fn(tx); err != nil { tx.Rollback(); return err }
  return tx.Commit()
}
```

---

## 9) Handler Tipis, Service Gemuk

**Handler (Fiber):** parse/validate request → panggil service → encode response.  
**Service (domain):** business logic → panggil repo/cache/mq.

```go
// internal/http/handler/auth.go
func (h *AuthHandler) Login(c *fiber.Ctx) error {
  var req LoginRequest
  if err := c.BodyParser(&req); err != nil { return response.BadRequest(c, err) }
  if err := h.v.Struct(req); err != nil { return response.Validation(c, err) }
  res, err := h.svc.Login(c.Context(), req.Email, req.Password, requestID(c))
  if err != nil { return response.MapError(c, err) }
  return response.OK(c, res)
}
```

---

## 10) Background Jobs & MQ

- **RabbitMQ** untuk event driven & antrian berat (email, audit trail, export).
- **Outbox pattern** (opsional) untuk memastikan *at least once* saat publish dari transaksi DB.
- Consumer worker proses di `internal/jobs` (graceful shutdown, prefetch, retry dengan DLX).
- **Redis** dapat dipakai untuk cache, rate limit, dan ephemeral locks.

---

## 11) Keamanan

- Hash password: `bcrypt` (cost sesuai perf), tambah pepper opsional.
- CSRF: Tidak perlu untuk pure API; untuk panel admin gunakan token.
- CORS: ketat; `Allow-Credentials` hanya jika perlu, domain whitelist.
- Rate limit: per IP + per user (akses sensitif).
- Sensitive logging: **jangan log** password/token/PII.
- Headers: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`.
- TLS di terminasi reverse proxy (Nginx/Traefik).

---

## 12) Performa

- Pool DB: set `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`.
- Gunakan `COPY`/batch ketika import besar.
- Cache read-heavy (Redis) dengan **cache aside** dan TTL.
- Hindari alokasi berlebih: reuse buffers, pre-size slice, gunakan `WriteString`.

---

## 13) Testing

- **Unit**: service & util (tanpa Fiber) → gunakan interface repo (mocks).
- **Integration**: spin-up Postgres/RabbitMQ dengan Docker Compose.
- **HTTP tests**: Fiber `httptest`, assert status & body.
- Fixture: migrasi up → seed → jalankan test → migrasi down.

---

## 14) Makefile (target inti)

```makefile
.PHONY: dev build test lint migrate-up migrate-down sqlc

dev:         ## run hot-reload (air optional) 
	air

build:       ## build binary
	go build -o bin/api ./cmd/api

test:
	go test ./... -count=1

lint:
	golangci-lint run

migrate-up:
	goose -dir db/schema postgres "$$DB_DSN" up

migrate-down:
	goose -dir db/schema postgres "$$DB_DSN" down

sqlc:
	sqlc generate
```

---

## 15) Docker & Compose (Produksi)

**Dockerfile (multi-stage ringkas):**
```dockerfile
# builder
FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/api ./cmd/api

# runtime
FROM gcr.io/distroless/base-debian12
ENV GIN_MODE=release
WORKDIR /app
COPY --from=build /out/api /app/api
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/api"]
```

**docker-compose.yml (ops ringkas):**
```yaml
version: "3.9"
services:
  api:
    build: .
    environment:
      - APP_ENV=production
      - APP_ADDR=:8080
      - DB_DSN=postgres://user:pass@postgres:5432/app?sslmode=disable
      - REDIS_ADDR=redis:6379
      - RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
      - JWT_SIGNING_KEY=${JWT_SIGNING_KEY}
    depends_on: [postgres, redis, rabbitmq]
    ports: ["8080:8080"]
    restart: unless-stopped

  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: app
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    volumes: [pgdata:/var/lib/postgresql/data]
    healthcheck: { test: ["CMD-SHELL","pg_isready -U user"], interval: 5s, timeout: 5s, retries: 10 }

  redis:
    image: redis:7
    command: ["redis-server","--appendonly","yes"]
    volumes: [redisdata:/data]

  rabbitmq:
    image: rabbitmq:3-management
    ports: ["15672:15672"]
    healthcheck: { test: ["CMD","rabbitmq-diagnostics","ping"], interval: 10s, timeout: 5s, retries: 10 }

volumes:
  pgdata:
  redisdata:
```

**Bootstrapping:**
1. `docker compose up -d postgres redis rabbitmq`
2. `make migrate-up`
3. `docker compose up -d api`

---

## 16) CI (GitHub Actions)

- Matrix: lint, test, build.
- Cache `~/.cache/go-build` dan `~/go/pkg/mod`.
- Push image ke registry (opsional) dengan tags `sha` & `semver`.

```yaml
name: ci
on: [push, pull_request]
jobs:
  build-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }
      - name: Cache
        uses: actions/cache@v4
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
      - run: go mod download
      - run: golangci-lint run
      - run: go test ./... -v
  docker:
    needs: build-test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKERHUB_USER }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}
      - uses: docker/build-push-action@v6
        with:
          push: true
          tags: yourrepo/gpci-api:latest, yourrepo/gpci-api:${{ github.sha }}
```

---

## 17) PR Checklist (Wajib)

- [ ] Endpoint baru terdokumentasi (path, body, response, error).
- [ ] Validasi & error mapping sesuai codex.
- [ ] Unit test untuk service; integration test bila sentuh DB/MQ.
- [ ] Log tidak mengandung PII sensitif.
- [ ] Query memiliki index yang sesuai; EXPLAIN analisis untuk query berat.
- [ ] Observability (counter/histogram) ditambahkan bila relevan.
- [ ] Backward compatibility dicek (schema & API).

---

## 18) Runbooks (Operasional)

**Graceful shutdown**
- Tangani `SIGTERM`: stop HTTP → drain MQ → tutup DB/Redis.

**Troubleshooting cepat**
- 500 ramai? cek `trace_id` → korelasi log.
- Latensi naik? lihat DB wait, pool saturation, redis hits/miss.
- Rate-limited? verifikasi header `Retry-After`.
- Token invalid? sinkronisasi `JWT_SIGNING_KEY` & clock skew.

**Backup/Restore**
- Postgres: `pg_dump`/`pg_restore`; jadwal harian.
- Redis: AOF + snapshot; verifikasi RDB.
- RabbitMQ: bukan store data jangka panjang; persist queue dengan quorum + DLX.

---

## 19) Style & Konvensi

- Nama paket **singular** (`user`, `auth`, `config`).
- Interface di domain, implementasi di `repo/*`.
- Error sentinel + wrap: `fmt.Errorf("context: %w", err)`.
- Context selalu parameter pertama di service/repo.
- Time: gunakan `time.Now().UTC()`; simpan UTC di DB.

---

## 20) Contoh Endpoint Minimal

**POST /api/v1/auth/login**
```http
Request:
{
  "email": "user@example.com",
  "password": "secret"
}

Response 200:
{
  "data": {
    "access_token": "eyJ...",
    "expires_in": 900,
    "refresh_token": "def502...",
    "token_type": "Bearer",
    "user": { "id":"uuid","email":"user@example.com","name":"User" }
  },
  "trace_id": "d3b07..."
}
```

**GET /api/v1/users?page=1&page_size=20&sort=created_at:desc**
```json
{
  "data": [{ "id":"uuid", "email":"a@b.c", "name":"A", "created_at":"..." }],
  "meta": {"page":1,"page_size":20,"total":123,"sort":["created_at:desc"]},
  "trace_id":"..."
}
```

---

## 21) Roadmap Opsional

- Outbox + tx interceptor untuk publish event.
- Feature flags (env/DB) untuk canary.
- Idempotency key untuk endpoints write.
- E2E tests dengan ephemeral env (compose).
- Blue/green atau canary deploy via Traefik/Nginx.

---

**Catatan akhir**  
Gunakan dokumen ini sebagai baseline. Jika sebuah kebutuhan proyek mengharuskan penyimpangan, tulis _design note_ singkat (alasan, risiko, mitigasi) dan tautkan di PR.
