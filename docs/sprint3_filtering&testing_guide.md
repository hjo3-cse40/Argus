# Sprint 3 - Per-Destination Filtering & Testing Guide

## Part 1: Filtering Implementation

### Overview

US 3.3: "As an admin, I want to define filters per destination (not just per source) so different channels can have different noise levels."

This feature allows keyword-based `include` and `exclude` filter rules to be attached to a platform (destination). When an event flows through the worker, it checks the platform's filters before delivering to Discord. Events that match an exclude rule are silently dropped. If include rules exist, the event must match at least one to pass through.

---

### What Was Built

#### 1. Database Schema — `destination_filters` table

A new table stores filter rules linked to a platform.

```sql
CREATE TABLE IF NOT EXISTS destination_filters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform_id UUID NOT NULL REFERENCES platforms(id) ON DELETE CASCADE,
    filter_type TEXT NOT NULL CHECK (filter_type IN ('keyword_include', 'keyword_exclude')),
    pattern TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_dest_filters_platform ON destination_filters(platform_id);
```

**File:** `backend/internal/store/migrations.go`

---

#### 2. DestinationFilter Model & MemoryStore Methods

Added the `DestinationFilter` struct and in-memory CRUD operations.

```go
type DestinationFilter struct {
    ID         string    `json:"id"`
    PlatformID string    `json:"platform_id"`
    FilterType string    `json:"filter_type"`
    Pattern    string    `json:"pattern"`
    CreatedAt  time.Time `json:"created_at"`
}
```

Methods added to `MemoryStore`:
- `AddFilter(filter DestinationFilter) error`
- `ListFilters(platformID string) []DestinationFilter`
- `DeleteFilter(id string) error`
- Cascade delete of filters when a platform is deleted

**File:** `backend/internal/store/store.go`

---

#### 3. Store Interface

Added filter methods to the `Store` interface so both `MemoryStore` and `PostgresStore` implement them.

```go
AddFilter(filter DestinationFilter) error
ListFilters(platformID string) []DestinationFilter
DeleteFilter(id string) error
```

**File:** `backend/internal/store/interface.go`

---

#### 4. PostgreSQL Implementation

Implemented `AddFilter`, `ListFilters`, and `DeleteFilter` using SQL queries against the `destination_filters` table.

**File:** `backend/internal/store/postgres.go`

---

#### 5. Filter Evaluation Engine

Core filtering logic in a dedicated package. The `Evaluate` function decides whether an event should be delivered.

**Rules:**
- No filters → event passes (backwards compatible)
- `keyword_exclude` → event is dropped if title or description matches ANY exclude pattern
- `keyword_include` → event must match at least ONE include pattern to pass
- Excludes are evaluated first and take priority over includes
- All matching is **case-insensitive**

```go
func Evaluate(ev *events.Event, filters []store.DestinationFilter) bool
```

**Files:**
- `backend/internal/filter/filter.go` — evaluation logic
- `backend/internal/filter/filter_test.go` — 12 unit tests

**Unit tests cover:**
- No filters (passes all)
- Exclude blocks matching events
- Exclude passes non-matching events
- Case-insensitive matching
- Include passes matching events
- Include blocks non-matching events
- Multiple includes (OR logic)
- Exclude takes priority over include
- Matching against description metadata
- Empty description handling

---

#### 6. Worker Integration

The worker was updated to:
1. Resolve the destination using `ResolveDestination()` which returns both the webhook URL and the `platform_id`
2. If a `platform_id` is available, load filters via `ListFilters(platformID)`
3. Run `filter.Evaluate(event, filters)` before sending
4. ACK and skip events that don't pass filters (logged as `FILTERED`)
5. Fall back to the global webhook if destination resolution fails

**File:** `backend/cmd/worker/main.go`

---

#### 7. Filter CRUD API

HTTP handlers for managing filters through the API.

