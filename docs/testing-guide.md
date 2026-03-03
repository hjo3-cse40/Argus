# Argus End-to-End Testing Guide

## Full Pipeline: NBA YouTube -> RabbitMQ -> Worker -> Discord

### Prerequisites

- Go installed
- Docker and Docker Compose installed

---

## Step 1: Start Infrastructure (Postgres, RabbitMQ, RSSHub)

```bash
cd /Users/samjo/code/argus/infra
docker compose up -d
```

Wait ~15 seconds for all services to become healthy. Verify:

```bash
docker compose ps
```

All three services (`db`, `rabbitmq`, `rsshub`) should show `healthy`.

---

## Step 2: Start the API Server

Open a **new terminal**:

```bash
cd /Users/samjo/code/argus/backend
go run ./cmd/api
```

Expected output:

```
Starting API in dev environment
Connected to PostgreSQL database at localhost:5432
API listening on http://localhost:8080
```

Leave this running.

---

## Step 3: Seed the Database with a Platform and Subsource

The RSS collector reads from the database, not the `.env` file. You need to create the YouTube platform and NBA subsource via the API.

Open a **new terminal** for curl commands.

### 3a. Create the YouTube platform

```bash
curl -s -X POST http://localhost:8080/api/platforms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "youtube",
    "discord_webhook": "https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN"
  }' | python3 -m json.tool
```

> **Note:** Replace the discord_webhook value with your actual Discord webhook URL from your `.env` file.

This returns a response with the platform `id`. **Copy that `id` value** for the next step.

### 3b. Create the NBA YouTube subsource

Replace `PLATFORM_ID` with the id from Step 3a:

```bash
curl -s -X POST http://localhost:8080/api/platforms/PLATFORM_ID/subsources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "NBA",
    "identifier": "UCWJ2lWNubArHWmf3FIHbfcQ"
  }' | python3 -m json.tool
```

`UCWJ2lWNubArHWmf3FIHbfcQ` is the NBA YouTube channel ID.

### 3c. Verify everything is saved

```bash
curl -s http://localhost:8080/api/platforms | python3 -m json.tool
```

You should see the youtube platform with its ID and webhook.

---

## Step 4: Start the Worker

Open a **new terminal**:

```bash
cd /Users/samjo/code/argus/backend
go run ./cmd/worker
```

Expected output:

```
Starting worker in dev environment
[OK] Discord webhook URL loaded
worker listening on raw_events
```

Leave this running. It consumes from the `raw_events` queue and sends to Discord.

---

## Step 5: Start the RSS Collector

Open a **new terminal**:

```bash
cd /Users/samjo/code/argus/backend
go run ./cmd/rss
```

Expected output:

```
Loaded subsource: youtube - NBA (identifier: UCWJ2lWNubArHWmf3FIHbfcQ)
Polling 1 subsource(s)
[youtube - NBA] Fetching: http://localhost:1200/youtube/channel/UCWJ2lWNubArHWmf3FIHbfcQ
[youtube - NBA] NBA — 5 items
[OK] [youtube - NBA] NBA -- <video title 1>
[OK] [youtube - NBA] NBA -- <video title 2>
...
```

---

## What Happens Behind the Scenes

1. **RSS collector** hits RSSHub at `http://localhost:1200/youtube/channel/UCWJ2lWNubArHWmf3FIHbfcQ`
2. **RSSHub** fetches the NBA YouTube channel feed, returns XML with recent videos
3. **RSS collector** parses the XML, creates Event objects with:
   - `Title` = video title
   - `URL` = YouTube video link
   - `Metadata` = `subsource_id`, `platform_name: "youtube"`, `subsource_name: "NBA"`, `description`, `author`, etc.
4. Events are published to the **`raw_events` RabbitMQ queue**
5. **Worker** consumes each event from the queue
6. Worker calls `notifier.SendDiscordWebhook()` with the global `DISCORD_WEBHOOK_URL`
7. **Discord** receives an embed with the video title, link, source info, and timestamp
8. Worker marks the delivery as `delivered` via `POST /debug/delivered`
9. Message is ACKed in RabbitMQ

---

## Step 6: Verify

### Check Discord
Look at the Discord channel associated with your webhook. You should see Argus embeds with NBA YouTube video titles.

### Check Delivery History via API

```bash
curl -s http://localhost:8080/deliveries | python3 -m json.tool
```

### Check RabbitMQ Management UI
Visit `http://localhost:15672` (login: `argus` / `argus`) to see queue statistics.

### Check Argus Web UI
Visit `http://localhost:8080` in a browser to see the Argus dashboard and notification history.

---

## Terminal Summary

| Terminal | Command                          | Purpose                          |
|----------|----------------------------------|----------------------------------|
| 1        | `docker compose up -d` (infra/)  | Postgres, RabbitMQ, RSSHub       |
| 2        | `go run ./cmd/api` (backend/)    | HTTP API server                  |
| 3        | One-off `curl` commands          | Seed DB with platform/subsource  |
| 4        | `go run ./cmd/worker` (backend/) | Queue consumer, sends to Discord |
| 5        | `go run ./cmd/rss` (backend/)    | RSS poller, publishes events     |

---

## Shutting Down

1. `Ctrl+C` in terminals 2, 4, and 5 (API, Worker, RSS)
2. Stop infrastructure:

```bash
cd /Users/samjo/code/argus/infra
docker compose down
```

To also remove stored data (clean slate):

```bash
docker compose down -v
```
