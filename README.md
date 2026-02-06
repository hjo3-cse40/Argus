Argus – Local Infrastructure (Sprint 1)
Prerequisites

Docker Desktop installed and running

Docker Compose available (docker compose)

How to Run
1-Run Docker Desktop on your computer
cd infra
docker compose up -d

Check containers:

docker compose ps

Verify RabbitMQ

URL: http://localhost:15672

Verify Postgres
docker exec -it $(docker compose ps -q db) psql -U argus -d argus

SELECT 1;
\q

Stop
docker compose down

2-Install GO

3-Run the API (from project root)
cd backend
go run ./cmd/api
Default: http://localhost:8080 (set PORT to override). Requires RabbitMQ (e.g. infra up).

Health check:
curl http://localhost:8080/health

5) Publish a Test Event (API → RabbitMQ)
curl -X POST http://localhost:8080/debug/publish


Expected response:

{ "ok": true, "event_id": "..." }


In RabbitMQ UI (raw_events queue), Ready should increase for each new event you publish.

6) Run the Worker (RabbitMQ → API)

In a new terminal:

cd backend
go run ./cmd/worker


Expected log:

worker listening on raw_events
(if you have any ready event: 2026/02/05 18:22:17 RECEIVED raw message: {"created_at":"2026-02-06T02:22:17Z","event_id":"x","source":"synthetic","title":"hello from argus","url":"https://example.com"}
2026/02/05 18:22:17 RECEIVED event_id=x
2026/02/05 18:22:17 marked delivered in API: status=200 OK
2026/02/05 18:22:17 DELIVERED + ACKED)

7) End-to-End Test (Full Pipeline)
curl -X POST http://localhost:8080/debug/publish


Expected behavior:

Worker logs RECEIVED event_id=...

Worker logs DELIVERED + ACKED

raw_events Ready count goes back to 0

8) View Delivery Status
curl http://localhost:8080/deliveries


Expected output:

[
  {
    "event_id": "...",
    "status": "delivered"
  }
]

Stop Everything
docker compose down