**Endpoints:**
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/platforms/{platform_id}/filters` | Create a filter |
| `GET` | `/api/platforms/{platform_id}/filters` | List filters for a platform |
| `DELETE` | `/api/filters/{id}` | Delete a filter |

**Filter types:** `keyword_include`, `keyword_exclude`

**Files:**
- `backend/internal/http/handlers/filters.go` — handler logic
- `backend/internal/http/router.go` — route registration

---

### Files Summary

| File | Change |
|------|--------|
| `backend/internal/store/migrations.go` | Added `destination_filters` table |
| `backend/internal/store/store.go` | Added `DestinationFilter` struct + MemoryStore methods |
| `backend/internal/store/interface.go` | Added filter methods to Store interface |
| `backend/internal/store/postgres.go` | PostgreSQL filter CRUD implementation |
| `backend/internal/filter/filter.go` | Filter evaluation engine (NEW) |
| `backend/internal/filter/filter_test.go` | 12 unit tests for filter logic (NEW) |
| `backend/internal/http/handlers/filters.go` | Filter API handlers (NEW) |
| `backend/internal/http/router.go` | Added filter routes |
| `backend/cmd/worker/main.go` | Integrated filtering into event processing |
| `backend/internal/notifier/discord.go` | Added `ResolveDestination` for platform ID lookup |

---

## Part 2: Testing Guide

### TESTING FILTERING

1. Start docker (own terminal) 

cd …/argus/infra && docker compose up -d


2. Health check for Docker 

cd …/argus/infra && docker compose ps 



3. Start API server (own terminal), verify that api is listening in terminal 

cd …/argus/backend && go run ./cmd/api



4. Start worker (own terminal), verify that worker is listening in terminal 

cd …/argus/backend && go run ./cmd/worker



5. Create Youtube Platform, will produce ID which we will use it as PLATFORM_ID
	(Platform name is CASE SENSITIVE: Youtube and youtube won't match) 

curl -s -X POST http://localhost:8080/api/platforms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "youtube",
    "discord_webhook": "https://discord.com/api/webhooks/1473840112663134273/M_MVvdVoLShRjcuKGyLk291SnZHwGLlqxZzRaYE_3ED3sxF4fM_fL0d9Xvm9o2-sN-y2"
  }' | python3 -m json.tool

OUTPUTS <PLATFORM_ID>

- To show platforms in database : 

curl -s http://localhost:8080/api/platforms | python3 -m json.tool

(Right now, since our docker is local, we need to create our own platforms on their machines until we connect to proxmox 




6. Create sub_source_ID AKA the actual channel (NBA, MLB, etc) , input name, ID, and url 
	will produce another id, which we will use as SUBSOURCE_ID

curl -s -X POST http://localhost:8080/api/platforms/<PLATFORM_ID>/subsources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "NBA",
    "identifier": "UCWJ2lWNubArHWmf3FIHbfcQ",
    "url": "https://www.youtube.com/channel/UCWJ2lWNubArHWmf3FIHbfcQ"
  }' | python3 -m json.tool

SUBSOURCE_ID: (…) 



7. Add exclude filter , replace FILTER with actual filter word 

curl -s -X POST http://localhost:8080/api/platforms/<PLATFORM_ID>/filters \
  -H "Content-Type: application/json" \
  -d '{
    "filter_type": "keyword_exclude",
    "pattern": "<FILTER>"
  }'

(Filter is case-insensitive so Shorts, SHORTS, sHorTS, etc will all be filtered the same)



8. Verify setup (platform and filter) 

curl -s http://localhost:8080/api/platforms | python3 -m json.tool
curl -s http://localhost:8080/api/platforms/<PLATFORM_ID>/filters | python3 -m json.tool



9. Send test event with blocked filter word, should NOT pass (exclude word: Shorts)
	video name SHOULD include filtered word. Check worker terminal for result


curl -s -X POST http://localhost:8080/api/ingest \
  -H "Content-Type: application/json" \
  -d '{
    "source": "youtube",
    "title": "NBA Shorts - Best Dunks of the Week",
    "url": "https://youtube.com/shorts/test123",
    "metadata": {
      "subsource_id": "<SUBSOURCE_ID>"
    }
  }' | python3 -m json.tool


Worker terminal should show: FILTERED event_id=... (did not pass destination filters). No message in Discord.


10. Send test event that WILL pass, title should NOT contain filter word 

curl -s -X POST http://localhost:8080/api/ingest \
  -H "Content-Type: application/json" \
  -d '{
    "source": "youtube",
    "title": "NBA Full Game Highlights Lakers vs Celtics",
    "url": "https://youtube.com/watch?v=allowed",
    "metadata": {
      "subsource_id": "<SUBSOURCE_ID>"
    }
  }' | python3 -m json.tool


Worker terminal should show: discord delivered event_id=... and DELIVERED + ACKED. Message should appear in Discord.


11. Delete filter (if desired)

   curl -s -X DELETE http://localhost:8080/api/filters/<FILTER_ID>


12. Shut down system

"Ctrl + C" in worker terminal and api terminal 
cd .../argus/infra && docker compose down
