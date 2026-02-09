Argus – Local Infrastructure (Sprint 1)

## Prerequisites

- Docker Desktop installed and running
- Docker Compose available (`docker compose`)
- Go 1.25+ installed

## How to Run

### 1) Start Infrastructure

Run Docker Desktop on your computer, then:

```bash
cd infra
docker compose up -d
```

Check containers:
```bash
docker compose ps
```

**Verify RabbitMQ:**
- URL: http://localhost:15672
- Login: argus / argus

**Verify Postgres:**
```bash
docker exec -it $(docker compose ps -q db) psql -U argus -d argus
SELECT 1;
\q
```

**Stop:**
```bash
docker compose down
```

### 2) Configuration

The application supports environment-based configuration (dev/stage/prod). See `docs/configuration-guide.md` for details.

**Quick start (development):**
```bash
# Defaults work for local development
# Or explicitly set:
export ENV=dev
export PORT=8080
export RABBITMQ_URL=amqp://argus:argus@localhost:5672/
```

### 3) Run the API

From project root:
```bash
cd backend
go run ./cmd/api
```

Default: http://localhost:8080 (set `PORT` env var to override). Requires RabbitMQ (e.g. infra up).

**Health check:**
```bash
curl http://localhost:8080/health
```

### 4) Run the Worker

In a new terminal:
```bash
cd backend
go run ./cmd/worker
```

Expected log:
```
Starting worker in dev environment
worker listening on raw_events
```

### 5) Publish Test Events

#### Option A: Using the API endpoint
```bash
curl -X POST http://localhost:8080/debug/publish
```

Expected response:
```json
{ "ok": true, "event_id": "..." }
```

#### Option B: Using the CLI tool (NEW)
```bash
cd backend
go run ./cmd/cli

# With custom options:
go run ./cmd/cli -source="my-source" -title="My Event" -url="https://example.com"

# Publish multiple events:
go run ./cmd/cli -count=5

# See all options:
go run ./cmd/cli -help
```

In RabbitMQ UI (raw_events queue), Ready should increase for each new event you publish.

### 6) End-to-End Test (Full Pipeline)

1. Publish an event (using API or CLI)
2. Watch worker logs - should show:
   ```
   RECEIVED raw message: {...}
   RECEIVED event_id=...
   marked delivered in API: status=200 OK
   DELIVERED + ACKED
   ```
3. Check RabbitMQ UI - `raw_events` Ready count should go back to 0

### 7) View Delivery Status
```bash
curl http://localhost:8080/deliveries
```

Expected output:
```json
[
  {
    "event_id": "...",
    "status": "delivered",
    "source": "synthetic",
    "title": "hello from argus",
    "url": "https://example.com"
  }
]
```

## Event Schema

Events follow a standardized schema defined in `backend/internal/events/schema.go`:

```go
type Event struct {
    EventID   string    `json:"event_id"`
    Source     string    `json:"source"`
    Title      string    `json:"title"`
    URL        string    `json:"url"`
    CreatedAt  time.Time `json:"created_at"`
    Metadata   map[string]interface{} `json:"metadata,omitempty"`
}
```

## Collector Team Features

- ✅ Environment configuration (dev/stage/prod)
- ✅ Formalized event schema
- ✅ CLI tool for publishing events
- ✅ Research document on data collection methods

See `docs/` for:
- `configuration-guide.md` - Detailed configuration documentation
- `data-collection-research.md` - Research on data collection methods
- `env.*.example` - Example environment files

## Stop Everything
```bash
docker compose down
```